package fts

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
)

// BuiltInLister returns the current snapshot of all built-in bundles and templates.
// Overlay should already be applied, so that enabled flags are correct.
type BuiltInLister func() (
	bundles map[bundleitemutils.BundleID]spec.PromptBundle,
	templates map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate,
	err error,
)

// StartBuiltInPromptsFTSRebuild spawns a goroutine that synchronises the built-in templates with the FTS index.
// Safe to call multiple times; the work is executed once.
func StartBuiltInPromptsFTSRebuild(
	ctx context.Context,
	lister BuiltInLister,
	engine *ftsengine.Engine,
) {
	if lister == nil || engine == nil {
		return
	}

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("builtin-fts panic",
					"err", rec, "stack", string(debug.Stack()))
			}
		}()
		if err := syncBuiltInsToFTS(ctx, lister, engine); err != nil {
			slog.Error("builtin-fts initial sync failed", "err", err)
		}
	}()
}

// ReindexOneBuiltIn updates exactly one built-in template (used by enable / disable mutators).
func ReindexOneBuiltIn(
	ctx context.Context,
	bundleID bundleitemutils.BundleID,
	bundleSlug bundleitemutils.BundleSlug,
	template spec.PromptTemplate,
	engine *ftsengine.Engine,
) error {
	if engine == nil {
		return spec.ErrInvalidRequest
	}

	docID, vals, ok := buildDoc(bundleID, bundleSlug, template)
	if !ok {
		return fmt.Errorf(
			"builtin-fts cannot create doc for bundleID %q, bundleSlug %q, templateID %q",
			bundleID, bundleSlug, template.ID,
		)
	}
	if err := engine.Upsert(ctx, docID, vals); err != nil {
		return fmt.Errorf("builtin-fts upsert(one) failed for %s: %w", docID, err)
	}
	return nil
}

// syncBuiltInsToFTS performs a full synchronisation of the built-in data-set with the FTS table.
func syncBuiltInsToFTS(
	ctx context.Context,
	lister BuiltInLister,
	engine *ftsengine.Engine,
) error {
	// Iterator that emits SyncDecision events.
	iter := func(
		getPrev ftsengine.GetPrevCmp,
		emit func(ftsengine.SyncDecision) error,
	) error {
		bundles, tpls, err := lister()
		if err != nil {
			return err
		}
		for bid, b := range bundles {
			for _, tpl := range tpls[bid] {
				id, vals, ok := buildDoc(bid, b.Slug, tpl)
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

	// Predicate that tells the helper which rows belong to us.
	belongs := func(id string) bool {
		return strings.HasPrefix(id, BuiltInDocPrefix)
	}

	// Delegate to the generic helper.
	return ftsengine.SyncIterToFTS(
		ctx,
		engine,
		compareColumn,
		upsertBatchSize,
		iter,
		belongs,
	)
}

// buildDoc converts one template to (docID, columnMap).
// If anything is structurally wrong - should never happen with validated built-ins - it logs and returns ok == false.
func buildDoc(
	bid bundleitemutils.BundleID,
	bslug bundleitemutils.BundleSlug,
	tpl spec.PromptTemplate,
) (docID string, vals map[string]string, ok bool) {
	dirInfo, err := bundleitemutils.BuildBundleDir(bid, bslug)
	if err != nil {
		slog.Error("builtin-fts: BuildBundleDir failed",
			"bundleID", bid, "err", err)
		return docID, vals, ok
	}
	fileInfo, err := bundleitemutils.BuildItemFileInfo(tpl.Slug, tpl.Version)
	if err != nil {
		slog.Error("builtin-fts: BuildItemFileInfo failed",
			"bundleID", bid, "slug", tpl.Slug, "ver", tpl.Version, "err", err)
		return docID, vals, ok
	}

	docID = path.Join(BuiltInDocPrefix+dirInfo.DirName, fileInfo.FileName)
	doc := templateToFTSDoc(bid, tpl)
	return docID, doc.ToMap(), true
}

// templateToFTSDoc fills an ftsDoc from a PromptTemplate.
func templateToFTSDoc(bid bundleitemutils.BundleID, tpl spec.PromptTemplate) ftsDoc {
	doc := ftsDoc{
		Slug:        tpl.Slug,
		DisplayName: tpl.DisplayName,
		Desc:        tpl.Description,
		BundleID:    bid,
		MTime:       tpl.ModifiedAt.UTC().Format(time.RFC3339Nano),
		Enabled:     enabledFalse,
	}
	if tpl.IsEnabled {
		doc.Enabled = enabledTrue
	}
	for _, mb := range tpl.Blocks {
		doc.Messages += mb.Content + newline
	}
	for _, tg := range tpl.Tags {
		doc.Tags += tg + newline
	}
	return doc
}
