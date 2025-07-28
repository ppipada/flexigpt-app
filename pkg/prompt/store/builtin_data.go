// Package store manages the read-only built-in prompt assets together with a
// writable overlay that enables or disables individual bundles or templates.
package store

import (
	"encoding/json"
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
	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
)

type BuiltInBundleID bundleitemutils.BundleID

func (BuiltInBundleID) Group() booloverlay.GroupID { return "bundles" }
func (b BuiltInBundleID) ID() booloverlay.KeyID    { return booloverlay.KeyID(b) }

type BuiltInTemplateID bundleitemutils.ItemID

func (BuiltInTemplateID) Group() booloverlay.GroupID { return "templates" }
func (t BuiltInTemplateID) ID() booloverlay.KeyID    { return booloverlay.KeyID(t) }

// BuiltInData keeps the built-in prompt assets and an overlay with enable flags.
type BuiltInData struct {
	bundlesFS      fs.FS
	bundlesDir     string
	overlayBaseDir string

	bundles   map[bundleitemutils.BundleID]spec.PromptBundle
	templates map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate
	store     *booloverlay.Store

	mu            sync.RWMutex
	viewBundles   map[bundleitemutils.BundleID]spec.PromptBundle
	viewTemplates map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate

	rebuilder *builtin.AsyncRebuilder
}

type BuiltInDataOption func(*BuiltInData)

// WithBundlesFS overrides the default embedded FS.
func WithBundlesFS(fsys fs.FS, rootDir string) BuiltInDataOption {
	return func(b *BuiltInData) {
		b.bundlesFS = fsys
		b.bundlesDir = rootDir
	}
}

func NewBuiltInData(
	overlayBaseDir string,
	builtInSnapshotMaxAge time.Duration,
	opts ...BuiltInDataOption,
) (*BuiltInData, error) {
	if builtInSnapshotMaxAge <= 0 {
		builtInSnapshotMaxAge = time.Hour
	}
	if overlayBaseDir == "" {
		return nil, fmt.Errorf("%w: overlayBaseDir", spec.ErrInvalidDir)
	}
	if err := os.MkdirAll(overlayBaseDir, 0o755); err != nil {
		return nil, err
	}
	store, err := booloverlay.NewStore(
		filepath.Join(overlayBaseDir, spec.PromptBuiltInOverlayFileName),
		booloverlay.WithKeyType[BuiltInBundleID](),
		booloverlay.WithKeyType[BuiltInTemplateID](),
	)
	if err != nil {
		return nil, err
	}

	data := &BuiltInData{
		bundlesFS:      builtin.BuiltInPromptBundlesFS,
		bundlesDir:     builtin.BuiltInPromptBundlesRootDir,
		overlayBaseDir: overlayBaseDir,
		store:          store,
	}
	for _, o := range opts {
		o(data)
	}

	if err := data.populateDataFromFS(); err != nil {
		return nil, err
	}
	data.rebuilder = builtin.NewAsyncRebuilder(
		builtInSnapshotMaxAge,
		func() error {
			data.mu.Lock()
			defer data.mu.Unlock()
			return data.rebuildSnapshot()
		},
	)
	data.rebuilder.MarkFresh()

	return data, nil
}

// ListBuiltInData returns a deep copy of the cached snapshot.
func (d *BuiltInData) ListBuiltInData() (
	bundles map[bundleitemutils.BundleID]spec.PromptBundle,
	templates map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate,
	err error,
) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	bundles = maps.Clone(d.viewBundles)
	templates = cloneTemplates(d.viewTemplates)
	return bundles, templates, nil
}

// SetBundleEnabled toggles a bundle flag.
func (d *BuiltInData) SetBundleEnabled(
	id bundleitemutils.BundleID,
	enabled bool,
) (bundle spec.PromptBundle, err error) {
	if _, ok := d.bundles[id]; !ok {
		return spec.PromptBundle{}, fmt.Errorf(
			"bundleID: %q, err: %w",
			id,
			spec.ErrBuiltInBundleNotFound,
		)
	}
	flag, err := d.store.SetFlag(BuiltInBundleID(id), enabled)
	if err != nil {
		return spec.PromptBundle{}, err
	}

	d.mu.Lock()
	b := d.viewBundles[id]
	b.IsEnabled = enabled
	b.ModifiedAt = flag.ModifiedAt // update timestamp from overlay
	d.viewBundles[id] = b
	d.mu.Unlock()

	d.rebuilder.Trigger()
	return b, nil
}

func (d *BuiltInData) GetBuiltInBundle(id bundleitemutils.BundleID) (spec.PromptBundle, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	b, ok := d.viewBundles[id]
	if !ok {
		return spec.PromptBundle{}, spec.ErrBundleNotFound
	}
	return b, nil
}

// SetTemplateEnabled toggles a template flag.
func (d *BuiltInData) SetTemplateEnabled(
	bundleID bundleitemutils.BundleID,
	slug bundleitemutils.ItemSlug,
	version bundleitemutils.ItemVersion,
	enabled bool,
) (template spec.PromptTemplate, err error) {
	template, err = d.GetBuiltInTemplate(bundleID, slug, version)
	if err != nil {
		return template, err
	}
	flag, err := d.store.SetFlag(BuiltInTemplateID(template.ID), enabled)
	if err != nil {
		return spec.PromptTemplate{}, err
	}

	d.mu.Lock()
	template.IsEnabled = enabled
	template.ModifiedAt = flag.ModifiedAt // update timestamp from overlay
	d.viewTemplates[bundleID][template.ID] = template
	d.mu.Unlock()

	d.rebuilder.Trigger()
	return template, nil
}

func (d *BuiltInData) GetBuiltInTemplate(
	bundleID bundleitemutils.BundleID,
	slug bundleitemutils.ItemSlug,
	version bundleitemutils.ItemVersion,
) (spec.PromptTemplate, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	templates, ok := d.viewTemplates[bundleID]
	if !ok {
		return spec.PromptTemplate{}, spec.ErrBundleNotFound
	}
	for _, tpl := range templates {
		if tpl.Slug == slug && tpl.Version == version {
			return tpl, nil
		}
	}
	return spec.PromptTemplate{}, spec.ErrTemplateNotFound
}

func (d *BuiltInData) populateDataFromFS() error {
	bundlesFS, err := resolveBundlesFS(d.bundlesFS, d.bundlesDir)
	if err != nil {
		return err
	}
	rawBundles, err := fs.ReadFile(bundlesFS, builtin.BuiltInPromptBundlesJSON)
	if err != nil {
		return err
	}
	var manifest spec.AllBundles
	if err := json.Unmarshal(rawBundles, &manifest); err != nil {
		return err
	}

	bundleMap := make(map[bundleitemutils.BundleID]spec.PromptBundle, len(manifest.Bundles))
	templateMap := make(map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate)
	for id, b := range manifest.Bundles {
		b.IsBuiltIn = true
		bundleMap[id] = b
		templateMap[id] = make(map[bundleitemutils.ItemID]spec.PromptTemplate)
	}
	if len(bundleMap) == 0 {
		return fmt.Errorf(
			"built-in data: %s/%s contains no bundles",
			builtin.BuiltInPromptBundlesJSON,
			d.bundlesDir,
		)
	}

	seenTpl := make(map[bundleitemutils.ItemID]string)
	err = fs.WalkDir(
		bundlesFS,
		".",
		func(inPath string, de fs.DirEntry, _ error) error {
			if de.IsDir() || path.Ext(inPath) != ".json" {
				return nil
			}
			fn := path.Base(inPath)
			if fn == builtin.BuiltInPromptBundlesJSON || fn == spec.PromptBuiltInOverlayFileName {
				return nil
			}

			dir := path.Base(path.Dir(inPath))
			dirInfo, derr := bundleitemutils.ParseBundleDir(dir)
			if derr != nil {
				return fmt.Errorf("%s: %w", inPath, derr)
			}
			bundleID := dirInfo.ID

			bDef, ok := bundleMap[bundleID]
			if !ok {
				return fmt.Errorf(
					"%s: bundle dir %q not in %s",
					inPath,
					bundleID,
					builtin.BuiltInPromptBundlesJSON,
				)
			}
			if dirInfo.Slug != bDef.Slug {
				return fmt.Errorf("%s: dir slug %q not equal to manifest slug %q",
					inPath, dirInfo.Slug, bDef.Slug)
			}

			raw, err := fs.ReadFile(bundlesFS, inPath)
			if err != nil {
				return err
			}
			var tpl spec.PromptTemplate
			if err := json.Unmarshal(raw, &tpl); err != nil {
				return fmt.Errorf("%s: %w", inPath, err)
			}
			tpl.IsBuiltIn = true

			if err := bundleitemutils.ValidateItemSlug(tpl.Slug); err != nil {
				return fmt.Errorf("%s: %w", inPath, err)
			}
			if err := bundleitemutils.ValidateItemVersion(tpl.Version); err != nil {
				return fmt.Errorf("%s: %w", inPath, err)
			}

			info, err := bundleitemutils.ParseItemFileName(fn)
			if err != nil {
				return fmt.Errorf("%s: %w", inPath, err)
			}
			if info.Slug != tpl.Slug ||
				info.Version != tpl.Version {
				return fmt.Errorf(
					"%s: filename (slug=%q,ver=%q) not equal to JSON (slug=%q,ver=%q)",
					inPath,
					info.Slug,
					info.Version,
					tpl.Slug,
					tpl.Version,
				)
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
		return err
	}

	for id, tm := range templateMap {
		if len(tm) == 0 {
			return fmt.Errorf("built-in data: bundle %s has no templates", id)
		}
	}

	d.bundles = bundleMap
	d.templates = templateMap
	d.mu.Lock()
	if err := d.rebuildSnapshot(); err != nil {
		d.mu.Unlock()
		return err
	}
	d.mu.Unlock()
	return nil
}

// rebuildSnapshot regenerates the overlay-applied view.
// Assumes that mu.Lock is held by caller.
func (d *BuiltInData) rebuildSnapshot() error {
	newBundles := make(map[bundleitemutils.BundleID]spec.PromptBundle, len(d.bundles))
	newTemplates := make(
		map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate,
		len(d.templates),
	)

	for id, b := range d.bundles {
		flag, ok, err := d.store.GetFlag(BuiltInBundleID(id))
		if err != nil {
			return err
		}
		bc := b
		if ok {
			bc.IsEnabled = flag.Enabled
			bc.ModifiedAt = flag.ModifiedAt // take overlay timestamp
		}
		newBundles[id] = bc
	}

	for bid, tm := range d.templates {
		sub := make(map[bundleitemutils.ItemID]spec.PromptTemplate, len(tm))
		for tid, t := range tm {
			flag, ok, err := d.store.GetFlag(BuiltInTemplateID(tid))
			if err != nil {
				return err
			}
			tc := t
			if ok {
				tc.IsEnabled = flag.Enabled
				tc.ModifiedAt = flag.ModifiedAt // take overlay timestamp
			}
			sub[tid] = tc
		}
		newTemplates[bid] = sub
	}

	d.viewBundles = newBundles
	d.viewTemplates = newTemplates
	return nil
}

func cloneTemplates(
	src map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate,
) map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate {
	dst := make(
		map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate,
		len(src),
	)
	for bid, inner := range src {
		dst[bid] = maps.Clone(inner)
	}
	return dst
}

func resolveBundlesFS(fsys fs.FS, dir string) (fs.FS, error) {
	if dir == "" || dir == "." {
		return fsys, nil
	}
	return fs.Sub(fsys, dir)
}
