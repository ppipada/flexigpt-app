// Package store keeps the read-only built-in tool assets together with
// a writable overlay that enables or disables individual bundles or
// tools.
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
	"github.com/ppipada/flexigpt-app/pkg/tool/spec"
)

type BuiltInToolBundleID bundleitemutils.BundleID

func (BuiltInToolBundleID) Group() booloverlay.GroupID { return "bundles" }
func (b BuiltInToolBundleID) ID() booloverlay.KeyID    { return booloverlay.KeyID(b) }

type BuiltInToolID bundleitemutils.ItemID

func (BuiltInToolID) Group() booloverlay.GroupID { return "tools" }
func (t BuiltInToolID) ID() booloverlay.KeyID    { return booloverlay.KeyID(t) }

type BuiltInToolData struct {
	toolsFS        fs.FS
	toolsDir       string
	overlayBaseDir string

	bundles map[bundleitemutils.BundleID]spec.ToolBundle
	tools   map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool
	store   *booloverlay.Store

	mu          sync.RWMutex
	viewBundles map[bundleitemutils.BundleID]spec.ToolBundle
	viewTools   map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool

	rebuilder *builtin.AsyncRebuilder
}

type BuiltInToolDataOption func(*BuiltInToolData)

func WithToolBundlesFS(fsys fs.FS, rootDir string) BuiltInToolDataOption {
	return func(d *BuiltInToolData) {
		d.toolsFS = fsys
		d.toolsDir = rootDir
	}
}

func NewBuiltInToolData(
	overlayBaseDir string,
	snapshotMaxAge time.Duration,
	opts ...BuiltInToolDataOption,
) (*BuiltInToolData, error) {
	if snapshotMaxAge <= 0 {
		snapshotMaxAge = time.Hour
	}
	if overlayBaseDir == "" {
		return nil, fmt.Errorf("%w: overlayBaseDir", spec.ErrInvalidDir)
	}
	if err := os.MkdirAll(overlayBaseDir, 0o755); err != nil {
		return nil, err
	}

	store, err := booloverlay.NewStore(
		filepath.Join(overlayBaseDir, spec.ToolBuiltInOverlayFileName),
		booloverlay.WithKeyType[BuiltInToolBundleID](),
		booloverlay.WithKeyType[BuiltInToolID](),
	)
	if err != nil {
		return nil, err
	}

	d := &BuiltInToolData{
		toolsFS:        builtin.BuiltInToolBundlesFS,
		toolsDir:       builtin.BuiltInToolBundlesRootDir,
		overlayBaseDir: overlayBaseDir,
		store:          store,
	}
	for _, o := range opts {
		o(d)
	}

	if err := d.populateDataFromFS(); err != nil {
		return nil, err
	}

	d.rebuilder = builtin.NewAsyncRebuilder(
		snapshotMaxAge,
		func() error {
			d.mu.Lock()
			defer d.mu.Unlock()
			return d.rebuildSnapshot()
		},
	)
	d.rebuilder.MarkFresh()

	return d, nil
}

// SetToolBundleEnabled toggles a bundle flag and returns the updated bundle.
func (d *BuiltInToolData) SetToolBundleEnabled(
	id bundleitemutils.BundleID,
	enabled bool,
) (spec.ToolBundle, error) {
	if _, ok := d.bundles[id]; !ok {
		return spec.ToolBundle{}, fmt.Errorf(
			"bundleID: %q, err: %w", id, spec.ErrBuiltInBundleNotFound)
	}

	flag, err := d.store.SetFlag(BuiltInToolBundleID(id), enabled)
	if err != nil {
		return spec.ToolBundle{}, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	b := d.viewBundles[id]
	b.IsEnabled = enabled
	b.ModifiedAt = flag.ModifiedAt
	d.viewBundles[id] = b

	d.rebuilder.Trigger()

	return b, nil
}

// SetToolEnabled toggles one tool (identified by slug+version) and
// returns the updated object.  Deriving the primary key from slug &
// version keeps the API parallel to the prompt store.
func (d *BuiltInToolData) SetToolEnabled(
	bundleID bundleitemutils.BundleID,
	slug bundleitemutils.ItemSlug,
	version bundleitemutils.ItemVersion,
	enabled bool,
) (spec.Tool, error) {
	tool, err := d.GetBuiltInTool(bundleID, slug, version)
	if err != nil {
		return spec.Tool{}, err
	}
	flag, err := d.store.SetFlag(BuiltInToolID(tool.ID), enabled)
	if err != nil {
		return spec.Tool{}, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	tool.IsEnabled = enabled
	tool.ModifiedAt = flag.ModifiedAt
	d.viewTools[bundleID][tool.ID] = tool

	d.rebuilder.Trigger()

	return tool, nil
}

func (d *BuiltInToolData) ListBuiltInToolData() (
	bundles map[bundleitemutils.BundleID]spec.ToolBundle,
	tools map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool,
	err error,
) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	bundles = maps.Clone(d.viewBundles)
	tools = cloneToolSpecs(d.viewTools)
	return bundles, tools, nil
}

func (d *BuiltInToolData) GetBuiltInToolBundle(
	id bundleitemutils.BundleID,
) (spec.ToolBundle, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	b, ok := d.viewBundles[id]
	if !ok {
		return spec.ToolBundle{}, spec.ErrBundleNotFound
	}
	return b, nil
}

func (d *BuiltInToolData) GetBuiltInTool(
	bundleID bundleitemutils.BundleID,
	slug bundleitemutils.ItemSlug,
	version bundleitemutils.ItemVersion,
) (spec.Tool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	tools, ok := d.viewTools[bundleID]
	if !ok {
		return spec.Tool{}, spec.ErrBundleNotFound
	}
	for _, tl := range tools {
		if tl.Slug == slug && tl.Version == version {
			return tl, nil
		}
	}
	return spec.Tool{}, spec.ErrToolNotFound
}

func (d *BuiltInToolData) populateDataFromFS() error {
	toolsFS, err := resolveToolBundlesFS(d.toolsFS, d.toolsDir)
	if err != nil {
		return err
	}

	rawManifest, err := fs.ReadFile(toolsFS, builtin.BuiltInToolBundlesJSON)
	if err != nil {
		return err
	}

	var manifest spec.AllBundles
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		return err
	}

	bundleMap := make(
		map[bundleitemutils.BundleID]spec.ToolBundle,
		len(manifest.Bundles),
	)
	toolMap := make(
		map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool,
	)
	for id, b := range manifest.Bundles {
		b.IsBuiltIn = true
		bundleMap[id] = b
		toolMap[id] = make(map[bundleitemutils.ItemID]spec.Tool)
	}
	if len(bundleMap) == 0 {
		return fmt.Errorf("built-in data: %s/%s contains no bundles",
			d.toolsDir, builtin.BuiltInToolBundlesJSON)
	}

	seenTools := make(map[bundleitemutils.ItemID]string)

	err = fs.WalkDir(
		toolsFS,
		".",
		func(inPath string, de fs.DirEntry, _ error) error {
			if de.IsDir() || path.Ext(inPath) != ".json" {
				return nil
			}
			fn := path.Base(inPath)
			if fn == builtin.BuiltInToolBundlesJSON || fn == spec.ToolBuiltInOverlayFileName {
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
					builtin.BuiltInToolBundlesJSON,
				)
			}
			if dirInfo.Slug != bDef.Slug {
				return fmt.Errorf("%s: dir slug %q not equal to manifest slug %q",
					inPath, dirInfo.Slug, bDef.Slug)
			}

			raw, err := fs.ReadFile(toolsFS, inPath)
			if err != nil {
				return err
			}
			var tool spec.Tool
			if err := json.Unmarshal(raw, &tool); err != nil {
				return fmt.Errorf("%s: %w", inPath, err)
			}
			tool.IsBuiltIn = true

			if err := validateTool(&tool); err != nil {
				return fmt.Errorf("%s: invalid tool: %w", inPath, err)
			}

			info, err := bundleitemutils.ParseItemFileName(fn)
			if err != nil {
				return fmt.Errorf("%s: %w", inPath, err)
			}
			if info.Slug != tool.Slug || info.Version != tool.Version {
				return fmt.Errorf(
					"%s: filename (slug=%q,ver=%q) not equal to JSON (slug=%q,ver=%q)",
					inPath,
					info.Slug,
					info.Version,
					tool.Slug,
					tool.Version,
				)
			}

			if prev := seenTools[tool.ID]; prev != "" {
				return fmt.Errorf("%s: duplicate tool ID %s (also %s)",
					inPath, tool.ID, prev)
			}
			seenTools[tool.ID] = inPath
			toolMap[bundleID][tool.ID] = tool
			return nil
		},
	)
	if err != nil {
		return err
	}

	for id, tm := range toolMap {
		if len(tm) == 0 {
			return fmt.Errorf("built-in data: bundle %s has no tools", id)
		}
	}

	d.bundles = bundleMap
	d.tools = toolMap

	d.mu.Lock()
	if err := d.rebuildSnapshot(); err != nil {
		d.mu.Unlock()
		return err
	}
	d.mu.Unlock()

	return nil
}

// Assumes d.mu is already locked.
func (d *BuiltInToolData) rebuildSnapshot() error {
	newBundles := make(
		map[bundleitemutils.BundleID]spec.ToolBundle,
		len(d.bundles),
	)
	newTools := make(
		map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool,
		len(d.tools),
	)

	for id, b := range d.bundles {
		flag, ok, err := d.store.GetFlag(BuiltInToolBundleID(id))
		if err != nil {
			return err
		}
		bc := b
		if ok {
			bc.IsEnabled = flag.Enabled
			bc.ModifiedAt = flag.ModifiedAt
		}
		newBundles[id] = bc
	}

	for bid, tm := range d.tools {
		sub := make(map[bundleitemutils.ItemID]spec.Tool, len(tm))
		for tid, t := range tm {
			flag, ok, err := d.store.GetFlag(BuiltInToolID(tid))
			if err != nil {
				return err
			}
			tc := t
			if ok {
				tc.IsEnabled = flag.Enabled
				tc.ModifiedAt = flag.ModifiedAt
			}
			sub[tid] = tc
		}
		newTools[bid] = sub
	}

	d.viewBundles = newBundles
	d.viewTools = newTools
	return nil
}

func resolveToolBundlesFS(fsys fs.FS, dir string) (fs.FS, error) {
	if dir == "" || dir == "." {
		return fsys, nil
	}
	return fs.Sub(fsys, dir)
}

func cloneToolSpecs(
	src map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool,
) map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool {
	dst := make(
		map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool,
		len(src),
	)
	for bid, inner := range src {
		dst[bid] = maps.Clone(inner)
	}
	return dst
}
