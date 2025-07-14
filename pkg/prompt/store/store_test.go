package store

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/prompt/nameutils"
	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
)

// newTestStore creates a new PromptTemplateStore for testing and returns a cleanup function.
func newTestStore(t *testing.T) (store *PromptTemplateStore, cleanup func()) {
	t.Helper()
	dir := t.TempDir()
	store, err := NewPromptTemplateStore(dir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	cleanup = func() { store.Close(); os.RemoveAll(dir) }
	return store, cleanup
}

// newTestStoreWithFTS creates a new PromptTemplateStore with FTS enabled.
func newTestStoreWithFTS(t *testing.T) (store *PromptTemplateStore, cleanup func()) {
	t.Helper()
	dir := t.TempDir()
	store, err := NewPromptTemplateStore(dir, WithFTS(true))
	if err != nil {
		t.Fatalf("Failed to create store with FTS: %v", err)
	}
	cleanup = func() { store.Close(); os.RemoveAll(dir) }
	return store, cleanup
}

// mustPutBundle inserts a bundle or fails the test.
func mustPutBundle(
	t *testing.T,
	s *PromptTemplateStore,
	id spec.BundleID,
	slug spec.BundleSlug,
	name string,
	enabled bool,
) {
	t.Helper()
	_, err := s.PutPromptBundle(t.Context(), &spec.PutPromptBundleRequest{
		BundleID: id,
		Body: &spec.PutPromptBundleRequestBody{
			Slug:        slug,
			DisplayName: name,
			Description: "Test bundle: " + name,
			IsEnabled:   enabled,
		},
	})
	if err != nil {
		t.Fatalf("PutPromptBundle failed: %v", err)
	}
}

// mustPutTemplate inserts a template or fails the test.
func mustPutTemplate(
	t *testing.T,
	s *PromptTemplateStore,
	bundleID spec.BundleID, slug spec.TemplateSlug, version spec.TemplateVersion, name string,
	enabled bool,
	tags ...string,
) {
	t.Helper()
	_, err := s.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
		BundleID:     bundleID,
		TemplateSlug: slug,
		Body: &spec.PutPromptTemplateRequestBody{
			DisplayName: name,
			Description: "Test template: " + name,
			IsEnabled:   enabled,
			Blocks:      []spec.MessageBlock{{ID: "1", Role: spec.User, Content: "Hello " + name}},
			Version:     version,
			Tags:        tags,
		},
	})
	if err != nil {
		t.Fatalf("PutPromptTemplate failed: %v", err)
	}
}

// --- A. Bundle Operations ---.

func TestBundleCRUD(t *testing.T) {
	tests := []struct {
		name      string
		bundleID  spec.BundleID
		slug      spec.BundleSlug
		dispName  string
		enabled   bool
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid bundle",
			bundleID:  "b1",
			slug:      "valid-slug",
			dispName:  "Valid Bundle",
			enabled:   true,
			wantError: false,
		},
		{
			name:      "disabled bundle",
			bundleID:  "b2",
			slug:      "disabled-slug",
			dispName:  "Disabled Bundle",
			enabled:   false,
			wantError: false,
		},
		{
			name:      "empty bundle ID",
			bundleID:  "",
			slug:      "slug",
			dispName:  "Name",
			enabled:   true,
			wantError: true,
			errorMsg:  "required",
		},
		{
			name:      "empty slug",
			bundleID:  "b3",
			slug:      "",
			dispName:  "Name",
			enabled:   true,
			wantError: true,
			errorMsg:  "required",
		},
		{
			name:      "empty display name",
			bundleID:  "b4",
			slug:      "slug",
			dispName:  "",
			enabled:   true,
			wantError: true,
			errorMsg:  "required",
		},
		{
			name:      "invalid slug with dot",
			bundleID:  "b5",
			slug:      "bad.slug",
			dispName:  "Name",
			enabled:   true,
			wantError: true,
			errorMsg:  "invalid slug",
		},
		{
			name:      "invalid slug with underscore",
			bundleID:  "b6",
			slug:      "bad_slug",
			dispName:  "Name",
			enabled:   true,
			wantError: true,
			errorMsg:  "invalid slug",
		},
		{
			name:      "invalid slug with space",
			bundleID:  "b7",
			slug:      "bad slug",
			dispName:  "Name",
			enabled:   true,
			wantError: true,
			errorMsg:  "invalid slug",
		},
		{
			name:      "slug too long",
			bundleID:  "b8",
			slug:      spec.BundleSlug(strings.Repeat("a", 65)),
			dispName:  "Name",
			enabled:   true,
			wantError: true,
			errorMsg:  "invalid slug",
		},
		{
			name:      "valid long slug",
			bundleID:  "b9",
			slug:      spec.BundleSlug(strings.Repeat("a", 64)),
			dispName:  "Name",
			enabled:   true,
			wantError: false,
		},
		{
			name:      "valid slug with dash",
			bundleID:  "b10",
			slug:      "valid-slug-with-dash",
			dispName:  "Name",
			enabled:   true,
			wantError: false,
		},
		{
			name:      "valid slug with numbers",
			bundleID:  "b11",
			slug:      "slug123",
			dispName:  "Name",
			enabled:   true,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, cleanup := newTestStore(t)
			defer cleanup()

			_, err := store.PutPromptBundle(t.Context(), &spec.PutPromptBundleRequest{
				BundleID: tt.bundleID,
				Body: &spec.PutPromptBundleRequestBody{
					Slug:        tt.slug,
					DisplayName: tt.dispName,
					IsEnabled:   tt.enabled,
				},
			})

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestBundleUpdate(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	// Create initial bundle.
	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)

	// Update bundle (same ID, different fields).
	_, err := store.PutPromptBundle(t.Context(), &spec.PutPromptBundleRequest{
		BundleID: "b1",
		Body: &spec.PutPromptBundleRequestBody{
			Slug:        "updated-slug",
			DisplayName: "Updated Bundle",
			Description: "Updated description",
			IsEnabled:   false,
		},
	})
	if err != nil {
		t.Fatalf("Bundle update failed: %v", err)
	}

	// Verify update.
	bundle, err := store.getBundle("b1")
	if err != nil {
		t.Fatalf("Failed to get updated bundle: %v", err)
	}
	if bundle.Slug != "updated-slug" || bundle.DisplayName != "Updated Bundle" || bundle.IsEnabled {
		t.Errorf("Bundle not updated correctly: %+v", bundle)
	}
}

func TestBundlePatch(t *testing.T) {
	tests := []struct {
		name      string
		bundleID  spec.BundleID
		initial   bool
		patch     bool
		wantError bool
		errorMsg  string
	}{
		{
			name:      "enable disabled bundle",
			bundleID:  "b1",
			initial:   false,
			patch:     true,
			wantError: false,
		},
		{
			name:      "disable enabled bundle",
			bundleID:  "b2",
			initial:   true,
			patch:     false,
			wantError: false,
		},
		{
			name:      "patch non-existent bundle",
			bundleID:  "nope",
			initial:   true,
			patch:     false,
			wantError: true,
			errorMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, cleanup := newTestStore(t)
			defer cleanup()

			if tt.bundleID != "nope" {
				mustPutBundle(t, store, tt.bundleID, "slug", "Bundle", tt.initial)
			}

			_, err := store.PatchPromptBundle(t.Context(), &spec.PatchPromptBundleRequest{
				BundleID: tt.bundleID,
				Body:     &spec.PatchPromptBundleRequestBody{IsEnabled: tt.patch},
			})

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Verify patch.
				bundle, err := store.getBundle(tt.bundleID)
				if err != nil {
					t.Errorf("Failed to get patched bundle: %v", err)
				} else if bundle.IsEnabled != tt.patch {
					t.Errorf("Bundle not patched correctly: enabled=%v, want=%v", bundle.IsEnabled, tt.patch)
				}
			}
		})
	}
}

func TestBundleDelete(t *testing.T) {
	tests := []struct {
		name         string
		bundleID     spec.BundleID
		hasTemplates bool
		wantError    bool
		errorMsg     string
	}{
		{
			name:         "delete empty bundle",
			bundleID:     "b1",
			hasTemplates: false,
			wantError:    false,
		},
		{
			name:         "delete non-empty bundle",
			bundleID:     "b2",
			hasTemplates: true,
			wantError:    true,
			errorMsg:     "still contains templates",
		},
		{
			name:      "delete non-existent bundle",
			bundleID:  "nope",
			wantError: true,
			errorMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, cleanup := newTestStore(t)
			defer cleanup()

			if tt.bundleID != "nope" {
				mustPutBundle(t, store, tt.bundleID, "slug", "Bundle", true)
				if tt.hasTemplates {
					mustPutTemplate(t, store, tt.bundleID, "template", "v1", "Template", true)
				}
			}

			_, err := store.DeletePromptBundle(
				t.Context(),
				&spec.DeletePromptBundleRequest{
					BundleID: tt.bundleID,
				},
			)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Verify soft delete.
				_, err = store.getBundle(tt.bundleID)
				if err == nil || !errors.Is(err, ErrBundleDeleting) {
					t.Errorf("Expected soft-deleted error, got %v", err)
				}
			}
		})
	}
}

func TestBundleListFiltering(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	// Create test data.
	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)
	mustPutBundle(t, store, "b2", "slug2", "Bundle 2", false)
	mustPutBundle(t, store, "b3", "slug3", "Bundle 3", true)
	mustPutBundle(t, store, "b4", "slug4", "Bundle 4", false)

	tests := []struct {
		name            string
		bundleIDs       []spec.BundleID
		includeDisabled bool
		wantCount       int
		wantBundles     []spec.BundleID
	}{
		{
			name:            "enabled only",
			bundleIDs:       nil,
			includeDisabled: false,
			wantCount:       2,
			wantBundles:     []spec.BundleID{"b1", "b3"},
		},
		{
			name:            "all bundles",
			bundleIDs:       nil,
			includeDisabled: true,
			wantCount:       4,
			wantBundles:     []spec.BundleID{"b1", "b2", "b3", "b4"},
		},
		{
			name:            "specific bundle IDs",
			bundleIDs:       []spec.BundleID{"b1", "b3"},
			includeDisabled: false,
			wantCount:       2,
			wantBundles:     []spec.BundleID{"b1", "b3"},
		},
		{
			name:            "specific bundle IDs with disabled",
			bundleIDs:       []spec.BundleID{"b2", "b4"},
			includeDisabled: true,
			wantCount:       2,
			wantBundles:     []spec.BundleID{"b2", "b4"},
		},
		{
			name:            "specific bundle IDs without disabled",
			bundleIDs:       []spec.BundleID{"b2", "b4"},
			includeDisabled: false,
			wantCount:       0,
			wantBundles:     []spec.BundleID{},
		},
		{
			name:            "non-existent bundle IDs",
			bundleIDs:       []spec.BundleID{"nope"},
			includeDisabled: true,
			wantCount:       0,
			wantBundles:     []spec.BundleID{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := store.ListPromptBundles(
				t.Context(),
				&spec.ListPromptBundlesRequest{
					BundleIDs:       tt.bundleIDs,
					IncludeDisabled: tt.includeDisabled,
				},
			)
			if err != nil {
				t.Fatalf("ListPromptBundles failed: %v", err)
			}

			if len(resp.Body.PromptBundles) != tt.wantCount {
				t.Errorf("Expected %d bundles, got %d", tt.wantCount, len(resp.Body.PromptBundles))
			}

			gotBundles := make(map[spec.BundleID]bool)
			for _, b := range resp.Body.PromptBundles {
				gotBundles[b.ID] = true
			}

			for _, wantID := range tt.wantBundles {
				if !gotBundles[wantID] {
					t.Errorf("Expected bundle %s not found", wantID)
				}
			}
		})
	}
}

func TestBundleListPagination(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	// Create test data.
	for i := range 25 {
		mustPutBundle(
			t,
			store,
			spec.BundleID((fmt.Sprintf("b%d", i))),
			spec.BundleSlug(fmt.Sprintf("slug%d", i)),
			fmt.Sprintf("Bundle %d", i),
			true,
		)
	}

	tests := []struct {
		name          string
		pageSize      int
		wantFirstPage int
		wantHasNext   bool
	}{
		{
			name:          "default page size",
			pageSize:      0,
			wantFirstPage: 25,
			wantHasNext:   false,
		},
		{
			name:          "small page size",
			pageSize:      5,
			wantFirstPage: 5,
			wantHasNext:   true,
		},
		{
			name:          "large page size",
			pageSize:      30,
			wantFirstPage: 25,
			wantHasNext:   false,
		},
		{
			name:          "exact page size",
			pageSize:      25,
			wantFirstPage: 25,
			wantHasNext:   false,
		},
		{
			name:          "max page size",
			pageSize:      256,
			wantFirstPage: 25,
			wantHasNext:   false,
		},
		{
			name:          "over max page size",
			pageSize:      300,
			wantFirstPage: 25,
			wantHasNext:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := store.ListPromptBundles(
				t.Context(),
				&spec.ListPromptBundlesRequest{
					PageSize: tt.pageSize,
				},
			)
			if err != nil {
				t.Fatalf("ListPromptBundles failed: %v", err)
			}

			if len(resp.Body.PromptBundles) != tt.wantFirstPage {
				t.Errorf(
					"Expected %d bundles on first page, got %d",
					tt.wantFirstPage,
					len(resp.Body.PromptBundles),
				)
			}

			hasNext := resp.Body.NextPageToken != nil && *resp.Body.NextPageToken != ""
			if hasNext != tt.wantHasNext {
				t.Errorf("Expected hasNext=%v, got %v", tt.wantHasNext, hasNext)
			}

			// Test pagination continuity.
			if hasNext {
				resp2, err := store.ListPromptBundles(
					t.Context(),
					&spec.ListPromptBundlesRequest{
						PageSize:  tt.pageSize,
						PageToken: *resp.Body.NextPageToken,
					},
				)
				if err != nil {
					t.Fatalf("Second page request failed: %v", err)
				}
				if len(resp2.Body.PromptBundles) == 0 {
					t.Errorf("Expected results on second page")
				}
			}
		})
	}
}

// --- B. Template Operations ---.

func TestTemplateCRUD(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	// Setup bundle.
	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)

	tests := []struct {
		name         string
		bundleID     spec.BundleID
		templateSlug spec.TemplateSlug
		version      spec.TemplateVersion
		displayName  string
		enabled      bool
		wantError    bool
		errorMsg     string
	}{
		{
			name:         "valid template",
			bundleID:     "b1",
			templateSlug: "template1",
			version:      "v1",
			displayName:  "Template 1",
			enabled:      true,
			wantError:    false,
		},
		{
			name:         "disabled template",
			bundleID:     "b1",
			templateSlug: "template2",
			version:      "v1",
			displayName:  "Template 2",
			enabled:      false,
			wantError:    false,
		},
		{
			name:         "empty bundle ID",
			bundleID:     "",
			templateSlug: "template3",
			version:      "v1",
			displayName:  "Template 3",
			enabled:      true,
			wantError:    true,
			errorMsg:     "required",
		},
		{
			name:         "empty template slug",
			bundleID:     "b1",
			templateSlug: "",
			version:      "v1",
			displayName:  "Template 4",
			enabled:      true,
			wantError:    true,
			errorMsg:     "required",
		},
		{
			name:         "empty version",
			bundleID:     "b1",
			templateSlug: "template5",
			version:      "",
			displayName:  "Template 5",
			enabled:      true,
			wantError:    true,
			errorMsg:     "required",
		},
		{
			name:         "empty display name",
			bundleID:     "b1",
			templateSlug: "template6",
			version:      "v1",
			displayName:  "",
			enabled:      true,
			wantError:    true,
			errorMsg:     "required",
		},
		{
			name:         "invalid template slug",
			bundleID:     "b1",
			templateSlug: "bad.slug",
			version:      "v1",
			displayName:  "Template 7",
			enabled:      true,
			wantError:    true,
			errorMsg:     "invalid slug",
		},
		{
			name:         "invalid version",
			bundleID:     "b1",
			templateSlug: "template8",
			version:      "bad&version",
			displayName:  "Template 8",
			enabled:      true,
			wantError:    true,
			errorMsg:     "invalid version",
		},
		{
			name:         "non-existent bundle",
			bundleID:     "nope",
			templateSlug: "template9",
			version:      "v1",
			displayName:  "Template 9",
			enabled:      true,
			wantError:    true,
			errorMsg:     "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
				BundleID:     tt.bundleID,
				TemplateSlug: tt.templateSlug,
				Body: &spec.PutPromptTemplateRequestBody{
					DisplayName: tt.displayName,
					IsEnabled:   tt.enabled,
					Blocks:      []spec.MessageBlock{{ID: "1", Role: spec.User, Content: "Hello"}},
					Version:     tt.version,
				},
			})

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTemplateVersionConflict(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)
	mustPutTemplate(t, store, "b1", "template1", "v1", "Template 1", true)

	// Try to create duplicate.
	_, err := store.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
		BundleID:     "b1",
		TemplateSlug: "template1",
		Body: &spec.PutPromptTemplateRequestBody{
			DisplayName: "Template 1 Duplicate",
			IsEnabled:   true,
			Blocks:      []spec.MessageBlock{{ID: "1", Role: spec.User, Content: "Hello"}},
			Version:     "v1",
		},
	})

	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected already exists error, got %v", err)
	}
}

func TestTemplateDisabledBundle(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", false)

	_, err := store.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
		BundleID:     "b1",
		TemplateSlug: "template1",
		Body: &spec.PutPromptTemplateRequestBody{
			DisplayName: "Template 1",
			IsEnabled:   true,
			Blocks:      []spec.MessageBlock{{ID: "1", Role: spec.User, Content: "Hello"}},
			Version:     "v1",
		},
	})

	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Errorf("Expected disabled bundle error, got %v", err)
	}
}

func TestTemplateMultiVersion(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)

	// Create multiple versions with different modification times.
	mustPutTemplate(t, store, "b1", "template1", "v1", "Template 1 v1", true)
	time.Sleep(10 * time.Millisecond)
	mustPutTemplate(t, store, "b1", "template1", "v2", "Template 1 v2", false)
	time.Sleep(10 * time.Millisecond)
	mustPutTemplate(t, store, "b1", "template1", "v3", "Template 1 v3", true)

	// Test getting specific version.
	resp, err := store.GetPromptTemplate(t.Context(), &spec.GetPromptTemplateRequest{
		BundleID:     "b1",
		TemplateSlug: "template1",
		Version:      "v2",
	})
	if err != nil {
		t.Fatalf("GetPromptTemplate failed: %v", err)
	}
	if resp.Body.Version != "v2" {
		t.Errorf("Expected version v2, got %s", resp.Body.Version)
	}

	// Test getting active version (should be v3 - latest enabled).
	resp, err = store.GetPromptTemplate(t.Context(), &spec.GetPromptTemplateRequest{
		BundleID:     "b1",
		TemplateSlug: "template1",
	})
	if err != nil {
		t.Fatalf("GetPromptTemplate failed: %v", err)
	}
	if resp.Body.Version != "v3" {
		t.Errorf("Expected active version v3, got %s", resp.Body.Version)
	}
}

func TestTemplateActiveVersionSelection(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)

	tests := []struct {
		name     string
		slugName spec.TemplateSlug
		versions []struct {
			version spec.TemplateVersion
			enabled bool
		}
		expectedActive spec.TemplateVersion
	}{
		{
			name:     "single enabled version",
			slugName: "singleEnabledVersion",
			versions: []struct {
				version spec.TemplateVersion
				enabled bool
			}{
				{"v1", true},
			},
			expectedActive: "v1",
		},
		{
			name:     "latest enabled wins",
			slugName: "latestEnabledWins",
			versions: []struct {
				version spec.TemplateVersion
				enabled bool
			}{
				{"v1", true},
				{"v2", true},
			},
			expectedActive: "v2",
		},
		{
			name:     "skip disabled versions",
			slugName: "skipDisabledVersions",
			versions: []struct {
				version spec.TemplateVersion
				enabled bool
			}{
				{"v1", true},
				{"v2", false},
				{"v3", true},
			},
			expectedActive: "v3",
		},
		{
			name:     "no enabled versions",
			slugName: "noEnabledVersions",
			versions: []struct {
				version spec.TemplateVersion
				enabled bool
			}{
				{"v1", false},
				{"v2", false},
			},
			expectedActive: "", // Should fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templateSlug := "template-" + tt.slugName

			// Create versions.
			for i, v := range tt.versions {
				mustPutTemplate(
					t,
					store,
					"b1",
					templateSlug,
					v.version,
					fmt.Sprintf("Template %d", i),
					v.enabled,
				)
				if i < len(tt.versions)-1 {
					time.Sleep(10 * time.Millisecond) // Ensure different modification times
				}
			}

			// Test getting active version.
			resp, err := store.GetPromptTemplate(
				t.Context(),
				&spec.GetPromptTemplateRequest{
					BundleID:     "b1",
					TemplateSlug: templateSlug,
				},
			)

			if tt.expectedActive == "" {
				if err == nil {
					t.Errorf("Expected error for no active version, got success")
				}
			} else {
				if err != nil {
					t.Fatalf("GetPromptTemplate failed: %v", err)
				}
				if resp.Body.Version != tt.expectedActive {
					t.Errorf("Expected active version %s, got %s", tt.expectedActive, resp.Body.Version)
				}
			}
		})
	}
}

func TestTemplatePatch(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)
	mustPutTemplate(t, store, "b1", "template1", "v1", "Template 1", false)

	// Test patch enable.
	_, err := store.PatchPromptTemplate(t.Context(), &spec.PatchPromptTemplateRequest{
		BundleID:     "b1",
		TemplateSlug: "template1",
		Body:         &spec.PatchPromptTemplateRequestBody{Version: "v1", IsEnabled: true},
	})
	if err != nil {
		t.Fatalf("PatchPromptTemplate failed: %v", err)
	}

	// Verify patch.
	resp, err := store.GetPromptTemplate(t.Context(), &spec.GetPromptTemplateRequest{
		BundleID:     "b1",
		TemplateSlug: "template1",
		Version:      "v1",
	})
	if err != nil {
		t.Fatalf("GetPromptTemplate failed: %v", err)
	}
	if !resp.Body.IsEnabled {
		t.Errorf("Expected template to be enabled after patch")
	}
}

func TestTemplateDelete(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)
	mustPutTemplate(t, store, "b1", "template1", "v1", "Template 1", true)

	// Delete template.
	_, err := store.DeletePromptTemplate(t.Context(), &spec.DeletePromptTemplateRequest{
		BundleID:     "b1",
		TemplateSlug: "template1",
		Version:      "v1",
	})
	if err != nil {
		t.Fatalf("DeletePromptTemplate failed: %v", err)
	}

	// Verify deletion.
	_, err = store.GetPromptTemplate(t.Context(), &spec.GetPromptTemplateRequest{
		BundleID:     "b1",
		TemplateSlug: "template1",
		Version:      "v1",
	})
	if err == nil {
		t.Errorf("Expected error for deleted template")
	}
}

func TestTemplateListFiltering(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	// Setup test data.
	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)
	mustPutBundle(t, store, "b2", "slug2", "Bundle 2", true)

	mustPutTemplate(t, store, "b1", "template1", "v1", "Template 1", true, "tag1", "tag2")
	mustPutTemplate(t, store, "b1", "template2", "v1", "Template 2", false, "tag2")
	mustPutTemplate(t, store, "b2", "template3", "v1", "Template 3", true, "tag3")

	tests := []struct {
		name            string
		bundleIDs       []spec.BundleID
		tags            []string
		includeDisabled bool
		allVersions     bool
		wantCount       int
	}{
		{
			name:            "all enabled templates",
			bundleIDs:       nil,
			tags:            nil,
			includeDisabled: false,
			allVersions:     false,
			wantCount:       2,
		},
		{
			name:            "all templates including disabled",
			bundleIDs:       nil,
			tags:            nil,
			includeDisabled: true,
			allVersions:     false,
			wantCount:       3,
		},
		{
			name:            "filter by bundle ID",
			bundleIDs:       []spec.BundleID{"b1"},
			tags:            nil,
			includeDisabled: true,
			allVersions:     false,
			wantCount:       2,
		},
		{
			name:            "filter by tag",
			bundleIDs:       nil,
			tags:            []string{"tag1"},
			includeDisabled: false,
			allVersions:     false,
			wantCount:       1,
		},
		{
			name:            "filter by multiple tags",
			bundleIDs:       nil,
			tags:            []string{"tag2"},
			includeDisabled: true,
			allVersions:     false,
			wantCount:       2,
		},
		{
			name:            "filter by non-existent tag",
			bundleIDs:       nil,
			tags:            []string{"nonexistent"},
			includeDisabled: true,
			allVersions:     false,
			wantCount:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := store.ListPromptTemplates(
				t.Context(),
				&spec.ListPromptTemplatesRequest{
					BundleIDs:       tt.bundleIDs,
					Tags:            tt.tags,
					IncludeDisabled: tt.includeDisabled,
					AllVersions:     tt.allVersions,
				},
			)
			if err != nil {
				t.Fatalf("ListPromptTemplates failed: %v", err)
			}

			if len(resp.Body.PromptTemplateListItems) != tt.wantCount {
				t.Errorf(
					"Expected %d templates, got %d",
					tt.wantCount,
					len(resp.Body.PromptTemplateListItems),
				)
			}
		})
	}
}

func TestTemplateListAllVersions(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)
	mustPutTemplate(t, store, "b1", "template1", "v1", "Template 1 v1", true)
	mustPutTemplate(t, store, "b1", "template1", "v2", "Template 1 v2", false)
	mustPutTemplate(t, store, "b1", "template1", "v3", "Template 1 v3", true)

	// Test without allVersions (should get only active version).
	resp, err := store.ListPromptTemplates(t.Context(), &spec.ListPromptTemplatesRequest{})
	if err != nil {
		t.Fatalf("ListPromptTemplates failed: %v", err)
	}
	if len(resp.Body.PromptTemplateListItems) != 1 {
		t.Errorf(
			"Expected 1 template (active version), got %d",
			len(resp.Body.PromptTemplateListItems),
		)
	}

	// Test with allVersions.
	resp, err = store.ListPromptTemplates(t.Context(), &spec.ListPromptTemplatesRequest{
		AllVersions:     true,
		IncludeDisabled: true,
	})
	if err != nil {
		t.Fatalf("ListPromptTemplates failed: %v", err)
	}
	if len(resp.Body.PromptTemplateListItems) != 3 {
		t.Errorf(
			"Expected 3 templates (all versions), got %d",
			len(resp.Body.PromptTemplateListItems),
		)
	}
}

func TestTemplateListPagination(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)

	// Create 15 templates.
	for i := range 15 {
		mustPutTemplate(
			t,
			store,
			"b1",
			spec.TemplateSlug(fmt.Sprintf("template%d", i)),
			"v1",
			fmt.Sprintf("Template %d", i),
			true,
		)
	}

	// Test pagination.
	resp, err := store.ListPromptTemplates(t.Context(), &spec.ListPromptTemplatesRequest{
		PageSize: 5,
	})
	if err != nil {
		t.Fatalf("ListPromptTemplates failed: %v", err)
	}

	if len(resp.Body.PromptTemplateListItems) != 5 {
		t.Errorf(
			"Expected 5 templates on first page, got %d",
			len(resp.Body.PromptTemplateListItems),
		)
	}

	if resp.Body.NextPageToken == nil {
		t.Errorf("Expected next page token")
	}

	// Test second page.
	resp2, err := store.ListPromptTemplates(t.Context(), &spec.ListPromptTemplatesRequest{
		PageSize:  5,
		PageToken: *resp.Body.NextPageToken,
	})
	if err != nil {
		t.Fatalf("Second page request failed: %v", err)
	}

	if len(resp2.Body.PromptTemplateListItems) != 5 {
		t.Errorf(
			"Expected 5 templates on second page, got %d",
			len(resp2.Body.PromptTemplateListItems),
		)
	}
}

// --- C. Search Operations ---.

func TestSearchTemplates(t *testing.T) {
	store, cleanup := newTestStoreWithFTS(t)
	defer cleanup()

	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)
	mustPutTemplate(t, store, "b1", "hello", "v1", "Hello World", true, "greeting")
	mustPutTemplate(t, store, "b1", "goodbye", "v1", "Goodbye World", true, "farewell")

	// Allow time for FTS indexing.
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name            string
		query           string
		includeDisabled bool
		wantError       bool
		minResults      int
	}{
		{
			name:            "search hello",
			query:           "hello",
			includeDisabled: false,
			wantError:       false,
			minResults:      1,
		},
		{
			name:            "search world",
			query:           "world",
			includeDisabled: false,
			wantError:       false,
			minResults:      2,
		},
		{
			name:            "search greeting",
			query:           "greeting",
			includeDisabled: false,
			wantError:       false,
			minResults:      1,
		},
		{
			name:            "empty query",
			query:           "",
			includeDisabled: false,
			wantError:       true,
			minResults:      0,
		},
		{
			name:            "non-existent term",
			query:           "nonexistent",
			includeDisabled: false,
			wantError:       false,
			minResults:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := store.SearchPromptTemplates(
				t.Context(),
				&spec.SearchPromptTemplatesRequest{
					Query:           tt.query,
					IncludeDisabled: tt.includeDisabled,
				},
			)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("SearchPromptTemplates failed: %v", err)
				}
				if len(resp.Body.PromptTemplateListItems) < tt.minResults {
					t.Errorf("Expected at least %d results, got %d", tt.minResults, len(resp.Body.PromptTemplateListItems))
				}
			}
		})
	}
}

func TestSearchTemplatesWithoutFTS(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	_, err := store.SearchPromptTemplates(t.Context(), &spec.SearchPromptTemplatesRequest{
		Query: "hello",
	})

	if err == nil || !errors.Is(err, ErrFTSDisabled) {
		t.Errorf("Expected FTS disabled error, got %v", err)
	}
}

// --- D. Validation ---.

func TestSlugVersionValidation(t *testing.T) {
	tests := []struct {
		name    string
		slug    spec.TemplateSlug
		version spec.TemplateVersion
		valid   bool
	}{
		{"valid basic", "abc", "v1", true},
		{"valid with dash", "a-b-c", "v-1", true},
		{"valid with numbers", "abc123", "v123", true},
		{"valid uppercase", "ABC", "V1", true},
		{
			"valid long",
			spec.TemplateSlug(strings.Repeat("a", 64)),
			spec.TemplateVersion(strings.Repeat("v", 64)),
			true,
		},
		{"empty slug", "", "v1", false},
		{"empty version", "abc", "", false},
		{"slug with dot", "abc.def", "v1", false},
		{"version with dot", "abc", "v.1", true},
		{"slug with underscore", "abc_def", "v1", false},
		{"version with underscore", "abc", "v_1", false},
		{"slug with space", "abc def", "v1", false},
		{"version with space", "abc", "v 1", false},
		{"slug with slash", "abc/def", "v1", false},
		{"version with slash", "abc", "v/1", false},
		{"slug too long", spec.TemplateSlug(strings.Repeat("a", 65)), "v1", false},
		{"version too long", "abc", spec.TemplateVersion(strings.Repeat("v", 65)), false},
		{"slug with special chars", "abc!", "v1", false},
		{"version with special chars", "abc", "v1!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slugErr := nameutils.ValidateTemplateSlug(tt.slug)
			versionErr := nameutils.ValidateTemplateVersion(tt.version)

			if tt.valid {
				if slugErr != nil {
					t.Errorf("Expected valid slug %q, got error: %v", tt.slug, slugErr)
				}
				if versionErr != nil {
					t.Errorf("Expected valid version %q, got error: %v", tt.version, versionErr)
				}
			} else if slugErr == nil && versionErr == nil {
				t.Errorf("Expected invalid for slug=%q version=%q", tt.slug, tt.version)
			}
		})
	}
}

// --- E. Edge Cases & Boundaries ---.

func TestPaginationEdgeCases(t *testing.T) {
	_, cleanup := newTestStore(t)
	defer cleanup()

	tests := []struct {
		name      string
		dataCount int
		pageSize  int
		wantPages int
	}{
		{"empty dataset", 0, 10, 0},
		{"single item", 1, 10, 1},
		{"exact page size", 10, 10, 1},
		{"page size + 1", 11, 10, 2},
		{"page size - 1", 9, 10, 1},
		{"multiple full pages", 20, 10, 2},
		{"multiple pages + partial", 25, 10, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh store for each test.
			localStore, localCleanup := newTestStore(t)
			defer localCleanup()

			// Create test data.
			for i := 0; i < tt.dataCount; i++ {
				mustPutBundle(
					t,
					localStore,
					spec.BundleID(fmt.Sprintf("b%d", i)),
					spec.BundleSlug(fmt.Sprintf("slug%d", i)),
					fmt.Sprintf("Bundle %d", i),
					true,
				)
			}

			// Test pagination.
			pageCount := 0
			pageToken := ""
			totalItems := 0

			for {
				resp, err := localStore.ListPromptBundles(
					t.Context(),
					&spec.ListPromptBundlesRequest{
						PageSize:  tt.pageSize,
						PageToken: pageToken,
					},
				)
				if err != nil {
					t.Fatalf("ListPromptBundles failed: %v", err)
				}

				if len(resp.Body.PromptBundles) == 0 {
					break
				}

				pageCount++
				totalItems += len(resp.Body.PromptBundles)

				if resp.Body.NextPageToken == nil || *resp.Body.NextPageToken == "" {
					break
				}
				pageToken = *resp.Body.NextPageToken
			}

			if pageCount != tt.wantPages {
				t.Errorf("Expected %d pages, got %d", tt.wantPages, pageCount)
			}
			if totalItems != tt.dataCount {
				t.Errorf("Expected %d total items, got %d", tt.dataCount, totalItems)
			}
		})
	}
}

func TestSoftDeleteTiming(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)

	// Delete bundle.
	_, err := store.DeletePromptBundle(t.Context(), &spec.DeletePromptBundleRequest{
		BundleID: "b1",
	})
	if err != nil {
		t.Fatalf("DeletePromptBundle failed: %v", err)
	}

	// Bundle should be soft-deleted.
	_, err = store.getBundle("b1")
	if err == nil || !errors.Is(err, ErrBundleDeleting) {
		t.Errorf("Expected bundle to be soft-deleted, got error: %v", err)
	}

	// Simulate grace period passed.
	b, _ := store.bundleStore.GetKey([]string{"bundles", "b1"})
	if mp, ok := b.(map[string]any); ok {
		mp["softDeletedAt"] = time.Now().Add(-2 * softDeleteGrace).UTC().Format(time.RFC3339Nano)
		_ = store.bundleStore.SetKey([]string{"bundles", "b1"}, mp)
	}

	// Run cleanup.
	store.sweepSoftDeleted()

	// Bundle should be hard-deleted.
	_, err = store.bundleStore.GetKey([]string{"bundles", "b1"})
	if err == nil {
		t.Error("Expected bundle to be hard-deleted")
	}
}

func TestConcurrentOperations(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)

	// Test concurrent template creation with same slug but different versions.
	done := make(chan bool, 2)
	errorsChan := make(chan error, 2)

	go func() {
		_, err := store.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
			BundleID:     "b1",
			TemplateSlug: "concurrent",
			Body: &spec.PutPromptTemplateRequestBody{
				DisplayName: "Concurrent 1",
				IsEnabled:   true,
				Blocks:      []spec.MessageBlock{{ID: "1", Role: spec.User, Content: "Hello"}},
				Version:     "v1",
			},
		})
		errorsChan <- err
		done <- true
	}()

	go func() {
		_, err := store.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
			BundleID:     "b1",
			TemplateSlug: "concurrent",
			Body: &spec.PutPromptTemplateRequestBody{
				DisplayName: "Concurrent 2",
				IsEnabled:   true,
				Blocks:      []spec.MessageBlock{{ID: "1", Role: spec.User, Content: "Hello"}},
				Version:     "v2",
			},
		})
		errorsChan <- err
		done <- true
	}()

	// Wait for both operations.
	<-done
	<-done

	// Both should succeed (different versions).
	err1 := <-errorsChan
	err2 := <-errorsChan

	if err1 != nil {
		t.Errorf("First concurrent operation failed: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Second concurrent operation failed: %v", err2)
	}
}

func TestInvalidRequests(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "nil bundle request",
			fn: func() error {
				_, err := store.PutPromptBundle(t.Context(), nil)
				return err
			},
		},
		{
			name: "nil bundle body",
			fn: func() error {
				_, err := store.PutPromptBundle(
					t.Context(),
					&spec.PutPromptBundleRequest{},
				)
				return err
			},
		},
		{
			name: "nil template request",
			fn: func() error {
				_, err := store.PutPromptTemplate(t.Context(), nil)
				return err
			},
		},
		{
			name: "nil template body",
			fn: func() error {
				_, err := store.PutPromptTemplate(
					t.Context(),
					&spec.PutPromptTemplateRequest{},
				)
				return err
			},
		},
		{
			name: "nil patch request",
			fn: func() error {
				_, err := store.PatchPromptBundle(t.Context(), nil)
				return err
			},
		},
		{
			name: "nil delete request",
			fn: func() error {
				_, err := store.DeletePromptBundle(t.Context(), nil)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Errorf("Expected error for %s", tt.name)
			}
		})
	}
}

func TestCreatedAtPreservation(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	// Create bundle.
	_, err := store.PutPromptBundle(t.Context(), &spec.PutPromptBundleRequest{
		BundleID: "b1",
		Body: &spec.PutPromptBundleRequestBody{
			Slug:        "slug1",
			DisplayName: "Bundle 1",
			IsEnabled:   true,
		},
	})
	if err != nil {
		t.Fatalf("PutPromptBundle failed: %v", err)
	}

	// Get original creation time.
	bundle1, err := store.getBundle("b1")
	if err != nil {
		t.Fatalf("getBundle failed: %v", err)
	}
	originalCreatedAt := bundle1.CreatedAt

	// Update bundle.
	time.Sleep(10 * time.Millisecond)
	_, err = store.PutPromptBundle(t.Context(), &spec.PutPromptBundleRequest{
		BundleID: "b1",
		Body: &spec.PutPromptBundleRequestBody{
			Slug:        "updated-slug",
			DisplayName: "Updated Bundle",
			IsEnabled:   false,
		},
	})
	if err != nil {
		t.Fatalf("PutPromptBundle update failed: %v", err)
	}

	// Verify CreatedAt is preserved.
	bundle2, err := store.getBundle("b1")
	if err != nil {
		t.Fatalf("getBundle failed: %v", err)
	}

	if !bundle2.CreatedAt.Equal(originalCreatedAt) {
		t.Errorf(
			"CreatedAt not preserved: original=%v, updated=%v",
			originalCreatedAt,
			bundle2.CreatedAt,
		)
	}

	if !bundle2.ModifiedAt.After(originalCreatedAt) {
		t.Errorf("ModifiedAt should be after CreatedAt")
	}
}

func TestEmptyResults(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	// Test empty bundle list.
	resp, err := store.ListPromptBundles(t.Context(), &spec.ListPromptBundlesRequest{})
	if err != nil {
		t.Fatalf("ListPromptBundles failed: %v", err)
	}
	if len(resp.Body.PromptBundles) != 0 {
		t.Errorf("Expected 0 bundles, got %d", len(resp.Body.PromptBundles))
	}
	if resp.Body.NextPageToken != nil {
		t.Errorf("Expected no next page token for empty results")
	}

	// Test empty template list.
	resp2, err := store.ListPromptTemplates(
		t.Context(),
		&spec.ListPromptTemplatesRequest{},
	)
	if err != nil {
		t.Fatalf("ListPromptTemplates failed: %v", err)
	}
	if len(resp2.Body.PromptTemplateListItems) != 0 {
		t.Errorf("Expected 0 templates, got %d", len(resp2.Body.PromptTemplateListItems))
	}
	if resp2.Body.NextPageToken != nil {
		t.Errorf("Expected no next page token for empty results")
	}
}
