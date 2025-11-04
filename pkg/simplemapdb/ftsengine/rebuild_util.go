package ftsengine

import (
	"context"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"time"
)

type SyncDecision struct {
	// External identifier (rowid in the virtual table).
	// Must be non-empty f Skip == false.
	ID string
	// Value for compareColumn (mtime / hash / ...). Ignored when Unchanged.
	CmpOut string
	// Column -> text map for FTS. Ignored when Unchanged.
	Vals map[string]string
	// The row is already up-to-date, nothing to do.
	Unchanged bool
	// Ignore this document entirely (also triggers delete if it existed).
	Skip bool
}

// GetPrevCmp allows producers to query the compareColumn value that is
// currently stored for a specific ID ("" == not indexed yet).
type GetPrevCmp func(id string) string

// ProcessFile is the directory-walker callback.
type ProcessFile func(
	ctx context.Context,
	baseDir, fullPath string,
	getPrev GetPrevCmp,
) (SyncDecision, error)

func SyncDirToFTS(
	ctx context.Context,
	engine *Engine,
	baseDir string,
	compareColumn string,
	batchSize int,
	processFile ProcessFile,
) error {
	// Factory that converts the WalkDir stream into SyncDecision events.
	iter := func(getPrev GetPrevCmp, emit func(SyncDecision) error) error {
		return filepath.WalkDir(baseDir,
			func(p string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil || d.IsDir() {
					return walkErr
				}
				dec, err := processFile(ctx, baseDir, p, getPrev)
				if err != nil {
					return err
				}
				return emit(dec)
			})
	}

	// A row belongs to this dataset when its ID starts with baseDir.
	belongs := func(id string) bool { return strings.HasPrefix(id, baseDir) }

	return SyncIterToFTS(
		ctx,
		engine,
		compareColumn,
		batchSize,
		iter,
		belongs,
	)
}

// Iterate is the generic producer contract.
// GetPrev      lets the producer look at the current compareColumn value.
// Emit(dec)    must be invoked exactly once for every document that belongs to this dataset.
type Iterate func(getPrev GetPrevCmp, emit func(SyncDecision) error) error

// SyncIterToFTS. Belongs(id) must return true for all rows owned by this producer so that vanished rows can be deleted.
func SyncIterToFTS(
	ctx context.Context,
	engine *Engine,
	compareColumn string,
	batchSize int,
	iter Iterate,
	belongs func(id string) bool,
) error {
	if batchSize <= 0 {
		batchSize = 1000
	}
	const listPage = 10_000
	start := time.Now()

	slog.Info("fts-sync start", "cmpCol", compareColumn)

	// Fetch current state (ID -> compareColumn value).
	existing := make(map[string]string)

	token := ""
	for {
		part, next, err := engine.BatchList(
			ctx,
			compareColumn,
			[]string{compareColumn},
			token,
			listPage,
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

	// Incremental diff while the producer iterates over its dataset.
	var (
		nProcessed, nSkipped, nUnchanged, nUpserted int
	)

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

	emit := func(dec SyncDecision) error {
		if dec.Skip || dec.ID == "" {
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
	}

	if err := iter(getPrev, emit); err != nil {
		return err
	}
	if err := flush(); err != nil {
		return err
	}

	// Delete documents that vanished from the producers dataset.
	var toDelete []string
	for id := range existing {
		if !belongs(id) { // ignore rows owned by other producers
			continue
		}
		if _, ok := seenNow[id]; !ok {
			toDelete = append(toDelete, id)
		}
	}
	if len(toDelete) != 0 {
		if err := engine.BatchDelete(ctx, toDelete); err != nil {
			return err
		}
	}

	// Done - statistics.
	slog.Info("fts-sync done",
		"took", time.Since(start),
		"processed", nProcessed,
		"upserted", nUpserted,
		"unchanged", nUnchanged,
		"skipped", nSkipped,
		"deleted", len(toDelete),
	)
	return nil
}
