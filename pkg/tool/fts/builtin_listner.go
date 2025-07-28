package fts

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"

	"github.com/ppipada/flexigpt-app/pkg/tool/spec"
)

// ToolBuiltInLister returns the current snapshot of all built-in bundles and tools.
// Overlay should already be applied, so that enabled flags are correct.
type ToolBuiltInLister func() (
	bundles map[bundleitemutils.BundleID]spec.ToolBundle,
	tools map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool,
	err error,
)

// StartBuiltInToolsFTSRebuild spawns a goroutine that synchronises the built-in tools with the FTS index.
func StartBuiltInToolsFTSRebuild(
	ctx context.Context,
	lister ToolBuiltInLister,
	engine *ftsengine.Engine,
) {
	if lister == nil || engine == nil {
		return
	}
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("builtin-tools-fts panic",
					"err", rec, "stack", string(debug.Stack()))
			}
		}()
		if err := syncBuiltInsToFTS(ctx, lister, engine); err != nil {
			slog.Error("builtin-tools-fts initial sync failed", "err", err)
		}
	}()
}

// ReindexOneBuiltInTool updates exactly one built-in tool.
func ReindexOneBuiltInTool(
	ctx context.Context,
	bundleID bundleitemutils.BundleID,
	bundleSlug bundleitemutils.BundleSlug,
	tool spec.Tool,
	engine *ftsengine.Engine,
) error {
	if engine == nil {
		return spec.ErrInvalidRequest
	}
	docID, vals, ok := buildDoc(bundleID, bundleSlug, tool)
	if !ok {
		return fmt.Errorf(
			"builtin-tools-fts cannot create doc for bundleID %q, bundleSlug %q, toolID %q",
			bundleID, bundleSlug, tool.ID,
		)
	}
	if err := engine.Upsert(ctx, docID, vals); err != nil {
		return fmt.Errorf("builtin-tools-fts upsert(one) failed for %s: %w", docID, err)
	}
	return nil
}

// syncBuiltInsToFTS performs a full synchronisation of the built-in data-set with the FTS table.
func syncBuiltInsToFTS(
	ctx context.Context,
	lister ToolBuiltInLister,
	engine *ftsengine.Engine,
) error {
	iter := func(
		getPrev ftsengine.GetPrevCmp,
		emit func(ftsengine.SyncDecision) error,
	) error {
		bundles, tools, err := lister()
		if err != nil {
			return err
		}
		for bid, b := range bundles {
			for _, tl := range tools[bid] {
				id, vals, ok := buildDoc(bid, b.Slug, tl)
				if !ok {
					continue
				}
				cmp := vals[compareColumn]
				dec := ftsengine.SyncDecision{
					ID:     id,
					CmpOut: cmp,
					Vals:   vals,
				}
				if getPrev(id) == cmp {
					dec.Unchanged = true
				}
				if err := emit(dec); err != nil {
					return err
				}
			}
		}
		return nil
	}

	belongs := func(id string) bool { return strings.HasPrefix(id, BuiltInDocPrefix) }

	return ftsengine.SyncIterToFTS(
		ctx, engine, compareColumn, upsertBatchSize, iter, belongs,
	)
}

// buildDoc converts one tool to (docID, columnMap).
func buildDoc(
	bid bundleitemutils.BundleID,
	bslug bundleitemutils.BundleSlug,
	tl spec.Tool,
) (docID string, vals map[string]string, ok bool) {
	dirInfo, err := bundleitemutils.BuildBundleDir(bid, bslug)
	if err != nil {
		slog.Error("builtin-tools-fts: BuildBundleDir failed",
			"bundleID", bid, "err", err)
		return docID, vals, ok
	}
	fileInfo, err := bundleitemutils.BuildItemFileInfo(tl.Slug, tl.Version)
	if err != nil {
		slog.Error("builtin-tools-fts: BuildItemFileInfo failed",
			"bundleID", bid, "slug", tl.Slug, "ver", tl.Version, "err", err)
		return docID, vals, ok
	}

	docID = path.Join(BuiltInDocPrefix+dirInfo.DirName, fileInfo.FileName)
	doc := toolToFTSDoc(bid, tl)
	return docID, doc.ToMap(), true
}

// toolToFTSDoc fills an ftsDoc from a ToolSpec.
func toolToFTSDoc(bid bundleitemutils.BundleID, tl spec.Tool) ftsDoc {
	doc := ftsDoc{
		Slug:        tl.Slug,
		DisplayName: tl.DisplayName,
		Desc:        tl.Description,
		Tags:        strings.Join(tl.Tags, newline) + newline,
		Args:        extractArgsFromRaw(tl.ArgSchema),
		Impl:        string(tl.Type),
		MTime:       tl.ModifiedAt.UTC().Format(time.RFC3339Nano),
		BundleID:    bid,
		Enabled:     enabledFalse,
	}

	if tl.IsEnabled {
		doc.Enabled = enabledTrue
	}

	switch tl.Type {
	case spec.ToolTypeGo:
		if tl.GoImpl != nil {
			doc.ImplMeta = tl.GoImpl.Func
		}
	case spec.ToolTypeHTTP:
		if tl.HTTP != nil {
			doc.ImplMeta = tl.HTTP.Request.Method + " " + tl.HTTP.Request.URLTemplate
		}
	}
	return doc
}
