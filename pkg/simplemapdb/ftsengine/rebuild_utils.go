package ftsengine

import (
	"context"
	"io/fs"
	"log/slog"
	"path/filepath"
	"time"
)

// SyncDecision tells the helper what to do with ONE file.
type SyncDecision struct {
	// The consumer id of the doc (i.e externalID), must be non-empty if Skip == false.
	ID string
	// New value for compareColumn (ignored if Unchanged).
	CmpOut string
	// FTS column map              (ignored if Unchanged).
	Vals map[string]string
	// Index already up-to-date, no processing needed.
	Unchanged bool
	// Ignore this file completely, delete if present.
	Skip bool
}

// GetPrevCmp lets the callback query the value that is *currently*
// stored in compareColumn for a specific externalID ("" = not indexed).
type GetPrevCmp func(id string) string

// ProcessFile is the single user callback.
// It may call getPrev(id) zero or more times.
// If this func returns a err the walk will be stopped.
type ProcessFile func(
	ctx context.Context,
	baseDir, fullPath string,
	getPrev GetPrevCmp,
) (SyncDecision, error)

// SyncDirToFTS walks “baseDir” and brings the FTS table in sync.
//
// The compareColumn must exist in Engine.Config.Columns (mtime, hash, etc).
func SyncDirToFTS(
	ctx context.Context,
	engine *Engine,
	baseDir string,
	compareColumn string,
	batchSize int,
	processFile ProcessFile,
) error {
	if batchSize <= 0 {
		batchSize = 1000
	}
	const batchListPageSize = 10000
	start := time.Now()

	slog.Info("fts-sync start", "dir", baseDir, "cmpCol", compareColumn)

	// Read current index to get a map of externalID to compareValue.
	existing := make(map[string]string)

	token := ""
	for {
		part, next, err := engine.BatchList(
			ctx,
			compareColumn,
			[]string{compareColumn},
			token,
			batchListPageSize,
		)
		if err != nil {
			return err
		}
		for _, row := range part {
			existing[row.ID] = row.Values[compareColumn]
		}
		if next == "" {
			break
		}
		token = next
	}
	getPrev := func(id string) string { return existing[id] }
	var (
		// Files whose ID was accepted (Skip==false).
		nProcessed int
		// Skip==true or empty ID.
		nSkipped int
		// Dec.Unchanged==true.
		nUnchanged int
		// Rows really written to the index.
		nUpserted int
	)
	// Walk directory – incremental updates in small batches.
	seenNow := make(map[string]struct{}, 4096)
	pending := make(map[string]map[string]string, batchSize)

	flush := func() error {
		if len(pending) == 0 {
			return nil
		}
		if err := engine.BatchUpsert(ctx, pending); err != nil {
			return err
		}
		nUpserted += len(pending)
		pending = make(map[string]map[string]string, batchSize)
		return nil
	}

	err := filepath.WalkDir(baseDir, func(p string, d fs.DirEntry, we error) error {
		if d.IsDir() || we != nil {
			// Its a dir, we dont want to process it other than walking.
			// There was some walk error in this particular path.
			return we
		}

		dec, err := processFile(ctx, baseDir, p, getPrev)
		if err != nil {
			// Consumer returned a error, they want us to stop the walk.
			return err
		}
		if dec.Skip || dec.ID == "" {
			// Silently skip on Skip==true for a path in consumer.
			// Or No valid id -> treat like Skip.
			nSkipped++
			return nil
		}

		seenNow[dec.ID] = struct{}{}
		nProcessed++

		if dec.Unchanged {
			nUnchanged++
			return nil
		}

		vals := dec.Vals
		if vals == nil {
			vals = map[string]string{}
		}
		vals[compareColumn] = dec.CmpOut
		pending[dec.ID] = vals

		if len(pending) >= batchSize {
			return flush()
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := flush(); err != nil {
		return err
	}

	// Delete documents whose source file vanished.
	var toDelete []string
	for id := range existing {
		if _, ok := seenNow[id]; !ok {
			toDelete = append(toDelete, id)
		}
	}
	if err := engine.BatchDelete(ctx, toDelete); err != nil {
		return err
	}

	slog.Info("fts-sync done",
		"dir", baseDir,
		"took", time.Since(start),
		"processed", nProcessed,
		"upserted", nUpserted,
		"unchanged", nUnchanged,
		"skipped", nSkipped,
		"deleted", len(toDelete),
	)
	return nil
}
