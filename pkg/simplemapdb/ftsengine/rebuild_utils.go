package ftsengine

import (
	"context"
	"io/fs"
	"log/slog"
	"path/filepath"
)

type GetUpsertData func(ctx context.Context, baseDir, fileFullPath string) (id string, vals map[string]string, skip bool)

func DirWalkAndRebuild(
	ctx context.Context,
	baseDir string,
	e *Engine,
	getUpsertData GetUpsertData,
) error {
	slog.Info("FTSEngine dirwalk rebuild start", "BaseDir", baseDir)
	err := filepath.WalkDir(baseDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			// Skip dir or err.
			return err
		}

		id, vals, skip := getUpsertData(ctx, baseDir, p)
		if skip {
			return nil
		}
		return e.Upsert(ctx, id, vals)
	})
	if err != nil {
		return err
	}
	slog.Info("FTSEngine dirwalk rebuild done", "BaseDir", baseDir)
	return nil
}

func DirWalkAndRebuildIfEmpty(
	ctx context.Context,
	baseDir string,
	e *Engine,
	getUpsertData GetUpsertData,
) error {
	isEmp, err := e.IsEmpty(ctx)
	if err != nil {
		slog.Error(
			"DB empty check issue. possibly delete the .sqlite file and restart app again",
			"basedir",
			baseDir,
			"error",
			err,
		)
		return err
	}
	if isEmp {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("FTSEngine rebuild panic", "error", r)
				}
			}()
			if err := DirWalkAndRebuild(ctx, baseDir, e, getUpsertData); err != nil {
				slog.Error("FTSEngine rebuild failed", "error", err)
			}
		}()
	}
	return nil
}
