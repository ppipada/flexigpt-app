package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"          // POSIX paths for embed.FS
	"path/filepath" // OS paths for the real FS
	"sync"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/booloverlay"
	"github.com/ppipada/flexigpt-app/pkg/builtin"
	"github.com/ppipada/flexigpt-app/pkg/prompt/nameutils"
	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
)

const overlayJSON = "overlay.json"

type BuiltInBundleID spec.BundleID

func (BuiltInBundleID) Group() string { return "bundles" }
func (b BuiltInBundleID) ID() string  { return string(b) }

type BuiltInTemplateID spec.TemplateID

func (BuiltInTemplateID) Group() string { return "templates" }
func (t BuiltInTemplateID) ID() string  { return string(t) }

// BuiltInData keeps the built-in prompt assets and an overlay with enable flags.
// A cached snapshot already has the overlay applied.
type BuiltInData struct {
	overlayBaseDir string
	bundles        map[spec.BundleID]spec.PromptBundle
	templates      map[spec.BundleID]map[spec.TemplateID]spec.PromptTemplate
	store          *booloverlay.Store

	mu            sync.RWMutex
	viewBundles   map[spec.BundleID]spec.PromptBundle
	viewTemplates map[spec.BundleID]map[spec.TemplateID]spec.PromptTemplate

	rebuilder *builtin.AsyncRebuilder
}

// SnapshotMaxAge controls after how much time a background rebuild is triggered.
func NewBuiltInData(overlayBaseDir string, snapshotMaxAge time.Duration) (*BuiltInData, error) {
	if snapshotMaxAge <= 0 {
		snapshotMaxAge = time.Hour
	}

	if err := os.MkdirAll(overlayBaseDir, 0o755); err != nil {
		return nil, err
	}
	store, err := booloverlay.NewStore(
		filepath.Join(overlayBaseDir, overlayJSON),
		booloverlay.WithKeyType[BuiltInBundleID](),
		booloverlay.WithKeyType[BuiltInTemplateID](),
	)
	if err != nil {
		return nil, err
	}

	bundlesFS, err := fs.Sub(
		builtin.BuiltinPromptBundlesFS,
		builtin.BuiltinPromptBundlesRootDir,
	)
	if err != nil {
		return nil, err
	}
	// Load bundles.json.
	rawBundles, err := fs.ReadFile(bundlesFS, builtin.BuiltinPromptBundlesJSON)
	if err != nil {
		return nil, err
	}
	var manifest spec.AllBundles
	if err := json.Unmarshal(rawBundles, &manifest); err != nil {
		return nil, err
	}

	bundleMap := make(map[spec.BundleID]spec.PromptBundle, len(manifest.Bundles))
	templateMap := make(map[spec.BundleID]map[spec.TemplateID]spec.PromptTemplate)
	for id, b := range manifest.Bundles {
		b.IsBuiltIn = true
		bundleMap[id] = b
		templateMap[id] = make(map[spec.TemplateID]spec.PromptTemplate)
	}

	// Walk & validate templates.
	seenTpl := make(map[spec.TemplateID]string)
	err = fs.WalkDir(
		bundlesFS,
		".",
		func(inPath string, d fs.DirEntry, _ error) error {
			if d.IsDir() || path.Ext(inPath) != ".json" {
				return nil
			}
			fn := path.Base(inPath)
			if fn == builtin.BuiltinPromptBundlesJSON ||
				fn == overlayJSON {
				return nil
			}

			// Validate bundle dir.
			dir := path.Base(path.Dir(inPath))
			dirInfo, derr := nameutils.ParseBundleDir(dir)
			if derr != nil {
				return fmt.Errorf("%s: %w", inPath, derr)
			}
			bundleID := dirInfo.ID

			bDef, ok := bundleMap[bundleID]
			if !ok {
				return fmt.Errorf("%s: bundle dir %q not in bundles.json", inPath, bundleID)
			}
			if dirInfo.Slug != bDef.Slug {
				return fmt.Errorf("%s: dir slug %q ≠ manifest slug %q",
					inPath, dirInfo.Slug, bDef.Slug)
			}

			// Load template JSON.
			raw, err := fs.ReadFile(bundlesFS, inPath)
			if err != nil {
				return err
			}
			var tpl spec.PromptTemplate
			if err := json.Unmarshal(raw, &tpl); err != nil {
				return fmt.Errorf("%s: %w", inPath, err)
			}
			tpl.IsBuiltIn = true

			if err := nameutils.ValidateTemplateSlug(tpl.Slug); err != nil {
				return fmt.Errorf("%s: %w", inPath, err)
			}
			if err := nameutils.ValidateTemplateVersion(tpl.Version); err != nil {
				return fmt.Errorf("%s: %w", inPath, err)
			}

			info, err := nameutils.ParseTemplateFileName(fn)
			if err != nil {
				return fmt.Errorf("%s: %w", inPath, err)
			}
			if info.Slug != tpl.Slug || info.Version != tpl.Version {
				return fmt.Errorf("%s: filename (slug=%q,ver=%q) ≠ JSON (slug=%q,ver=%q)",
					inPath, info.Slug, info.Version, tpl.Slug, tpl.Version)
			}

			if prev := seenTpl[tpl.ID]; prev != "" {
				return fmt.Errorf("%s: duplicate template ID %s (also %s)",
					inPath, tpl.ID, prev)
			}
			seenTpl[tpl.ID] = inPath

			templateMap[bundleID][tpl.ID] = tpl
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	data := &BuiltInData{
		overlayBaseDir: overlayBaseDir,
		bundles:        bundleMap,
		templates:      templateMap,
		store:          store,
	}
	data.mu.Lock()
	if err := data.rebuildSnapshot(); err != nil {
		data.mu.Unlock()
		return nil, err
	}
	data.mu.Unlock()

	data.rebuilder = builtin.NewAsyncRebuilder(
		snapshotMaxAge,
		func() error {
			data.mu.Lock()
			defer data.mu.Unlock()
			return data.rebuildSnapshot()
		},
	)
	data.rebuilder.MarkFresh()

	return data, nil
}

// SetBundleEnabled toggles a bundle and schedules a rebuild if needed.
func (d *BuiltInData) SetBundleEnabled(id spec.BundleID, enabled bool) error {
	if _, ok := d.bundles[id]; !ok {
		return errors.New("bundle not found in built-in data")
	}
	if err := d.store.SetEnabled(BuiltInBundleID(id), enabled); err != nil {
		return err
	}

	d.mu.Lock()
	b := d.viewBundles[id]
	b.IsEnabled = enabled
	d.viewBundles[id] = b
	d.mu.Unlock()

	d.rebuilder.Trigger()
	return nil
}

// SetTemplateEnabled toggles a template flag and schedules a rebuild if needed.
func (d *BuiltInData) SetTemplateEnabled(
	bundleID spec.BundleID,
	templateID spec.TemplateID,
	enabled bool,
) error {
	if _, ok := d.templates[bundleID]; !ok {
		return errors.New("bundle not found in built-in data")
	}
	if _, ok := d.templates[bundleID][templateID]; !ok {
		return errors.New("template not found in built-in data")
	}
	if err := d.store.SetEnabled(BuiltInTemplateID(templateID), enabled); err != nil {
		return err
	}

	d.mu.Lock()
	t := d.viewTemplates[bundleID][templateID]
	t.IsEnabled = enabled
	d.viewTemplates[bundleID][templateID] = t
	d.mu.Unlock()

	d.rebuilder.Trigger()
	return nil
}

// rebuildSnapshot regenerates the overlay-applied view.
// Assumes that mu.Lock is held by caller.
func (d *BuiltInData) rebuildSnapshot() error {
	newBundles := make(map[spec.BundleID]spec.PromptBundle, len(d.bundles))
	newTemplates := make(
		map[spec.BundleID]map[spec.TemplateID]spec.PromptTemplate,
		len(d.templates),
	)

	for id, b := range d.bundles {
		en, err := d.store.IsEnabled(BuiltInBundleID(id), b.IsEnabled)
		if err != nil {
			return err
		}
		bc := b
		bc.IsEnabled = en
		newBundles[id] = bc
	}

	for bid, tm := range d.templates {
		sub := make(map[spec.TemplateID]spec.PromptTemplate, len(tm))
		for tid, t := range tm {
			en, err := d.store.IsEnabled(BuiltInTemplateID(tid), t.IsEnabled)
			if err != nil {
				return err
			}
			tc := t
			tc.IsEnabled = en
			sub[tid] = tc
		}
		newTemplates[bid] = sub
	}

	d.viewBundles = newBundles
	d.viewTemplates = newTemplates

	return nil
}

func cloneTemplates(
	src map[spec.BundleID]map[spec.TemplateID]spec.PromptTemplate,
) map[spec.BundleID]map[spec.TemplateID]spec.PromptTemplate {
	dst := make(map[spec.BundleID]map[spec.TemplateID]spec.PromptTemplate, len(src))
	for bid, inner := range src {
		// Clone each inner map so callers can’t mutate the originals.
		dst[bid] = maps.Clone(inner)
	}
	return dst
}

// List returns a deep copy of the cached snapshot.
func (d *BuiltInData) List() (
	bundles map[spec.BundleID]spec.PromptBundle,
	templates map[spec.BundleID]map[spec.TemplateID]spec.PromptTemplate,
) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	bundles = maps.Clone(d.viewBundles)
	templates = cloneTemplates(d.viewTemplates)
	return bundles, templates
}
