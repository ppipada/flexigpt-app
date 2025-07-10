// Package store implements the prompt template storage and management logic.
// It provides CRUD operations for prompt bundles and templates, supports soft deletion,
// background cleanup, and optional full-text search (FTS) integration.
package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/dirstore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
	"github.com/ppipada/flexigpt-app/pkg/uuidv7filename"
)

const (
	maxPageSize     = 256              // Maximum allowed page size for listing.
	defPageSize     = 25               // Default page size for listing.
	softDeleteGrace = 60 * time.Minute // Grace period before hard-deleting a soft-deleted bundle.
	cleanupInterval = 24 * time.Hour   // Interval for periodic cleanup sweep.
)

// Error types for better error handling.
var (
	ErrBundleNotFound   = errors.New("bundle not found")
	ErrTemplateNotFound = errors.New("template not found")
	ErrBundleDisabled   = errors.New("bundle is disabled")
	ErrBundleDeleting   = errors.New("bundle is being deleted")
	ErrConflict         = errors.New("resource already exists")
	ErrBundleNotEmpty   = errors.New("bundle still contains templates")
	ErrInvalidRequest   = errors.New("invalid request")
	ErrFTSDisabled      = errors.New("FTS is disabled")
)

// nullableStr returns a pointer to the string, or nil if the string is empty.
func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// slugLocks manages a RW-mutex per bundleID|slug for concurrency control on templates.
// Note: This implementation allows indefinite growth of the lock map, which is acceptable
// for the current use case as the number of unique bundle|slug combinations is expected
// to be relatively small and bounded by actual usage patterns.
type slugLocks struct {
	mu sync.Mutex
	m  map[string]*sync.RWMutex
}

// newSlugLocks creates a new slugLocks instance.
func newSlugLocks() *slugLocks {
	return &slugLocks{m: map[string]*sync.RWMutex{}}
}

// lockKey returns the mutex for a given bundleID and slug, creating it if necessary.
func (l *slugLocks) lockKey(bundleID spec.BundleID, slug spec.TemplateSlug) *sync.RWMutex {
	k := string(bundleID) + "|" + string(slug)
	l.mu.Lock()
	defer l.mu.Unlock()
	if lk, ok := l.m[k]; ok {
		return lk
	}
	lk := &sync.RWMutex{}
	l.m[k] = lk
	return lk
}

// validateTemplate validates the structure and content of a prompt template.
// This is a placeholder implementation that will be filled in later.
func validateTemplate(tpl *spec.PromptTemplate) error {
	// TODO: Implement comprehensive template validation:
	// - Validate Variables reference existing placeholders in Blocks.
	// - Validate PreProcessors reference valid tools.
	// - Validate MessageBlock roles are valid.
	// - Validate variable types and constraints.
	// - Check for circular dependencies in preprocessors.
	return nil
}

// PromptTemplateStore is the main store for prompt bundles and templates.
// It manages bundle and template CRUD, soft deletion, background cleanup, and FTS integration.
type PromptTemplateStore struct {
	baseDir       string
	bundleStore   *filestore.MapFileStore
	templateStore *dirstore.MapDirectoryStore
	pp            dirstore.PartitionProvider

	enableFTS bool
	fts       *ftsengine.Engine

	cleanOnce sync.Once
	cleanKick chan struct{}
	cleanCtx  context.Context
	cleanStop context.CancelFunc
	wg        sync.WaitGroup

	slugLock *slugLocks

	// Mutex to coordinate sweep operations with bundle modifications.
	sweepMu sync.RWMutex
}

// Option is a functional option for PromptTemplateStore.
type Option func(*PromptTemplateStore) error

// WithFTS enables or disables FTS integration for the store.
func WithFTS(enabled bool) Option {
	return func(s *PromptTemplateStore) error {
		s.enableFTS = enabled
		return nil
	}
}

// NewPromptTemplateStore creates a new PromptTemplateStore at the given base directory.
// It applies any provided options (e.g., enabling FTS).
func NewPromptTemplateStore(baseDir string, opts ...Option) (*PromptTemplateStore, error) {
	s := &PromptTemplateStore{
		baseDir: filepath.Clean(baseDir),
		pp:      &BundlePartitionProvider{},
	}
	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}

	// Initialize bundle meta store (single JSON file).
	def, err := encdec.StructWithJSONTagsToMap(
		spec.AllBundles{Bundles: map[spec.BundleID]spec.PromptBundle{}},
	)
	if err != nil {
		return nil, err
	}
	s.bundleStore, err = filestore.NewMapFileStore(
		filepath.Join(s.baseDir, bundlesMetaFileName),
		def,
		filestore.WithCreateIfNotExists(true),
		filestore.WithAutoFlush(true),
		filestore.WithEncoderDecoder(encdec.JSONEncoderDecoder{}),
	)
	if err != nil {
		return nil, err
	}

	// Initialize FTS engine if enabled.
	if s.enableFTS {
		s.fts, err = ftsengine.NewEngine(ftsengine.Config{
			BaseDir:    s.baseDir,
			DBFileName: sqliteDBFileName,
			Table:      "prompttemplates",
			Columns: []ftsengine.Column{
				{Name: "slug", Weight: 1},
				{Name: "displayName", Weight: 2},
				{Name: "desc", Weight: 3},
				{Name: "messages", Weight: 4},
				{Name: "tags", Weight: 5},
				{Name: "enabled", Unindexed: true},
				{Name: "bundleId", Unindexed: true},
				{Name: "mtime", Unindexed: true},
			},
		})
		if err != nil {
			return nil, err
		}
		StartRebuild(context.Background(), s.baseDir, s.fts)
	}

	// Initialize template directory store (per-bundle folder).
	dirOpts := []dirstore.Option{dirstore.WithPartitionProvider(s.pp)}
	if s.fts != nil {
		dirOpts = append(dirOpts, dirstore.WithListeners(NewFTSListener(s.fts)))
	}
	s.templateStore, err = dirstore.NewMapDirectoryStore(s.baseDir, true, dirOpts...)
	if err != nil {
		return nil, err
	}

	s.slugLock = newSlugLocks()
	s.startCleanupLoop()
	slog.Info("Prompt-store ready.", "baseDir", s.baseDir, "fts", s.enableFTS)
	return s, nil
}

// isSoftDeleted returns true if the bundle is soft-deleted.
func isSoftDeleted(b spec.PromptBundle) bool {
	return b.SoftDeletedAt != nil && !b.SoftDeletedAt.IsZero()
}

// readAllBundles loads and decodes the meta-file.
func (s *PromptTemplateStore) readAllBundles(forceFetch bool) (spec.AllBundles, error) {
	raw, err := s.bundleStore.GetAll(forceFetch)
	if err != nil {
		return spec.AllBundles{}, err
	}
	var ab spec.AllBundles
	if err := encdec.MapToStructWithJSONTags(raw, &ab); err != nil {
		return ab, err
	}
	return ab, nil
}

// writeAllBundles encodes and writes the strongly-typed value.
func (s *PromptTemplateStore) writeAllBundles(ab spec.AllBundles) error {
	mp, _ := encdec.StructWithJSONTagsToMap(ab)
	return s.bundleStore.SetAll(mp)
}

// getBundle returns an active bundle (i.e., not soft-deleted) by ID.
func (s *PromptTemplateStore) getBundle(id spec.BundleID) (spec.PromptBundle, error) {
	all, err := s.readAllBundles(false)
	if err != nil {
		return spec.PromptBundle{}, err
	}

	b, ok := all.Bundles[id]
	if !ok {
		return spec.PromptBundle{}, fmt.Errorf("%w: %s", ErrBundleNotFound, id)
	}

	if isSoftDeleted(b) {
		return b, fmt.Errorf("%w: %s", ErrBundleDeleting, id)
	}
	return b, nil
}

// startCleanupLoop starts the background cleanup goroutine for hard-deleting bundles after the grace period.
func (s *PromptTemplateStore) startCleanupLoop() {
	s.cleanOnce.Do(func() {
		s.cleanKick = make(chan struct{}, 1)
		s.cleanCtx, s.cleanStop = context.WithCancel(context.Background())

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			ticker := time.NewTicker(cleanupInterval)
			defer ticker.Stop()
			defer func() {
				if r := recover(); r != nil {
					slog.Error("Panic in bundle cleanup loop.",
						"err", r,
						"stack", string(debug.Stack()))
				}
			}()

			// Run once at start-up to clean up bundles left after a crash.
			s.sweepSoftDeleted()

			for {
				select {
				case <-s.cleanCtx.Done():
					return
				case <-ticker.C:
					// Periodic sweep.
				case <-s.cleanKick:
					// Explicit sweep triggered by DeletePromptBundle.
				}
				s.sweepSoftDeleted()
			}
		}()
	})
}

// kickCleanupLoop signals the cleanup goroutine to run immediately.
func (s *PromptTemplateStore) kickCleanupLoop() {
	select {
	case s.cleanKick <- struct{}{}:
	default:
		// Queue already has a signal – that is good enough.
	}
}

// Close gracefully terminates the background cleanup goroutine.
func (s *PromptTemplateStore) Close() {
	if s.cleanStop != nil {
		s.cleanStop()
	}
	s.wg.Wait()
}

// SweepSoftDeleted performs hard deletion of bundles that have been soft-deleted and are past the grace period.
// Uses sweepMu to coordinate with bundle operations and prevent race conditions.
func (s *PromptTemplateStore) sweepSoftDeleted() {
	s.sweepMu.Lock()
	defer s.sweepMu.Unlock()

	all, err := s.readAllBundles(false)
	if err != nil {
		slog.Error("Sweep: readAllBundles.", "err", err)
		return
	}
	now := time.Now().UTC()
	changed := false

	for id, b := range all.Bundles {
		if b.SoftDeletedAt == nil || b.SoftDeletedAt.IsZero() {
			continue
		}
		if now.Sub(*b.SoftDeletedAt) < softDeleteGrace {
			continue
		}

		dirInfo, derr := buildBundleDir(b.ID, b.Slug)
		if derr != nil {
			slog.Error("Sweep: buildBundleDir failed.", "bundleID", id, "err", derr)
			continue
		}

		files, _, err := s.templateStore.ListFiles(
			dirstore.ListingConfig{
				FilterPartitions: []string{dirInfo.DirName},
				PageSize:         1,
			}, "",
		)
		if err != nil || len(files) != 0 {
			slog.Warn("Sweep: bundle still contains templates, skipping.", "bundleID", id)
			continue
		}

		delete(all.Bundles, id)
		changed = true

		_ = os.RemoveAll(filepath.Join(s.baseDir, dirInfo.DirName))
		slog.Info("Hard-deleted bundle.", "bundleID", id)
	}
	if changed {
		if err := s.writeAllBundles(all); err != nil {
			slog.Error("Sweep: writeAllBundles failed.", "err", err)
		}
	}
}

// PutPromptBundle creates or replaces a prompt bundle.
func (s *PromptTemplateStore) PutPromptBundle(
	ctx context.Context, req *spec.PutPromptBundleRequest,
) (*spec.PutPromptBundleResponse, error) {
	if req == nil || req.Body == nil ||
		req.BundleID == "" || req.Body.Slug == "" || req.Body.DisplayName == "" {
		return nil, fmt.Errorf("%w: id, slug & displayName are required", ErrInvalidRequest)
	}
	if err := req.Body.Slug.Validate(); err != nil {
		return nil, err
	}

	// Coordinate with sweep operations.
	s.sweepMu.Lock()
	defer s.sweepMu.Unlock()

	all, err := s.readAllBundles(false)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	createdAt := now

	if existing, ok := all.Bundles[req.BundleID]; ok {
		if !existing.CreatedAt.IsZero() {
			createdAt = existing.CreatedAt
		}
	}

	b := spec.PromptBundle{
		ID:            req.BundleID,
		Slug:          req.Body.Slug,
		DisplayName:   req.Body.DisplayName,
		Description:   req.Body.Description,
		IsEnabled:     req.Body.IsEnabled,
		CreatedAt:     createdAt,
		ModifiedAt:    now,
		IsBuiltIn:     false,
		SoftDeletedAt: nil,
	}
	all.Bundles[req.BundleID] = b
	if err := s.writeAllBundles(all); err != nil {
		return nil, err
	}
	slog.Info("PutPromptBundle.", "id", req.BundleID)
	return &spec.PutPromptBundleResponse{}, nil
}

// PatchPromptBundle toggles the enabled flag of a bundle.
func (s *PromptTemplateStore) PatchPromptBundle(
	ctx context.Context, req *spec.PatchPromptBundleRequest,
) (*spec.PatchPromptBundleResponse, error) {
	if req == nil || req.Body == nil || req.BundleID == "" {
		return nil, fmt.Errorf("%w: bundleID required", ErrInvalidRequest)
	}

	// Coordinate with sweep operations.
	s.sweepMu.Lock()
	defer s.sweepMu.Unlock()

	all, err := s.readAllBundles(false)
	if err != nil {
		return nil, err
	}
	bundle, ok := all.Bundles[req.BundleID]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrBundleNotFound, req.BundleID)
	}
	bundle.IsEnabled = req.Body.IsEnabled
	bundle.ModifiedAt = time.Now().UTC()
	all.Bundles[req.BundleID] = bundle

	if err := s.writeAllBundles(all); err != nil {
		return nil, err
	}
	slog.Info("PatchPromptBundle.", "id", req.BundleID, "enabled", req.Body.IsEnabled)
	return &spec.PatchPromptBundleResponse{}, nil
}

// DeletePromptBundle soft-deletes a bundle if it is empty.
func (s *PromptTemplateStore) DeletePromptBundle(
	ctx context.Context, req *spec.DeletePromptBundleRequest,
) (*spec.DeletePromptBundleResponse, error) {
	if req == nil || req.BundleID == "" {
		return nil, fmt.Errorf("%w: bundleID required", ErrInvalidRequest)
	}

	// Coordinate with sweep operations.
	s.sweepMu.Lock()
	defer s.sweepMu.Unlock()

	all, err := s.readAllBundles(false)
	if err != nil {
		return nil, err
	}
	bundle, ok := all.Bundles[req.BundleID]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrBundleNotFound, req.BundleID)
	}
	if isSoftDeleted(bundle) {
		return nil, fmt.Errorf("%w: %s", ErrBundleDeleting, req.BundleID)
	}

	dirInfo, derr := buildBundleDir(bundle.ID, bundle.Slug)
	if derr != nil {
		return nil, derr
	}

	files, _, err := s.templateStore.ListFiles(
		dirstore.ListingConfig{
			FilterPartitions: []string{dirInfo.DirName},
			PageSize:         1,
		}, "",
	)
	if err != nil {
		return nil, err
	}
	if len(files) != 0 {
		return nil, fmt.Errorf("%w: %s", ErrBundleNotEmpty, req.BundleID)
	}

	now := time.Now().UTC()
	bundle.IsEnabled = false
	bundle.SoftDeletedAt = &now
	all.Bundles[req.BundleID] = bundle

	if err := s.writeAllBundles(all); err != nil {
		return nil, err
	}
	slog.Info("DeletePromptBundle request.", "id", req.BundleID)
	s.kickCleanupLoop()
	return &spec.DeletePromptBundleResponse{}, nil
}

// ListPromptBundles lists prompt bundles with optional filtering and pagination.
func (s *PromptTemplateStore) ListPromptBundles(
	ctx context.Context,
	req *spec.ListPromptBundlesRequest,
) (*spec.ListPromptBundlesResponse, error) {
	all, err := s.readAllBundles(false)
	if err != nil {
		return nil, err
	}

	bundles := make([]spec.PromptBundle, 0, len(all.Bundles))
	for _, b := range all.Bundles {
		if isSoftDeleted(b) {
			continue
		}
		bundles = append(bundles, b)
	}

	wantIDs := map[spec.BundleID]struct{}{}
	if req != nil && len(req.BundleIDs) > 0 {
		for _, id := range req.BundleIDs {
			wantIDs[id] = struct{}{}
		}
	}
	includeDisabled := req != nil && req.IncludeDisabled
	filtered := make([]spec.PromptBundle, 0, len(bundles))
	for _, b := range bundles {
		if len(wantIDs) > 0 {
			if _, ok := wantIDs[b.ID]; !ok {
				continue
			}
		}
		if !includeDisabled && !b.IsEnabled {
			continue
		}
		filtered = append(filtered, b)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ModifiedAt.After(filtered[j].ModifiedAt)
	})

	pageSize := defPageSize
	if req != nil && req.PageSize > 0 && req.PageSize <= maxPageSize {
		pageSize = req.PageSize
	}

	start := 0
	if req != nil && req.PageToken != "" {
		if idx, err := strconv.Atoi(req.PageToken); err == nil && idx >= 0 && idx < len(filtered) {
			start = idx
		}
	}

	end := min(start+pageSize, len(filtered))
	nextTok := ""
	if end < len(filtered) {
		nextTok = strconv.Itoa(end)
	}

	return &spec.ListPromptBundlesResponse{
		Body: &spec.ListPromptBundlesResponseBody{
			PromptBundles: filtered[start:end],
			NextPageToken: nullableStr(nextTok),
		},
	}, nil
}

// findTemplate locates a template file by bundle, slug, and version.
// If wantVersion is empty, returns the active version (enabled, most recent ModifiedAt).
// Note: There is a potential race condition where the active version could be modified
// between selection and reading, but this is acceptable for the current use case.
func (s *PromptTemplateStore) findTemplate(
	bdi bundleDirInfo,
	slug spec.TemplateSlug,
	wantVersion spec.TemplateVersion,
) (fi fileInfo, fullPath string, err error) {
	list, _, err := s.templateStore.ListFiles(
		dirstore.ListingConfig{
			FilterPartitions: []string{bdi.DirName},
			PageSize:         10000,
		}, "")
	if err != nil {
		return fi, fullPath, err
	}

	var newestTime time.Time
	for _, p := range list {
		local := filepath.Base(p)
		finf, perr := parseTemplateFileName(local)
		if perr != nil {
			continue
		}

		if slug != "" && finf.Slug != string(slug) {
			continue
		}
		if wantVersion != "" && finf.Version != string(wantVersion) {
			continue
		}
		// Read minimal JSON to confirm slug matches.
		raw, gerr := s.templateStore.GetFileData(dirstore.FileKey{
			FileName: local, XAttr: bundlePartitionAttr(bdi.DirName),
		}, false)
		if gerr != nil {
			continue
		}
		if sVal, _ := raw["slug"].(string); sVal != finf.Slug {
			continue
		}

		if wantVersion != "" { // Exact match.
			return finf, p, nil
		}
		// Active version selection.
		enabled, _ := raw["isEnabled"].(bool)
		if !enabled {
			continue
		}

		var modAt time.Time
		if ts, ok := raw["modifiedAt"].(string); ok {
			modAt, _ = time.Parse(time.RFC3339Nano, ts)
		}
		if modAt.IsZero() {
			if st, _ := os.Stat(filepath.Join(s.baseDir, p)); st != nil {
				modAt = st.ModTime()
			}
		}
		if fi.FileName == "" || modAt.After(newestTime) {
			newestTime = modAt
			fi, fullPath = finf, p
		}
	}
	if fullPath == "" {
		err = fmt.Errorf("%w: %s", ErrTemplateNotFound, slug)
	}
	return fi, fullPath, err
}

// PutPromptTemplate creates a new prompt template version.
// Returns an error if the (slug, version) already exists.
func (s *PromptTemplateStore) PutPromptTemplate(
	ctx context.Context, req *spec.PutPromptTemplateRequest,
) (*spec.PutPromptTemplateResponse, error) {
	if req == nil || req.Body == nil {
		return nil, fmt.Errorf("%w: nil request/body", ErrInvalidRequest)
	}
	if req.BundleID == "" || req.TemplateSlug == "" || req.Body.Version == "" {
		return nil, fmt.Errorf("%w: bundleID, templateSlug, version required", ErrInvalidRequest)
	}
	if err := req.TemplateSlug.Validate(); err != nil {
		return nil, err
	}
	if err := req.Body.Version.Validate(); err != nil {
		return nil, err
	}
	if len(req.Body.Blocks) == 0 || req.Body.DisplayName == "" {
		return nil, fmt.Errorf("%w: displayName & ≥1 block required", ErrInvalidRequest)
	}

	bundle, err := s.getBundle(req.BundleID)
	if err != nil {
		return nil, err
	}
	if !bundle.IsEnabled {
		return nil, fmt.Errorf("%w: %s", ErrBundleDisabled, req.BundleID)
	}
	dirInfo, derr := buildBundleDir(bundle.ID, bundle.Slug)
	if derr != nil {
		return nil, derr
	}

	// Per-slug lock.
	lk := s.slugLock.lockKey(bundle.ID, req.TemplateSlug)
	lk.Lock()
	defer lk.Unlock()

	targetFN, ferr := templateFileName(req.TemplateSlug, req.Body.Version)
	if ferr != nil {
		return nil, ferr
	}
	existsList, _, _ := s.templateStore.ListFiles(
		dirstore.ListingConfig{
			FilterPartitions: []string{dirInfo.DirName},
			FilenamePrefix:   targetFN, PageSize: 10,
		}, "")

	// Check for exact filename match.
	for _, existing := range existsList {
		if filepath.Base(existing) == targetFN {
			return nil, fmt.Errorf("%w: slug+version already exists", ErrConflict)
		}
	}

	// Build object.
	now := time.Now().UTC()
	u, err := uuidv7filename.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("uuid not available. err: %w", err)
	}

	tpl := spec.PromptTemplate{
		ID:            spec.TemplateID(u),
		DisplayName:   req.Body.DisplayName,
		Slug:          req.TemplateSlug,
		Description:   req.Body.Description,
		IsEnabled:     req.Body.IsEnabled,
		Tags:          req.Body.Tags,
		Blocks:        req.Body.Blocks,
		Variables:     req.Body.Variables,
		PreProcessors: req.Body.PreProcessors,
		Version:       req.Body.Version,
		CreatedAt:     now,
		ModifiedAt:    now,
		IsBuiltIn:     false,
	}

	// Validate template structure.
	if err := validateTemplate(&tpl); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	mp, _ := encdec.StructWithJSONTagsToMap(tpl)

	if err := s.templateStore.SetFileData(dirstore.FileKey{
		FileName: targetFN,
		XAttr:    bundlePartitionAttr(dirInfo.DirName),
	}, mp); err != nil {
		return nil, err
	}
	slog.Debug(
		"PutPromptTemplate.",
		"bundleID",
		req.BundleID,
		"slug",
		req.TemplateSlug,
		"version",
		req.Body.Version,
	)
	return &spec.PutPromptTemplateResponse{}, nil
}

// DeletePromptTemplate hard-deletes a template version.
func (s *PromptTemplateStore) DeletePromptTemplate(
	ctx context.Context, req *spec.DeletePromptTemplateRequest,
) (*spec.DeletePromptTemplateResponse, error) {
	if req == nil || req.TemplateSlug == "" || req.BundleID == "" || req.Version == "" {
		return nil, fmt.Errorf("%w: bundleID, templateSlug, version required", ErrInvalidRequest)
	}
	if err := req.TemplateSlug.Validate(); err != nil {
		return nil, err
	}
	if err := req.Version.Validate(); err != nil {
		return nil, err
	}
	bundle, err := s.getBundle(req.BundleID)
	if err != nil {
		return nil, err
	}
	dirInfo, derr := buildBundleDir(bundle.ID, bundle.Slug)
	if derr != nil {
		return nil, derr
	}

	// Per-slug lock.
	lk := s.slugLock.lockKey(bundle.ID, req.TemplateSlug)
	lk.Lock()
	defer lk.Unlock()

	targetFN, ferr := templateFileName(req.TemplateSlug, req.Version)
	if ferr != nil {
		return nil, ferr
	}
	if err := s.templateStore.DeleteFile(dirstore.FileKey{
		FileName: targetFN, XAttr: bundlePartitionAttr(dirInfo.DirName),
	}); err != nil {
		return nil, err
	}
	slog.Info(
		"DeletePromptTemplate.",
		"bundleID",
		req.BundleID,
		"slug",
		req.TemplateSlug,
		"version",
		req.Version,
	)
	return &spec.DeletePromptTemplateResponse{}, nil
}

// PatchPromptTemplate enables or disables a template version.
func (s *PromptTemplateStore) PatchPromptTemplate(
	ctx context.Context, req *spec.PatchPromptTemplateRequest,
) (*spec.PatchPromptTemplateResponse, error) {
	if req == nil || req.Body == nil ||
		req.TemplateSlug == "" || req.BundleID == "" || req.Body.Version == "" {
		return nil, fmt.Errorf("%w: bundleID, templateSlug, version required", ErrInvalidRequest)
	}
	if err := req.TemplateSlug.Validate(); err != nil {
		return nil, err
	}
	if err := req.Body.Version.Validate(); err != nil {
		return nil, err
	}
	bundle, err := s.getBundle(req.BundleID)
	if err != nil {
		return nil, err
	}
	if !bundle.IsEnabled {
		return nil, fmt.Errorf("%w: %s", ErrBundleDisabled, req.BundleID)
	}
	dirInfo, derr := buildBundleDir(bundle.ID, bundle.Slug)
	if derr != nil {
		return nil, derr
	}

	// Per-slug lock.
	lk := s.slugLock.lockKey(bundle.ID, req.TemplateSlug)
	lk.Lock()
	defer lk.Unlock()

	targetFN, ferr := templateFileName(req.TemplateSlug, req.Body.Version)
	if ferr != nil {
		return nil, ferr
	}
	key := dirstore.FileKey{
		FileName: targetFN,
		XAttr:    bundlePartitionAttr(dirInfo.DirName),
	}

	raw, err := s.templateStore.GetFileData(key, false)
	if err != nil {
		return nil, err
	}
	var tpl spec.PromptTemplate
	if err := encdec.MapToStructWithJSONTags(raw, &tpl); err != nil {
		return nil, err
	}
	tpl.IsEnabled = req.Body.IsEnabled
	// Do not update ModifiedAt for enable/disable as per spec.

	mp, _ := encdec.StructWithJSONTagsToMap(tpl)
	if err := s.templateStore.SetFileData(key, mp); err != nil {
		return nil, err
	}
	slog.Info(
		"PatchPromptTemplate.",
		"bundleID",
		req.BundleID,
		"slug",
		req.TemplateSlug,
		"version",
		req.Body.Version,
		"enabled",
		req.Body.IsEnabled,
	)
	return &spec.PatchPromptTemplateResponse{}, nil
}

// GetPromptTemplate returns a template version or the active version if version is omitted.
func (s *PromptTemplateStore) GetPromptTemplate(
	ctx context.Context, req *spec.GetPromptTemplateRequest,
) (*spec.GetPromptTemplateResponse, error) {
	if req == nil || req.TemplateSlug == "" || req.BundleID == "" {
		return nil, fmt.Errorf("%w: bundleID & templateSlug required", ErrInvalidRequest)
	}
	if err := req.TemplateSlug.Validate(); err != nil {
		return nil, err
	}

	bundle, err := s.getBundle(req.BundleID)
	if err != nil {
		return nil, err
	}

	dirInfo, derr := buildBundleDir(bundle.ID, bundle.Slug)
	if derr != nil {
		return nil, derr
	}

	// Per-slug lock.
	lk := s.slugLock.lockKey(bundle.ID, req.TemplateSlug)
	lk.RLock()
	defer lk.RUnlock()

	var finf fileInfo

	if req.Version != "" {
		finf, _, err = s.findTemplate(dirInfo, req.TemplateSlug, req.Version)
	} else {
		finf, _, err = s.findTemplate(dirInfo, req.TemplateSlug, "")
	}
	if err != nil {
		return nil, err
	}
	raw, err := s.templateStore.GetFileData(dirstore.FileKey{
		FileName: finf.FileName,
		XAttr:    bundlePartitionAttr(dirInfo.DirName),
	}, false)
	if err != nil {
		return nil, err
	}
	var tpl spec.PromptTemplate
	if err := encdec.MapToStructWithJSONTags(raw, &tpl); err != nil {
		return nil, err
	}
	return &spec.GetPromptTemplateResponse{Body: &tpl}, nil
}

// ListPromptTemplates lists templates with optional filtering and pagination.
func (s *PromptTemplateStore) ListPromptTemplates(
	ctx context.Context, req *spec.ListPromptTemplatesRequest,
) (*spec.ListPromptTemplatesResponse, error) {
	pageSize, pageToken := defPageSize, ""
	includeDisabled, allVersions := false, false
	tagFilter := map[string]struct{}{}
	bundleFilter := map[spec.BundleID]struct{}{}

	if req != nil {
		if req.PageSize > 0 && req.PageSize <= maxPageSize {
			pageSize = req.PageSize
		}
		pageToken = req.PageToken
		includeDisabled = req.IncludeDisabled
		allVersions = req.AllVersions
		for _, t := range req.Tags {
			tagFilter[t] = struct{}{}
		}
		for _, id := range req.BundleIDs {
			bundleFilter[id] = struct{}{}
		}
	}

	// Use improved pagination with larger initial fetch for better performance.
	fetchSize := pageSize
	if !allVersions {
		// Need head-room because we might drop many versions
		// while picking only the latest enabled one per slug.
		fetchSize = pageSize * 2
	}
	files, next, err := s.templateStore.ListFiles(
		dirstore.ListingConfig{
			PageSize:  fetchSize,
			SortOrder: dirstore.SortOrderDescending,
		},
		pageToken)
	if err != nil {
		return nil, err
	}

	type rec struct {
		item spec.PromptTemplateListItem
		mt   time.Time
	}
	latest := make(map[string]rec)

	var items []spec.PromptTemplateListItem
	for _, p := range files {
		fn := filepath.Base(p)
		finf, err := parseTemplateFileName(fn)
		if err != nil {
			continue
		}

		dir := filepath.Base(filepath.Dir(p))
		bdi, err := parseBundleDir(dir)
		if err != nil {
			continue
		}

		bid := spec.BundleID(bdi.ID)
		bslug := spec.BundleSlug(bdi.Slug)
		tslug := spec.TemplateSlug(finf.Slug)
		tver := spec.TemplateVersion(finf.Version)

		if len(bundleFilter) > 0 {
			if _, ok := bundleFilter[bid]; !ok {
				continue
			}
		}
		if _, err := s.getBundle(bid); err != nil {
			continue
		}

		raw, err := s.templateStore.GetFileData(dirstore.FileKey{
			FileName: fn, XAttr: bundlePartitionAttr(dir),
		}, false)
		if err != nil {
			continue
		}
		var pt spec.PromptTemplate
		if err := encdec.MapToStructWithJSONTags(raw, &pt); err != nil {
			continue
		}

		if !pt.IsEnabled && !includeDisabled {
			continue
		}
		if len(tagFilter) != 0 {
			found := false
			for _, t := range pt.Tags {
				if _, ok := tagFilter[t]; ok {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		it := spec.PromptTemplateListItem{
			BundleID:        bid,
			BundleSlug:      bslug,
			TemplateSlug:    tslug,
			TemplateVersion: tver,
			IsBuiltIn:       false,
		}
		if allVersions {
			items = append(items, it)
			continue
		}
		k := string(bid) + "|" + string(tslug)
		if prev, ok := latest[k]; ok {
			if pt.ModifiedAt.After(prev.mt) {
				latest[k] = rec{item: it, mt: pt.ModifiedAt}
			}
		} else {
			latest[k] = rec{item: it, mt: pt.ModifiedAt}
		}
	}

	if !allVersions {
		for _, r := range latest {
			items = append(items, r.item)
		}
		// Sort by modification time for consistent ordering.
		sort.Slice(items, func(i, j int) bool {
			return latest[string(items[i].BundleID)+"|"+string(items[i].TemplateSlug)].mt.After(
				latest[string(items[j].BundleID)+"|"+string(items[j].TemplateSlug)].mt)
		})
	}

	if !allVersions && len(items) > pageSize {
		items = items[:pageSize]
	}

	return &spec.ListPromptTemplatesResponse{
		Body: &spec.ListPromptTemplatesResponseBody{
			PromptTemplateListItems: items,
			NextPageToken:           nullableStr(next),
		},
	}, nil
}

// SearchPromptTemplates performs a full-text search for templates using the FTS engine.
func (s *PromptTemplateStore) SearchPromptTemplates(
	ctx context.Context, req *spec.SearchPromptTemplatesRequest,
) (*spec.SearchPromptTemplatesResponse, error) {
	if req == nil || req.Query == "" {
		return nil, fmt.Errorf("%w: query required", ErrInvalidRequest)
	}
	if s.fts == nil {
		return nil, ErrFTSDisabled
	}

	pageSize := defPageSize
	if req.PageSize > 0 && req.PageSize <= maxPageSize {
		pageSize = req.PageSize
	}

	// Fetch more results to account for filtering.
	searchPageSize := pageSize * 2
	hits, next, err := s.fts.Search(ctx, req.Query, req.PageToken, searchPageSize)
	if err != nil {
		return nil, err
	}

	items := make([]spec.PromptTemplateListItem, 0, len(hits))
	for _, h := range hits {
		dir := filepath.Base(filepath.Dir(h.ID))
		bdi, err := parseBundleDir(dir)
		if err != nil {
			continue
		}
		bid := spec.BundleID(bdi.ID)
		bslug := spec.BundleSlug(bdi.Slug)

		if _, err := s.getBundle(bid); err != nil {
			continue
		}

		fn := filepath.Base(h.ID)
		finf, err := parseTemplateFileName(fn)
		if err != nil {
			continue
		}
		tslug := spec.TemplateSlug(finf.Slug)
		tver := spec.TemplateVersion(finf.Version)

		raw, err := s.templateStore.GetFileData(dirstore.FileKey{
			FileName: fn, XAttr: bundlePartitionAttr(dir),
		}, false)
		if err != nil {
			continue
		}
		var pt spec.PromptTemplate
		if err := encdec.MapToStructWithJSONTags(raw, &pt); err != nil {
			continue
		}

		if !req.IncludeDisabled && !pt.IsEnabled {
			continue
		}

		items = append(items, spec.PromptTemplateListItem{
			BundleID:        bid,
			BundleSlug:      bslug,
			TemplateSlug:    tslug,
			TemplateVersion: tver,
			IsBuiltIn:       false,
		})

		// Stop when we have enough results.
		if len(items) >= pageSize {
			break
		}
	}

	// Adjust next page token if we have fewer results than requested.
	if len(items) < pageSize {
		next = ""
	}

	return &spec.SearchPromptTemplatesResponse{
		Body: &spec.SearchPromptTemplatesResponseBody{
			PromptTemplateListItems: items,
			NextPageToken:           nullableStr(next),
		},
	}, nil
}
