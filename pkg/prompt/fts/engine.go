package fts

import (
	"context"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
)

func InitFTSListeners(
	ctx context.Context,
	baseDir string,
	builtInLister BuiltInLister,
) (*ftsengine.Engine, error) {
	ftsE, err := ftsengine.NewEngine(ftsengine.Config{
		BaseDir:    baseDir,
		DBFileName: spec.SqliteDBFileName,
		Table:      sqliteDBTableName,
		Columns:    ftsColumns,
	})
	if err != nil {
		return nil, err
	}
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
