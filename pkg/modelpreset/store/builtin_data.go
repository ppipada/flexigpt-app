// Package store manages read-only built-in provider/model presets together with
// a writable overlay that enables or disables individual entities.
package store

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"            // POSIX for embed.FS
	"path/filepath" // Native paths
	"sync"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/booloverlay"
	"github.com/ppipada/flexigpt-app/pkg/builtin"
	"github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

type builtInProviderKey spec.ProviderName

func (builtInProviderKey) Group() booloverlay.GroupID { return "providers" }
func (k builtInProviderKey) ID() booloverlay.KeyID    { return booloverlay.KeyID(k) }

type builtInModelKey spec.ModelPresetID

func (builtInModelKey) Group() booloverlay.GroupID { return "models" }
func (k builtInModelKey) ID() booloverlay.KeyID    { return booloverlay.KeyID(k) }

// BuiltInPresets loads built-in preset assets and maintains an overlay store.
type BuiltInPresets struct {
	// Immutable original data.
	providers map[spec.ProviderName]spec.ProviderPreset
	models    map[spec.ProviderName]map[spec.ModelPresetID]spec.ModelPreset

	// View after overlay application, guarded by mu.
	mu         sync.RWMutex
	viewProv   map[spec.ProviderName]spec.ProviderPreset
	viewModels map[spec.ProviderName]map[spec.ModelPresetID]spec.ModelPreset

	// IO.
	presetsFS      fs.FS
	presetsDir     string
	overlayBaseDir string
	store          *booloverlay.Store
	rebuilder      *builtin.AsyncRebuilder
}

type PresetStoreOption func(*BuiltInPresets)

// WithModelPresetsFS sets a custom FS + root for tests.
func WithModelPresetsFS(fsys fs.FS, root string) PresetStoreOption {
	return func(b *BuiltInPresets) {
		b.presetsFS = fsys
		b.presetsDir = root
	}
}

// NewBuiltInPresets prepares the presets store and loads a first snapshot.
func NewBuiltInPresets(
	overlayBaseDir string,
	maxSnapshotAge time.Duration,
	opts ...PresetStoreOption,
) (*BuiltInPresets, error) {
	if overlayBaseDir == "" {
		return nil, fmt.Errorf("%w: overlayBaseDir", spec.ErrInvalidDir)
	}
	if maxSnapshotAge <= 0 {
		maxSnapshotAge = time.Hour
	}
	if err := os.MkdirAll(overlayBaseDir, 0o755); err != nil {
		return nil, err
	}

	overlay, err := booloverlay.NewStore(
		filepath.Join(overlayBaseDir, spec.ModelPresetsBuiltInOverlayFileName),
		booloverlay.WithKeyType[builtInProviderKey](),
		booloverlay.WithKeyType[builtInModelKey](),
	)
	if err != nil {
		return nil, err
	}

	b := &BuiltInPresets{
		presetsFS:      builtin.BuiltInModelPresetsFS,
		presetsDir:     builtin.BuiltInModelPresetsRootDir,
		overlayBaseDir: overlayBaseDir,
		store:          overlay,
	}
	for _, o := range opts {
		o(b)
	}
	if err := b.loadFromFS(); err != nil {
		return nil, err
	}

	b.rebuilder = builtin.NewAsyncRebuilder(
		maxSnapshotAge,
		func() error {
			b.mu.Lock()
			defer b.mu.Unlock()
			return b.rebuildSnapshot()
		},
	)
	b.rebuilder.MarkFresh()
	return b, nil
}

// ListBuiltInPresets returns deep-copied snapshots.
func (b *BuiltInPresets) ListBuiltInPresets() (
	providerPresets map[spec.ProviderName]spec.ProviderPreset,
	modelPresets map[spec.ProviderName]map[spec.ModelPresetID]spec.ModelPreset,
	err error,
) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return maps.Clone(b.viewProv), cloneModels(b.viewModels), nil
}

// GetBuiltInProvider fetches a provider from the snapshot.
func (b *BuiltInPresets) GetBuiltInProvider(name spec.ProviderName) (spec.ProviderPreset, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	p, ok := b.viewProv[name]
	if !ok {
		return spec.ProviderPreset{}, spec.ErrProviderNotFound
	}
	return p, nil
}

// GetBuiltInModelPreset fetches a model preset.
func (b *BuiltInPresets) GetBuiltInModelPreset(
	provider spec.ProviderName,
	modelID spec.ModelPresetID,
) (spec.ModelPreset, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	pm, ok := b.viewModels[provider]
	if !ok {
		return spec.ModelPreset{}, spec.ErrProviderNotFound
	}
	mp, ok := pm[modelID]
	if !ok {
		return spec.ModelPreset{}, spec.ErrModelPresetNotFound
	}
	return mp, nil
}

// SetProviderEnabled toggles a provider.
func (b *BuiltInPresets) SetProviderEnabled(
	name spec.ProviderName,
	enabled bool,
) (spec.ProviderPreset, error) {
	if _, ok := b.providers[name]; !ok {
		return spec.ProviderPreset{}, spec.ErrBuiltInProviderAbsent
	}
	flag, err := b.store.SetFlag(builtInProviderKey(name), enabled)
	if err != nil {
		return spec.ProviderPreset{}, err
	}

	b.mu.Lock()
	pp := b.viewProv[name]
	pp.IsEnabled = enabled
	pp.ModifiedAt = flag.ModifiedAt
	b.viewProv[name] = pp
	b.mu.Unlock()

	b.rebuilder.Trigger()
	return pp, nil
}

// SetModelPresetEnabled toggles a model preset.
func (b *BuiltInPresets) SetModelPresetEnabled(
	provider spec.ProviderName,
	modelID spec.ModelPresetID,
	enabled bool,
) (spec.ModelPreset, error) {
	mp, err := b.GetBuiltInModelPreset(provider, modelID)
	if err != nil {
		return mp, err
	}
	flag, err := b.store.SetFlag(builtInModelKey(modelID), enabled)
	if err != nil {
		return spec.ModelPreset{}, err
	}

	b.mu.Lock()
	mp.IsEnabled = enabled
	mp.ModifiedAt = flag.ModifiedAt
	b.viewModels[provider][modelID] = mp
	b.mu.Unlock()

	b.rebuilder.Trigger()
	return mp, nil
}

func (b *BuiltInPresets) loadFromFS() error {
	subFS, err := resolvePresetsFS(b.presetsFS, b.presetsDir)
	if err != nil {
		return err
	}
	raw, err := fs.ReadFile(subFS, builtin.BuiltInModelPresetsJSON)
	if err != nil {
		return err
	}

	var schema spec.PresetsSchema
	if err := json.Unmarshal(raw, &schema); err != nil {
		return err
	}
	if schema.Version != spec.SchemaVersion {
		return fmt.Errorf("schemaVersion %q ≠ %q",
			schema.Version, spec.SchemaVersion)
	}
	if len(schema.ProviderPresets) == 0 {
		return fmt.Errorf("%s contains no providers", builtin.BuiltInModelPresetsJSON)
	}

	// Parse + validate.
	prov := make(map[spec.ProviderName]spec.ProviderPreset, len(schema.ProviderPresets))
	models := make(map[spec.ProviderName]map[spec.ModelPresetID]spec.ModelPreset)
	seenModelGlobal := map[spec.ModelPresetID]struct{}{}

	for name, pp := range schema.ProviderPresets {
		if err := validateProviderPreset(&pp); err != nil {
			return err
		}
		pp.IsBuiltIn = true
		prov[name] = pp

		sub := make(map[spec.ModelPresetID]spec.ModelPreset, len(pp.ModelPresets))
		for mid, mp := range pp.ModelPresets {
			if _, dup := seenModelGlobal[mid]; dup {
				return fmt.Errorf("duplicate modelPresetID %q across providers", mid)
			}
			seenModelGlobal[mid] = struct{}{}

			mp.IsBuiltIn = true
			sub[mid] = mp
		}
		models[name] = sub
	}

	b.providers = prov
	b.models = models

	b.mu.Lock()
	defer b.mu.Unlock()
	return b.rebuildSnapshot()
}

// rebuildSnapshot applies overlay flags onto the immutable base sets.
// Caller must hold write lock.
func (b *BuiltInPresets) rebuildSnapshot() error {
	newProv := make(map[spec.ProviderName]spec.ProviderPreset, len(b.providers))
	newModels := make(map[spec.ProviderName]map[spec.ModelPresetID]spec.ModelPreset, len(b.models))

	for pname, p := range b.providers {
		if flag, ok, err := b.store.GetFlag(builtInProviderKey(pname)); err != nil {
			return err
		} else if ok {
			p.IsEnabled = flag.Enabled
			p.ModifiedAt = flag.ModifiedAt
		}
		newProv[pname] = p
	}

	for pname, mm := range b.models {
		sub := make(map[spec.ModelPresetID]spec.ModelPreset, len(mm))
		for mid, m := range mm {
			if flag, ok, err := b.store.GetFlag(builtInModelKey(mid)); err != nil {
				return err
			} else if ok {
				m.IsEnabled = flag.Enabled
				m.ModifiedAt = flag.ModifiedAt
			}
			sub[mid] = m
		}
		newModels[pname] = sub
	}

	b.viewProv = newProv
	b.viewModels = newModels
	return nil
}

func cloneModels(
	src map[spec.ProviderName]map[spec.ModelPresetID]spec.ModelPreset,
) map[spec.ProviderName]map[spec.ModelPresetID]spec.ModelPreset {
	dst := make(map[spec.ProviderName]map[spec.ModelPresetID]spec.ModelPreset, len(src))
	for pname, inner := range src {
		dst[pname] = maps.Clone(inner)
	}
	return dst
}

func resolvePresetsFS(fsys fs.FS, dir string) (fs.FS, error) {
	if dir == "" || dir == "." {
		return fsys, nil
	}
	return fs.Sub(fsys, dir)
}
