package store

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
)

// newTestStore creates a new PromptTemplateStore for testing and returns a cleanup function.
func newTestStore(t *testing.T) (store *PromptTemplateStore, cleanupFunc func()) {
	t.Helper()
	dir := t.TempDir()
	store, err := NewPromptTemplateStore(dir)
	if err != nil {
		t.Fatalf("Failed to create store: %v.", err)
	}
	cleanupFunc = func() { store.Close(); os.RemoveAll(dir) }
	return store, cleanupFunc
}

// mustPutBundle inserts a bundle or fails the test.
func mustPutBundle(t *testing.T, s *PromptTemplateStore, id, slug, name string, enabled bool) {
	t.Helper()
	_, err := s.PutPromptBundle(t.Context(), &spec.PutPromptBundleRequest{
		BundleID: id,
		Body: &spec.PutPromptBundleRequestBody{
			Slug:        slug,
			DisplayName: name,
			IsEnabled:   enabled,
		},
	})
	if err != nil {
		t.Fatalf("PutPromptBundle failed: %v.", err)
	}
}

// mustPutTemplate inserts a template or fails the test.
func mustPutTemplate(
	t *testing.T,
	s *PromptTemplateStore,
	bundleID, slug, version, name string,
	enabled bool,
	tags ...string,
) {
	t.Helper()
	_, err := s.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
		BundleID:     bundleID,
		TemplateSlug: slug,
		Body: &spec.PutPromptTemplateRequestBody{
			DisplayName: name,
			IsEnabled:   enabled,
			Blocks:      []spec.MessageBlock{{ID: "1", Role: spec.User, Content: "Hello"}},
			Version:     version,
			Tags:        tags,
		},
	})
	if err != nil {
		t.Fatalf("PutPromptTemplate failed: %v.", err)
	}
}

func TestBundleCRUDAndValidation(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	t.Run("Create bundle happy path", func(t *testing.T) {
		_, err := store.PutPromptBundle(t.Context(), &spec.PutPromptBundleRequest{
			BundleID: "b1",
			Body: &spec.PutPromptBundleRequestBody{
				Slug:        "slug1",
				DisplayName: "Bundle 1",
				IsEnabled:   true,
			},
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v.", err)
		}
	})

	t.Run("Create bundle with invalid slug", func(t *testing.T) {
		_, err := store.PutPromptBundle(t.Context(), &spec.PutPromptBundleRequest{
			BundleID: "b2",
			Body: &spec.PutPromptBundleRequestBody{
				Slug:        "bad.slug",
				DisplayName: "Bundle 2",
				IsEnabled:   true,
			},
		})
		if err == nil || !strings.Contains(err.Error(), "invalid slug") {
			t.Errorf("Expected invalid slug error, got %v.", err)
		}
	})

	t.Run("Create bundle with missing fields", func(t *testing.T) {
		_, err := store.PutPromptBundle(t.Context(), &spec.PutPromptBundleRequest{
			BundleID: "",
			Body: &spec.PutPromptBundleRequestBody{
				Slug:        "slug",
				DisplayName: "Name",
				IsEnabled:   true,
			},
		})
		if err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("Expected required error, got %v.", err)
		}
		_, err = store.PutPromptBundle(t.Context(), &spec.PutPromptBundleRequest{
			BundleID: "b3",
			Body: &spec.PutPromptBundleRequestBody{
				Slug:        "",
				DisplayName: "Name",
				IsEnabled:   true,
			},
		})
		if err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("Expected required error, got %v.", err)
		}
		_, err = store.PutPromptBundle(t.Context(), &spec.PutPromptBundleRequest{
			BundleID: "b3",
			Body: &spec.PutPromptBundleRequestBody{
				Slug:        "slug",
				DisplayName: "",
				IsEnabled:   true,
			},
		})
		if err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("Expected required error, got %v.", err)
		}
	})

	t.Run("Patch bundle enabled/disabled", func(t *testing.T) {
		mustPutBundle(t, store, "b4", "slug4", "Bundle 4", false)
		_, err := store.PatchPromptBundle(t.Context(), &spec.PatchPromptBundleRequest{
			BundleID: "b4",
			Body:     &spec.PatchPromptBundleRequestBody{IsEnabled: true},
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v.", err)
		}
		b, _ := store.getBundle("b4")
		if !b.IsEnabled {
			t.Errorf("Expected enabled=true, got false.")
		}
		_, err = store.PatchPromptBundle(t.Context(), &spec.PatchPromptBundleRequest{
			BundleID: "b4",
			Body:     &spec.PatchPromptBundleRequestBody{IsEnabled: false},
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v.", err)
		}
		b, _ = store.getBundle("b4")
		if b.IsEnabled {
			t.Errorf("Expected enabled=false, got true.")
		}
	})

	t.Run("Patch non-existent bundle", func(t *testing.T) {
		_, err := store.PatchPromptBundle(t.Context(), &spec.PatchPromptBundleRequest{
			BundleID: "nope",
			Body:     &spec.PatchPromptBundleRequestBody{IsEnabled: false},
		})
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error, got %v.", err)
		}
	})

	t.Run("Delete empty bundle", func(t *testing.T) {
		mustPutBundle(t, store, "b5", "slug5", "Bundle 5", true)
		_, err := store.DeletePromptBundle(
			t.Context(),
			&spec.DeletePromptBundleRequest{BundleID: "b5"},
		)
		if err != nil {
			t.Fatalf("Unexpected error: %v.", err)
		}
		_, err = store.getBundle("b5")
		if err == nil || !errors.Is(err, ErrBundleDeleting) {
			t.Errorf("Expected soft-deleted error, got %v.", err)
		}
	})

	t.Run("Delete non-existent bundle", func(t *testing.T) {
		_, err := store.DeletePromptBundle(
			t.Context(),
			&spec.DeletePromptBundleRequest{BundleID: "nope"},
		)
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error, got %v.", err)
		}
	})

	t.Run("Delete non-empty bundle", func(t *testing.T) {
		mustPutBundle(t, store, "b6", "slug6", "Bundle 6", true)
		mustPutTemplate(t, store, "b6", "t1", "v1", "T1", true)
		_, err := store.DeletePromptBundle(
			t.Context(),
			&spec.DeletePromptBundleRequest{BundleID: "b6"},
		)
		if err == nil || !strings.Contains(err.Error(), "still contains templates") {
			t.Errorf("Expected bundle not empty error, got %v.", err)
		}
	})

	t.Run("Delete bundle twice", func(t *testing.T) {
		mustPutBundle(t, store, "b7", "slug7", "Bundle 7", true)
		_, err := store.DeletePromptBundle(
			t.Context(),
			&spec.DeletePromptBundleRequest{BundleID: "b7"},
		)
		if err != nil {
			t.Fatalf("Unexpected error: %v.", err)
		}
		_, err = store.DeletePromptBundle(
			t.Context(),
			&spec.DeletePromptBundleRequest{BundleID: "b7"},
		)
		if err == nil || !strings.Contains(err.Error(), "being deleted") {
			t.Errorf("Expected being deleted error, got %v.", err)
		}
	})
}

func TestBundleListingAndPagination(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()
	for i := 0; i < 7; i++ {
		mustPutBundle(
			t,
			store,
			fmt.Sprintf("b%d", i),
			fmt.Sprintf("slug%d", i),
			fmt.Sprintf("Bundle%d", i),
			i%2 == 0,
		)
	}
	t.Run("List enabled only", func(t *testing.T) {
		resp, err := store.ListPromptBundles(t.Context(), &spec.ListPromptBundlesRequest{})
		if err != nil {
			t.Fatalf("Unexpected error: %v.", err)
		}
		for _, b := range resp.Body.PromptBundles {
			if !b.IsEnabled {
				t.Errorf("Expected only enabled bundles, got disabled: %v.", b.ID)
			}
		}
	})

	t.Run("List all with disabled", func(t *testing.T) {
		resp, err := store.ListPromptBundles(
			t.Context(),
			&spec.ListPromptBundlesRequest{IncludeDisabled: true},
		)
		if err != nil {
			t.Fatalf("Unexpected error: %v.", err)
		}
		if len(resp.Body.PromptBundles) != 7 {
			t.Errorf("Expected 7 bundles, got %d.", len(resp.Body.PromptBundles))
		}
	})

	t.Run("List by IDs", func(t *testing.T) {
		resp, err := store.ListPromptBundles(
			t.Context(),
			&spec.ListPromptBundlesRequest{
				BundleIDs:       []string{"b1", "b3", "b5"},
				IncludeDisabled: true,
			},
		)
		if err != nil {
			t.Fatalf("Unexpected error: %v.", err)
		}
		if len(resp.Body.PromptBundles) != 3 {
			t.Errorf("Expected 3 bundles, got %d.", len(resp.Body.PromptBundles))
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		resp, err := store.ListPromptBundles(
			t.Context(),
			&spec.ListPromptBundlesRequest{PageSize: 2, IncludeDisabled: true},
		)
		if err != nil {
			t.Fatalf("Unexpected error: %v.", err)
		}
		if len(resp.Body.PromptBundles) != 2 {
			t.Errorf("Expected 2 bundles, got %d.", len(resp.Body.PromptBundles))
		}
		if resp.Body.NextPageToken == nil || *resp.Body.NextPageToken == "" {
			t.Errorf("Expected next page token, got %v.", resp.Body.NextPageToken)
		}
		// Fetch next page.
		resp2, err := store.ListPromptBundles(
			t.Context(),
			&spec.ListPromptBundlesRequest{
				PageSize:        2,
				IncludeDisabled: true,
				PageToken:       *resp.Body.NextPageToken,
			},
		)
		if err != nil {
			t.Fatalf("Unexpected error: %v.", err)
		}
		if len(resp2.Body.PromptBundles) == 0 {
			t.Errorf("Expected more bundles in next page.")
		}
	})
}

func TestTemplateCRUDAndValidation(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()
	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)

	t.Run("Create template happy path", func(t *testing.T) {
		mustPutTemplate(t, store, "b1", "t1", "v1", "T1", true)
	})

	t.Run("Create duplicate slug+version", func(t *testing.T) {
		mustPutTemplate(t, store, "b1", "dup", "v1", "Dup", true)
		_, err := store.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
			BundleID:     "b1",
			TemplateSlug: "dup",
			Body: &spec.PutPromptTemplateRequestBody{
				DisplayName: "Dup",
				IsEnabled:   true,
				Blocks:      []spec.MessageBlock{{ID: "1", Role: spec.User, Content: "Hello"}},
				Version:     "v1",
			},
		})
		if err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Expected already exists error, got %v.", err)
		}
	})

	t.Run("Create with invalid slug", func(t *testing.T) {
		_, err := store.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
			BundleID:     "b1",
			TemplateSlug: "bad.slug",
			Body: &spec.PutPromptTemplateRequestBody{
				DisplayName: "T2",
				IsEnabled:   true,
				Blocks:      []spec.MessageBlock{{ID: "1", Role: spec.User, Content: "Hello"}},
				Version:     "v1",
			},
		})
		if err == nil || !strings.Contains(err.Error(), "invalid slug") {
			t.Errorf("Expected invalid slug error, got %v.", err)
		}
	})

	t.Run("Create with invalid version", func(t *testing.T) {
		_, err := store.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
			BundleID:     "b1",
			TemplateSlug: "t2",
			Body: &spec.PutPromptTemplateRequestBody{
				DisplayName: "T2",
				IsEnabled:   true,
				Blocks:      []spec.MessageBlock{{ID: "1", Role: spec.User, Content: "Hello"}},
				Version:     "bad.version",
			},
		})
		if err == nil || !strings.Contains(err.Error(), "invalid version") {
			t.Errorf("Expected invalid version error, got %v.", err)
		}
	})

	t.Run("Create with missing fields", func(t *testing.T) {
		_, err := store.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
			BundleID:     "",
			TemplateSlug: "t2",
			Body: &spec.PutPromptTemplateRequestBody{
				DisplayName: "T2",
				IsEnabled:   true,
				Blocks:      []spec.MessageBlock{{ID: "1", Role: spec.User, Content: "Hello"}},
				Version:     "v1",
			},
		})
		if err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("Expected required error, got %v.", err)
		}
		_, err = store.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
			BundleID:     "b1",
			TemplateSlug: "",
			Body: &spec.PutPromptTemplateRequestBody{
				DisplayName: "T2",
				IsEnabled:   true,
				Blocks:      []spec.MessageBlock{{ID: "1", Role: spec.User, Content: "Hello"}},
				Version:     "v1",
			},
		})
		if err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("Expected required error, got %v.", err)
		}
		_, err = store.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
			BundleID:     "b1",
			TemplateSlug: "t2",
			Body: &spec.PutPromptTemplateRequestBody{
				DisplayName: "",
				IsEnabled:   true,
				Blocks:      []spec.MessageBlock{{ID: "1", Role: spec.User, Content: "Hello"}},
				Version:     "v1",
			},
		})
		if err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("Expected required error, got %v.", err)
		}
		_, err = store.PutPromptTemplate(t.Context(), &spec.PutPromptTemplateRequest{
			BundleID:     "b1",
			TemplateSlug: "t2",
			Body: &spec.PutPromptTemplateRequestBody{
				DisplayName: "T2",
				IsEnabled:   true,
				Blocks:      nil,
				Version:     "v1",
			},
		})
		if err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("Expected required error, got %v.", err)
		}
	})

	t.Run("Patch template enabled/disabled", func(t *testing.T) {
		mustPutTemplate(t, store, "b1", "t2", "v1", "T2", false)
		_, err := store.PatchPromptTemplate(t.Context(), &spec.PatchPromptTemplateRequest{
			BundleID:     "b1",
			TemplateSlug: "t2",
			Body:         &spec.PatchPromptTemplateRequestBody{Version: "v1", IsEnabled: true},
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v.", err)
		}
	})

	t.Run("Patch non-existent template", func(t *testing.T) {
		_, err := store.PatchPromptTemplate(t.Context(), &spec.PatchPromptTemplateRequest{
			BundleID:     "b1",
			TemplateSlug: "nope",
			Body:         &spec.PatchPromptTemplateRequestBody{Version: "v1", IsEnabled: false},
		})
		if err == nil || !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("Expected does not exist error, got %v.", err)
		}
	})

	t.Run("Delete template", func(t *testing.T) {
		mustPutTemplate(t, store, "b1", "t3", "v1", "T3", true)
		_, err := store.DeletePromptTemplate(
			t.Context(),
			&spec.DeletePromptTemplateRequest{
				BundleID:     "b1",
				TemplateSlug: "t3",
				Version:      "v1",
			},
		)
		if err != nil {
			t.Fatalf("Unexpected error: %v.", err)
		}
	})

	t.Run("Delete non-existent template", func(t *testing.T) {
		_, err := store.DeletePromptTemplate(
			t.Context(),
			&spec.DeletePromptTemplateRequest{
				BundleID:     "b1",
				TemplateSlug: "nope",
				Version:      "v1",
			},
		)
		if err == nil || !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("Expected does not exist error, got %v.", err)
		}
	})

	t.Run("Delete with invalid slug/version", func(t *testing.T) {
		_, err := store.DeletePromptTemplate(
			t.Context(),
			&spec.DeletePromptTemplateRequest{
				BundleID:     "b1",
				TemplateSlug: "bad.slug",
				Version:      "v1",
			},
		)
		if err == nil || !strings.Contains(err.Error(), "invalid slug") {
			t.Errorf("Expected invalid slug error, got %v.", err)
		}
		_, err = store.DeletePromptTemplate(
			t.Context(),
			&spec.DeletePromptTemplateRequest{
				BundleID:     "b1",
				TemplateSlug: "t1",
				Version:      "bad.version",
			},
		)
		if err == nil || !strings.Contains(err.Error(), "invalid version") {
			t.Errorf("Expected invalid version error, got %v.", err)
		}
	})
}

func TestTemplateGetListActiveFiltering(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()
	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)
	// Insert multiple versions.
	mustPutTemplate(t, store, "b1", "t1", "v1", "T1", true, "tag1")
	time.Sleep(10 * time.Millisecond)
	mustPutTemplate(t, store, "b1", "t1", "v2", "T1v2", false, "tag2")
	time.Sleep(10 * time.Millisecond)
	mustPutTemplate(t, store, "b1", "t1", "v3", "T1v3", true, "tag1", "tag2")
	mustPutTemplate(t, store, "b1", "t2", "v1", "T2", true, "tag2")
	mustPutTemplate(t, store, "b1", "t4", "v10", "T4v10", false, "tag2")

	t.Run("Get by slug+version", func(t *testing.T) {
		resp, err := store.GetPromptTemplate(t.Context(), &spec.GetPromptTemplateRequest{
			BundleID:     "b1",
			TemplateSlug: "t1",
			Version:      "v2",
		})
		if err != nil || resp.Body == nil || resp.Body.Version != "v2" {
			t.Fatalf("Expected v2, got %v, err=%v.", resp.Body, err)
		}
	})

	t.Run("Get active version", func(t *testing.T) {
		resp, err := store.GetPromptTemplate(t.Context(), &spec.GetPromptTemplateRequest{
			BundleID:     "b1",
			TemplateSlug: "t1",
		})
		if err != nil || resp.Body == nil || resp.Body.Version != "v3" {
			t.Fatalf("Expected v3, got %v, err=%v.", resp.Body, err)
		}
	})

	t.Run("Get non-existent", func(t *testing.T) {
		_, err := store.GetPromptTemplate(t.Context(), &spec.GetPromptTemplateRequest{
			BundleID:     "b1",
			TemplateSlug: "nope",
			Version:      "v1",
		})
		if err == nil {
			t.Error("Expected error for non-existent template.")
		}
	})

	t.Run("List all templates", func(t *testing.T) {
		resp, err := store.ListPromptTemplates(
			t.Context(),
			&spec.ListPromptTemplatesRequest{},
		)
		if err != nil {
			t.Fatalf("ListPromptTemplates failed: %v.", err)
		}
		if len(resp.Body.PromptTemplateListItems) < 2 {
			t.Errorf(
				"Expected at least 2 templates, got %d.",
				len(resp.Body.PromptTemplateListItems),
			)
		}
	})

	t.Run("List with tag filter", func(t *testing.T) {
		resp, err := store.ListPromptTemplates(
			t.Context(),
			&spec.ListPromptTemplatesRequest{
				Tags: []string{"tag1"},
			},
		)
		if err != nil {
			t.Fatalf("ListPromptTemplates failed: %v.", err)
		}
		for _, it := range resp.Body.PromptTemplateListItems {
			gt, err := store.GetPromptTemplate(t.Context(), &spec.GetPromptTemplateRequest{
				BundleID:     it.BundleID,
				TemplateSlug: it.TemplateSlug,
				Version:      it.TemplateVersion,
			})
			if err != nil {
				t.Fatalf("GetPromptTemplate failed: %v.", err)
			}
			found := false
			for _, tag := range gt.Body.Tags {
				if tag == "tag1" {
					found = true
				}
			}
			if !found {
				t.Errorf("Expected tag1 in tags, got %v.", gt.Body.Tags)
			}
		}
	})

	t.Run("List with allVersions", func(t *testing.T) {
		resp, err := store.ListPromptTemplates(
			t.Context(),
			&spec.ListPromptTemplatesRequest{
				AllVersions:     true,
				IncludeDisabled: true,
			},
		)
		if err != nil {
			t.Fatalf("ListPromptTemplates failed: %v.", err)
		}
		count := 0
		for _, it := range resp.Body.PromptTemplateListItems {
			if it.TemplateSlug == "t1" {
				count++
			}
		}
		if count != 3 {
			t.Errorf("Expected 3 versions of t1, got %d.", count)
		}
	})

	t.Run("List with includeDisabled", func(t *testing.T) {
		resp, err := store.ListPromptTemplates(
			t.Context(),
			&spec.ListPromptTemplatesRequest{
				IncludeDisabled: true,
			},
		)
		if err != nil {
			t.Fatalf("ListPromptTemplates failed: %v.", err)
		}
		found := false
		for _, it := range resp.Body.PromptTemplateListItems {
			if it.TemplateSlug == "t4" && it.TemplateVersion == "v10" {
				found = true
			}
		}
		if !found {
			t.Errorf("Expected to find disabled v10.")
		}
	})

	t.Run("List with bundleIDs filter", func(t *testing.T) {
		resp, err := store.ListPromptTemplates(
			t.Context(),
			&spec.ListPromptTemplatesRequest{
				BundleIDs: []string{"b1"},
			},
		)
		if err != nil {
			t.Fatalf("ListPromptTemplates failed: %v.", err)
		}
		for _, it := range resp.Body.PromptTemplateListItems {
			if it.BundleID != "b1" {
				t.Errorf("Expected only b1, got %v.", it.BundleID)
			}
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		// Ensure at least 3 templates exist for pagination to be meaningful.
		mustPutTemplate(t, store, "b1", "t3", "v1", "T3", true)
		resp, err := store.ListPromptTemplates(
			t.Context(),
			&spec.ListPromptTemplatesRequest{PageSize: 2},
		)
		if err != nil {
			t.Fatal(err)
		}
		if len(resp.Body.PromptTemplateListItems) != 2 || resp.Body.NextPageToken == nil {
			t.Errorf("Expected 2 templates and nextPageToken, got %v.", resp.Body)
		}
		// Fetch next page.
		resp2, err := store.ListPromptTemplates(
			t.Context(),
			&spec.ListPromptTemplatesRequest{PageSize: 2, PageToken: *resp.Body.NextPageToken},
		)
		if err != nil {
			t.Fatal(err)
		}
		if len(resp2.Body.PromptTemplateListItems) == 0 {
			t.Errorf("Expected more templates in next page.")
		}
	})
}

func TestSoftAndHardDeleteLogic(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()
	mustPutBundle(t, store, "b1", "slug1", "Bundle 1", true)
	_, err := store.DeletePromptBundle(
		t.Context(),
		&spec.DeletePromptBundleRequest{BundleID: "b1"},
	)
	if err != nil {
		t.Fatalf("DeletePromptBundle failed: %v.", err)
	}
	// Simulate grace period passed.
	b, _ := store.bundleStore.GetKey([]string{"bundles", "b1"})
	if mp, ok := b.(map[string]any); ok {
		mp["softDeletedAt"] = time.Now().Add(-2 * softDeleteGrace).UTC().Format(time.RFC3339Nano)
		_ = store.bundleStore.SetKey([]string{"bundles", "b1"}, mp)
	}
	store.sweepSoftDeleted()
	_, err = store.bundleStore.GetKey([]string{"bundles", "b1"})
	if err == nil {
		t.Error("Expected bundle to be hard deleted.")
	}
}

func TestSoftDeleteNotHardIfNotEmpty(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()
	mustPutBundle(t, store, "b2", "slug2", "Bundle 2", true)
	mustPutTemplate(t, store, "b2", "t1", "v1", "T1", true)
	_, err := store.DeletePromptBundle(
		t.Context(),
		&spec.DeletePromptBundleRequest{BundleID: "b2"},
	)
	if err == nil || !strings.Contains(err.Error(), "still contains templates") {
		t.Fatalf("Expected error for non-empty bundle, got %v.", err)
	}
	// The bundle should still exist.
	_, err = store.bundleStore.GetKey([]string{"bundles", "b2"})
	if err != nil {
		t.Error("Expected bundle to remain after failed delete.")
	}
}

func TestSlugVersionValidation(t *testing.T) {
	cases := []struct {
		slug, version string
		valid         bool
	}{
		{"abc", "v1", true},
		{"a-b-c", "v-1", true},
		{"abc123", "v1", true},
		{"", "v1", false},
		{"abc", "", false},
		{"abc!", "v1", false},
		{"abc", "v1!", false},
		{strings.Repeat("a", 65), "v1", false},
		{"abc", strings.Repeat("v", 65), false},
		{"ABC", "V1", true},
		{"abc-def", "v-1", true},
		{"abc_def", "v1", false},
		{"abc.def", "v1", false},
		{"abc def", "v1", false},
		{"abc/v1", "v1", false},
	}
	for _, c := range cases {
		err1 := ValidateSlug(c.slug)
		err2 := ValidateVersion(c.version)
		if c.valid && (err1 != nil || err2 != nil) {
			t.Errorf(
				"Expected valid for slug=%q version=%q, got err1=%v err2=%v.",
				c.slug,
				c.version,
				err1,
				err2,
			)
		}
		if !c.valid && (err1 == nil && err2 == nil) {
			t.Errorf("Expected invalid for slug=%q version=%q.", c.slug, c.version)
		}
	}
}
