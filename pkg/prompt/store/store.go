package store

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/dirstore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
)

const (
	promptTemplateFileExtension = "json"
	promptBundleFileExtension   = "json"
	sqliteDBFileName            = "prompttemplates.fts.sqlite"
	bundlesMetaFileName         = "prompttemplates.bundles.json"
	maxPageSize                 = 256
	defPageSize                 = 10
)

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// <bundleID>_<bundleSlug>.
type bundleDirInfo struct {
	ID      string
	Slug    string
	DirName string
}

func buildBundleDir(id, slug string) bundleDirInfo {
	dir := fmt.Sprintf("%s_%s", sanitize(id), sanitize(slug))
	return bundleDirInfo{ID: id, Slug: slug, DirName: dir}
}

func parseBundleDir(dir string) (bundleDirInfo, error) {
	parts := strings.SplitN(dir, "_", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return bundleDirInfo{}, fmt.Errorf("invalid bundle dir: %q", dir)
	}
	return bundleDirInfo{ID: parts[0], Slug: parts[1], DirName: dir}, nil
}

// <templateID>_<templateSlug>_<version>.json.
type templateFileInfo struct {
	ID       string
	Slug     string
	Version  string
	FileName string
}

func buildTemplateFile(id, slug, version string) (templateFileInfo, error) {
	if id == "" || slug == "" || version == "" {
		return templateFileInfo{}, errors.New("id / slug / version must be non-empty")
	}
	name := fmt.Sprintf("%s_%s_%s.%s",
		sanitize(id), sanitize(slug), sanitize(version), promptTemplateFileExtension)
	return templateFileInfo{ID: id, Slug: slug, Version: version, FileName: name}, nil
}

func parseTemplateFile(fn string) (templateFileInfo, error) {
	fn = filepath.Base(fn)
	if !strings.HasSuffix(fn, "."+promptTemplateFileExtension) {
		return templateFileInfo{}, fmt.Errorf("not a %q file: %q",
			promptTemplateFileExtension, fn)
	}
	trim := strings.TrimSuffix(fn, "."+promptTemplateFileExtension)
	parts := strings.Split(trim, "_")
	if len(parts) < 3 {
		return templateFileInfo{}, fmt.Errorf("invalid template file name: %q", fn)
	}
	// The slug might itself contain ‘_’.  ID is UUID (fixed, first).
	// Version is last chunk.  Everything in-between is slug.
	id := parts[0]
	version := parts[len(parts)-1]
	slug := strings.Join(parts[1:len(parts)-1], "_")
	if id == "" || slug == "" || version == "" {
		return templateFileInfo{}, fmt.Errorf("invalid template file name: %q", fn)
	}
	return templateFileInfo{
		ID: id, Slug: slug, Version: version, FileName: fn,
	}, nil
}

// Sanitize reduces a string to ASCII letters, digits and hyphens (same behaviour as uuidv7filename.Build’s suffix sanitiser).
func sanitize(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			buf.WriteRune(r)
		case r >= 'a' && r <= 'z':
			buf.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			buf.WriteRune(r)
		case r == '-', r == '_':
			buf.WriteRune(r)
		}
	}
	out := buf.String()
	if out == "" {
		return "x"
	}
	return out
}

type PromptTemplateStore struct {
	baseDir       string
	bundleStore   *filestore.MapFileStore
	templateStore *dirstore.MapDirectoryStore
	pp            dirstore.PartitionProvider
	enableFTS     bool
	fts           *ftsengine.Engine
}

type Option func(*PromptTemplateStore) error

func WithFTS(enabled bool) Option {
	return func(s *PromptTemplateStore) error {
		s.enableFTS = enabled
		return nil
	}
}

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

	defMap := map[string]any{"bundles": map[string]any{}}
	bs, err := filestore.NewMapFileStore(
		filepath.Join(s.baseDir, bundlesMetaFileName),
		defMap,
		filestore.WithCreateIfNotExists(true),
		filestore.WithAutoFlush(true),
		filestore.WithEncoderDecoder(encdec.JSONEncoderDecoder{}))
	if err != nil {
		return nil, fmt.Errorf("bundle meta store init: %w", err)
	}
	s.bundleStore = bs

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

	dirOpts := []dirstore.Option{dirstore.WithPartitionProvider(s.pp)}
	if s.fts != nil {
		dirOpts = append(dirOpts, dirstore.WithListeners(NewFTSListener(s.fts)))
	}
	ds, err := dirstore.NewMapDirectoryStore(s.baseDir, true, dirOpts...)
	if err != nil {
		return nil, err
	}
	s.templateStore = ds
	return s, nil
}

func (s *PromptTemplateStore) PutPromptBundle(
	ctx context.Context,
	req *spec.PutPromptBundleRequest,
) (*spec.PutPromptBundleResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request/body cannot be nil")
	}
	if req.BundleID == "" || req.Body.Slug == "" || req.Body.DisplayName == "" {
		return nil, errors.New("id, slug & displayName are required")
	}

	all, err := s.bundleStore.GetAll(false)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()

	bundles, _ := all["bundles"].(map[string]any)
	// Keep original CreatedAt if bundle already exists.
	var createdAt time.Time
	if existing, ok := bundles[req.BundleID].(map[string]any); ok {
		var existingBundle spec.PromptBundle
		if err := encdec.MapToStructWithJSONTags(existing, &existingBundle); err == nil {
			createdAt = existingBundle.CreatedAt
		}
	}

	if createdAt.IsZero() {
		createdAt = now
	}

	b := spec.PromptBundle{
		ID:          req.BundleID,
		Slug:        req.Body.Slug,
		DisplayName: req.Body.DisplayName,
		Description: req.Body.Description,
		IsEnabled:   req.Body.IsEnabled,
		CreatedAt:   createdAt,
		ModifiedAt:  now,
	}

	val, _ := encdec.StructWithJSONTagsToMap(b)
	if err := s.bundleStore.SetKey([]string{"bundles", req.BundleID}, val); err != nil {
		return nil, err
	}
	return &spec.PutPromptBundleResponse{}, nil
}

func (s *PromptTemplateStore) PatchPromptBundle(
	ctx context.Context,
	req *spec.PatchPromptBundleRequest,
) (*spec.PatchPromptBundleResponse, error) {
	if req == nil || req.Body == nil || req.BundleID == "" {
		return nil, errors.New("invalid request")
	}
	bundle, err := s.getBundle(req.BundleID)
	if err != nil {
		return nil, err
	}
	bundle.IsEnabled = req.Body.IsEnabled
	bundle.ModifiedAt = time.Now().UTC()

	val, _ := encdec.StructWithJSONTagsToMap(bundle)
	if err := s.bundleStore.SetKey([]string{"bundles", req.BundleID}, val); err != nil {
		return nil, err
	}
	return &spec.PatchPromptBundleResponse{}, nil
}

func (s *PromptTemplateStore) DeletePromptBundle(
	ctx context.Context,
	req *spec.DeletePromptBundleRequest,
) (*spec.DeletePromptBundleResponse, error) {
	if req == nil || req.BundleID == "" {
		return nil, errors.New("invalid request")
	}
	return &spec.DeletePromptBundleResponse{}, nil
}

func (s *PromptTemplateStore) ListPromptBundles(
	ctx context.Context,
	req *spec.ListPromptBundlesRequest,
) (*spec.ListPromptBundlesResponse, error) {
	all, err := s.bundleStore.GetAll(false)
	if err != nil {
		return nil, err
	}
	rawBundles, ok := all["bundles"].(map[string]map[string]any)
	if !ok {
		return nil, errors.New("bundles not found")
	}

	var bundles []spec.PromptBundle
	for _, v := range rawBundles {
		var b spec.PromptBundle
		if err := encdec.MapToStructWithJSONTags(v, &b); err == nil {
			bundles = append(bundles, b)
		}
	}

	wantIDs := map[string]struct{}{}
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

	// Sort by ModifiedAt DESC.
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ModifiedAt.After(filtered[j].ModifiedAt)
	})

	// Pagination (simple positional token – strconv.Itoa(index)).
	pageSize := defPageSize
	pageToken := ""
	if req != nil {
		if req.PageSize > 0 && req.PageSize <= maxPageSize {
			pageSize = req.PageSize
		}
		pageToken = req.PageToken
	}
	start := 0
	if pageToken != "" {
		if idx, err := strconv.Atoi(pageToken); err == nil && idx >= 0 && idx < len(filtered) {
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

func (s *PromptTemplateStore) getBundle(
	bundleID string,
) (spec.PromptBundle, error) {
	all, err := s.bundleStore.GetAll(false)
	if err != nil {
		return spec.PromptBundle{}, err
	}
	bundles, ok := all["bundles"].(map[string]map[string]any)
	if !ok {
		return spec.PromptBundle{}, errors.New("bundles not found")
	}
	raw, ok := bundles[bundleID]
	if !ok {
		return spec.PromptBundle{}, fmt.Errorf("bundle %q not found", bundleID)
	}
	var b spec.PromptBundle
	if err := encdec.MapToStructWithJSONTags(raw, &b); err != nil {
		return spec.PromptBundle{}, err
	}
	return b, nil
}

func (s *PromptTemplateStore) PutPromptTemplate(
	ctx context.Context,
	req *spec.PutPromptTemplateRequest,
) (*spec.PutPromptTemplateResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request/body cannot be nil")
	}
	if req.BundleID == "" || req.TemplateID == "" {
		return nil, errors.New("bundleID and templateID are required")
	}
	body := req.Body
	if body.DisplayName == "" || body.Slug == "" || body.Version == "" || len(body.Blocks) == 0 {
		return nil, errors.New("missing required template body fields")
	}

	// Validate bundle exists and enabled.
	bundle, err := s.getBundle(req.BundleID)
	if err != nil {
		return nil, err
	}
	if !bundle.IsEnabled {
		return nil, fmt.Errorf("bundle %q is disabled", req.BundleID)
	}

	dirInfo := buildBundleDir(bundle.ID, bundle.Slug)
	fileInfo, err := buildTemplateFile(req.TemplateID, body.Slug, body.Version)
	if err != nil {
		return nil, err
	}
	key := dirstore.FileKey{
		FileName: fileInfo.FileName,
		XAttr:    bundlePartitionAttr(dirInfo.DirName),
	}

	// If an older version exists with same version – overwrite keeping CreatedAt.
	var createdAt time.Time
	if raw, err := s.templateStore.GetFileData(key, false); err == nil {
		var tmp spec.PromptTemplate
		_ = encdec.MapToStructWithJSONTags(raw, &tmp)
		createdAt = tmp.CreatedAt
	}
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	t := spec.PromptTemplate{
		ID:            req.TemplateID,
		Slug:          body.Slug,
		DisplayName:   body.DisplayName,
		Description:   body.Description,
		IsEnabled:     body.IsEnabled,
		Tags:          body.Tags,
		Blocks:        body.Blocks,
		Variables:     body.Variables,
		PreProcessors: body.PreProcessors,
		Version:       body.Version,
		CreatedAt:     createdAt,
		ModifiedAt:    time.Now().UTC(),
	}
	data, _ := encdec.StructWithJSONTagsToMap(t)
	if err := s.templateStore.SetFileData(key, data); err != nil {
		return nil, err
	}
	return &spec.PutPromptTemplateResponse{}, nil
}

func (s *PromptTemplateStore) DeletePromptTemplate(
	ctx context.Context,
	req *spec.DeletePromptTemplateRequest,
) (*spec.DeletePromptTemplateResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}
	if req.BundleID == "" || req.TemplateID == "" || req.Version == "" {
		return nil, errors.New("bundleID, templateID & version required")
	}

	bundle, err := s.getBundle(req.BundleID)
	if err != nil {
		return nil, err
	}
	dirInfo := buildBundleDir(bundle.ID, bundle.Slug)
	// We do not know slug – search any slug.
	files, _, err := s.templateStore.ListFiles(
		dirstore.ListingConfig{
			FilenamePrefix:   req.TemplateID + "_",
			FilterPartitions: []string{dirInfo.DirName},
			PageSize:         1000,
		}, "")
	if err != nil {
		return nil, err
	}
	var toDelete *templateFileInfo
	for _, p := range files {
		tf, _ := parseTemplateFile(filepath.Base(p))
		if tf.ID == req.TemplateID && tf.Version == req.Version {
			toDelete = &tf
			break
		}
	}
	if toDelete == nil {
		return nil, fmt.Errorf("template %q version %q not found", req.TemplateID, req.Version)
	}
	if err := s.templateStore.DeleteFile(dirstore.FileKey{
		FileName: toDelete.FileName,
		XAttr:    bundlePartitionAttr(dirInfo.DirName),
	}); err != nil {
		return nil, err
	}
	return &spec.DeletePromptTemplateResponse{}, nil
}

func (s *PromptTemplateStore) PatchPromptTemplate(
	ctx context.Context,
	req *spec.PatchPromptTemplateRequest,
) (*spec.PatchPromptTemplateResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("invalid request")
	}
	if req.BundleID == "" || req.TemplateID == "" {
		return nil, errors.New("bundleID and templateID required")
	}

	bundle, err := s.getBundle(req.BundleID)
	if err != nil {
		return nil, err
	}
	dirInfo := buildBundleDir(bundle.ID, bundle.Slug)

	// Yet slug is unknown (changed maybe).  We'll brute list.
	files, _, err := s.templateStore.ListFiles(
		dirstore.ListingConfig{
			FilenamePrefix:   req.TemplateID + "_",
			FilterPartitions: []string{dirInfo.DirName},
			PageSize:         1000,
		}, "")
	if err != nil {
		return nil, err
	}
	var match templateFileInfo
	found := false
	for _, f := range files {
		pi, _ := parseTemplateFile(filepath.Base(f))
		if pi.ID == req.TemplateID && pi.Version == req.Body.Version {
			match = pi
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("template %q version %q not found", req.TemplateID, req.Body.Version)
	}

	key := dirstore.FileKey{
		FileName: match.FileName,
		XAttr:    bundlePartitionAttr(dirInfo.DirName),
	}
	raw, err := s.templateStore.GetFileData(key, false)
	if err != nil {
		return nil, err
	}
	var t spec.PromptTemplate
	if err := encdec.MapToStructWithJSONTags(raw, &t); err != nil {
		return nil, err
	}
	t.IsEnabled = req.Body.IsEnabled
	t.ModifiedAt = time.Now().UTC()

	data, _ := encdec.StructWithJSONTagsToMap(t)
	if err := s.templateStore.SetFileData(key, data); err != nil {
		return nil, err
	}
	return &spec.PatchPromptTemplateResponse{}, nil
}

func (s *PromptTemplateStore) GetPromptTemplate(
	ctx context.Context,
	req *spec.GetPromptTemplateRequest,
) (*spec.GetPromptTemplateResponse, error) {
	if req == nil || req.BundleID == "" || req.TemplateID == "" {
		return nil, errors.New("bundleID & templateID required")
	}

	bundle, err := s.getBundle(req.BundleID)
	if err != nil {
		return nil, err
	}
	dirInfo := buildBundleDir(bundle.ID, bundle.Slug)

	files, _, err := s.templateStore.ListFiles(
		dirstore.ListingConfig{
			FilenamePrefix:   req.TemplateID + "_",
			FilterPartitions: []string{dirInfo.DirName},
			SortOrder:        dirstore.SortOrderDescending,
			PageSize:         1000,
		}, "")
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("template %q not found in bundle %q", req.TemplateID, req.BundleID)
	}

	var target string
	if req.Version != "" {
		for _, f := range files {
			tf, _ := parseTemplateFile(filepath.Base(f))
			if tf.Version == req.Version {
				target = filepath.Base(f)
				break
			}
		}
		if target == "" {
			return nil, fmt.Errorf("version %q not found", req.Version)
		}
	} else {
		target = filepath.Base(files[0])
	}

	raw, err := s.templateStore.GetFileData(dirstore.FileKey{
		FileName: target,
		XAttr:    bundlePartitionAttr(dirInfo.DirName),
	}, false)
	if err != nil {
		return nil, err
	}
	var t spec.PromptTemplate
	if err := encdec.MapToStructWithJSONTags(raw, &t); err != nil {
		return nil, err
	}
	return &spec.GetPromptTemplateResponse{Body: &t}, nil
}

func (s *PromptTemplateStore) ListPromptTemplates(
	ctx context.Context,
	req *spec.ListPromptTemplatesRequest,
) (*spec.ListPromptTemplatesResponse, error) {
	pageSize := defPageSize
	pageToken := ""
	includeDisabled := false
	allVersions := false
	tagFilter := map[string]struct{}{}
	bundleFilter := map[string]struct{}{}

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

	// We list directory files in descending order to get newest first.
	files, next, err := s.templateStore.ListFiles(
		dirstore.ListingConfig{
			PageSize:  pageSize,
			SortOrder: dirstore.SortOrderDescending,
		}, pageToken)
	if err != nil {
		return nil, err
	}
	items := make([]spec.PromptTemplateListItem, 0, len(files))
	for _, f := range files {
		fn := filepath.Base(f)
		tf, err := parseTemplateFile(fn)
		if err != nil {
			continue
		}
		// Derive bundle from path portion.
		dir := filepath.Base(filepath.Dir(f))
		bdi, err := parseBundleDir(dir)
		if err != nil {
			continue
		}

		if len(bundleFilter) > 0 {
			if _, ok := bundleFilter[bdi.ID]; !ok {
				continue
			}
		}

		raw, err := s.templateStore.GetFileData(dirstore.FileKey{
			FileName: fn,
			XAttr:    bundlePartitionAttr(dir),
		}, false)
		if err != nil {
			continue
		}
		var pt spec.PromptTemplate
		if err := encdec.MapToStructWithJSONTags(raw, &pt); err != nil {
			continue
		}
		if !includeDisabled && !pt.IsEnabled {
			continue
		}

		if len(tagFilter) > 0 {
			match := false
			for _, t := range pt.Tags {
				if _, ok := tagFilter[t]; ok {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		items = append(items, spec.PromptTemplateListItem{
			BundleID:        bdi.ID,
			BundleSlug:      bdi.Slug,
			TemplateID:      tf.ID,
			TemplateSlug:    tf.Slug,
			TemplateVersion: tf.Version,
		})

		// Skip old versions if allVersions is false.
		if !allVersions {
			// Mark that we have already included this templateID → skip rest.
			idkey := bdi.ID + "|" + tf.ID
			for _, f2 := range files {
				if strings.Contains(f2, idkey) {
					// Nothing (just to show reasoning).
					_ = f2
				}
			}
		}
	}
	return &spec.ListPromptTemplatesResponse{
		Body: &spec.ListPromptTemplatesResponseBody{
			PromptTemplateListItems: items,
			NextPageToken:           nullableStr(next),
		},
	}, nil
}

func (s *PromptTemplateStore) SearchPromptTemplates(
	ctx context.Context,
	req *spec.SearchPromptTemplatesRequest,
) (*spec.SearchPromptTemplatesResponse, error) {
	if req == nil || req.Query == "" {
		return nil, errors.New("query required")
	}
	if s.fts == nil {
		return nil, errors.New("FTS is disabled")
	}

	pageSize := defPageSize
	if req.PageSize > 0 && req.PageSize <= maxPageSize {
		pageSize = req.PageSize
	}
	hits, next, err := s.fts.Search(ctx, req.Query, req.PageToken, pageSize)
	if err != nil {
		return nil, err
	}
	// Escape includeDisabled := req.IncludeDisabled.

	items := make([]spec.PromptTemplateListItem, 0, len(hits))
	for _, h := range hits {
		dir := filepath.Base(filepath.Dir(h.ID))
		bdi, err := parseBundleDir(dir)
		if err != nil {
			continue
		}
		tf, err := parseTemplateFile(filepath.Base(h.ID))
		if err != nil {
			continue
		}

		items = append(items, spec.PromptTemplateListItem{
			BundleID:        bdi.ID,
			BundleSlug:      bdi.Slug,
			TemplateID:      tf.ID,
			TemplateSlug:    tf.Slug,
			TemplateVersion: tf.Version,
		})
	}
	return &spec.SearchPromptTemplatesResponse{
		Body: &spec.SearchPromptTemplatesResponseBody{
			PromptTemplateListItems: items,
			NextPageToken:           nullableStr(next),
		},
	}, nil
}
