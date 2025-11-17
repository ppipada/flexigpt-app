package fts

import (
	"context"

	"github.com/ppipada/flexigpt-app/pkg/tool/spec"
	"github.com/ppipada/mapstore-go/ftsengine"
)

// InitToolFTSListeners initialises the tools FTS database and starts background rebuild jobs.
func InitToolFTSListeners(
	baseDir string,
	builtInLister ToolBuiltInLister,
) (*ftsengine.Engine, error) {
	cfg := ftsengine.Config{
		BaseDir:    baseDir,
		DBFileName: spec.ToolDBFileName,
		Table:      sqliteDBTableName,
		Columns:    ftsColumns,
	}
	ftsE, err := ftsengine.NewEngine(cfg)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	StartUserToolsFTSRebuild(ctx, baseDir, ftsE)

	if builtInLister != nil {
		StartBuiltInToolsFTSRebuild(ctx, builtInLister, ftsE)
	}
	return ftsE, nil
}
