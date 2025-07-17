package fts

import (
	"context"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
)

func InitFTSListeners(
	baseDir string,
	builtInLister BuiltInLister,
) (*ftsengine.Engine, error) {
	cfg := ftsengine.Config{
		BaseDir:    baseDir,
		DBFileName: spec.SqliteDBFileName,
		Table:      sqliteDBTableName,
		Columns:    ftsColumns,
	}
	ftsE, err := ftsengine.NewEngine(cfg)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	StartUserPromptsFTSRebuild(ctx, baseDir, ftsE)

	if builtInLister != nil {
		StartBuiltInPromptsFTSRebuild(
			ctx,
			builtInLister,
			ftsE,
		)
	}
	return ftsE, nil
}
