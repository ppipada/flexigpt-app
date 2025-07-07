package store

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
)

func drainErrors(t *testing.T, ch <-chan error) {
	t.Helper()
	for err := range ch {
		if err != nil {
			t.Errorf("scale-test error: %v", err)
		}
	}
}

// collectTemplates pages through ListPromptTemplates until exhaustion and
// returns the collected list-items.
func collectTemplates(
	ctx context.Context,
	s *PromptTemplateStore,
	baseReq spec.ListPromptTemplatesRequest,
	pageSize int,
) ([]spec.PromptTemplateListItem, error) {
	var (
		all   []spec.PromptTemplateListItem
		token string
	)

	for {
		req := baseReq          // copy
		req.PageSize = pageSize // ensure deterministic page size
		req.PageToken = token

		resp, err := s.ListPromptTemplates(ctx, &req)
		if err != nil {
			return nil, err
		}
		all = append(all, resp.Body.PromptTemplateListItems...)

		if resp.Body.NextPageToken == nil || *resp.Body.NextPageToken == "" {
			break
		}
		token = *resp.Body.NextPageToken
	}
	return all, nil
}

// uniqCntActive returns the number of unique (bundleID|slug) combinations.
func uniqCntActive(list []spec.PromptTemplateListItem) int {
	m := map[string]struct{}{}
	for _, it := range list {
		key := it.BundleID + "|" + it.TemplateSlug
		m[key] = struct{}{}
	}
	return len(m)
}

// uniqCntAll returns the number of unique (bundleID|slug|version) triples.
func uniqCntAll(list []spec.PromptTemplateListItem) int {
	m := map[string]struct{}{}
	for _, it := range list {
		key := it.BundleID + "|" + it.TemplateSlug + "|" + it.TemplateVersion
		m[key] = struct{}{}
	}
	return len(m)
}

func TestScale_LotsOfBundles(t *testing.T) {
	const (
		nBundles      = 300
		nonEmptyEvery = 10 // every 10-th bundle -> try to add one template
		deleteEvery   = 5  // try to delete every 5-th bundle
		patchEvery    = 2  // toggle enabled flag every 2-nd bundle
		concurrency   = 20
	)

	s, cleanup := newTestStore(t)
	defer cleanup()

	ctx := t.Context()
	errCh := make(chan error, 8_000)

	//    1. CREATE all bundles  (some will get exactly one template).
	var wg sync.WaitGroup
	wg.Add(concurrency)

	createJobs := make(chan int, nBundles)
	for i := 0; i < nBundles; i++ {
		createJobs <- i
	}
	close(createJobs)

	for g := 0; g < concurrency; g++ {
		go func() {
			defer wg.Done()
			for i := range createJobs {
				bID := "bundle-" + strconv.Itoa(i)

				_, err := s.PutPromptBundle(ctx, &spec.PutPromptBundleRequest{
					BundleID: bID,
					Body: &spec.PutPromptBundleRequestBody{
						Slug:        "slug-" + strconv.Itoa(i),
						DisplayName: "Bundle " + strconv.Itoa(i),
						IsEnabled:   i%3 == 0, // enabled for 0,3,6,…
					},
				})
				if err != nil {
					errCh <- err
					continue
				}

				// Try to add a template – this succeeds only while the
				//   bundle is ENABLED.  Ignore ErrBundleDisabled because
				//   in that case the bundle legitimately stays empty.
				if i%nonEmptyEvery == 0 {
					_, err = s.PutPromptTemplate(ctx, &spec.PutPromptTemplateRequest{
						BundleID:     bID,
						TemplateSlug: "t",
						Body: &spec.PutPromptTemplateRequestBody{
							DisplayName: "dummy",
							IsEnabled:   true,
							Version:     "v1",
							Blocks: []spec.MessageBlock{{
								ID:      "1",
								Role:    spec.User,
								Content: "hi",
							}},
						},
					})
					if err != nil && !errors.Is(err, ErrBundleDisabled) {
						errCh <- err
					}
				}
			}
		}()
	}
	wg.Wait()

	//   2. Concurrent PATCH (toggle) + DELETE.
	wg.Add(concurrency)

	modJobs := make(chan int, nBundles)
	for i := 0; i < nBundles; i++ {
		modJobs <- i
	}
	close(modJobs)

	for g := 0; g < concurrency; g++ {
		go func() {
			defer wg.Done()
			for i := range modJobs {
				bID := "bundle-" + strconv.Itoa(i)

				// PATCH – invert the enabled flag for every 2-nd bundle.
				if i%patchEvery == 0 {
					_, err := s.PatchPromptBundle(ctx, &spec.PatchPromptBundleRequest{
						BundleID: bID,
						Body: &spec.PatchPromptBundleRequestBody{
							IsEnabled: i%3 != 0, // invert previous state
						},
					})
					if err != nil {
						errCh <- err
					}
				}

				// DELETE – every 5-th bundle.
				if i%deleteEvery == 0 {
					_, err := s.DeletePromptBundle(ctx, &spec.DeletePromptBundleRequest{
						BundleID: bID,
					})

					/* A bundle is *really* non-empty only when
					   - we TRIED to add a template (i%nonEmptyEvery==0) AND
					   - the bundle was enabled at that time (i%3==0)
					   i.e. i divisible by both 10 and 3  → by 30            */
					nonEmpty := (i%nonEmptyEvery == 0) && (i%3 == 0)

					if nonEmpty && err == nil {
						errCh <- fmt.Errorf("expected delete to fail for non-empty bundle %s", bID)
					}
					if !nonEmpty && err != nil {
						errCh <- fmt.Errorf("delete empty bundle %s failed: %w", bID, err)
					}
				}
			}
		}()
	}
	wg.Wait()

	//   3. LIST sanity check  (default list must hide disabled).
	resp, err := s.ListPromptBundles(ctx, &spec.ListPromptBundlesRequest{})
	if err != nil {
		errCh <- err
	} else {
		for _, b := range resp.Body.PromptBundles {
			if !b.IsEnabled {
				errCh <- fmt.Errorf("disabled bundle %s appeared in enabled-only list", b.ID)
			}
		}
	}

	close(errCh)
	drainErrors(t, errCh)
}

// SCALE TEST B –  few bundles, many template versions.
func TestScale_LotsOfTemplates(t *testing.T) {
	const (
		nBundles    = 3
		nSlugs      = 120 // per bundle
		nVersions   = 4
		concurrency = 40
	)

	s, cleanup := newTestStore(t)
	defer cleanup()
	ctx := t.Context()
	errCh := make(chan error, 20_000)

	// 1. CREATE the bundles first.
	for i := 0; i < nBundles; i++ {
		_, err := s.PutPromptBundle(ctx, &spec.PutPromptBundleRequest{
			BundleID: fmt.Sprintf("b%d", i),
			Body: &spec.PutPromptBundleRequestBody{
				Slug:        fmt.Sprintf("slug-b%d", i),
				DisplayName: fmt.Sprintf("Bundle %d", i),
				IsEnabled:   true,
			},
		})
		if err != nil {
			t.Fatalf("Put bundle: %v", err)
		}
	}

	// 2. MASS-INSERT template versions concurrently.
	type job struct{ b, s, v int }
	insertJobs := make(chan job, nBundles*nSlugs*nVersions)
	for b := range nBundles {
		for sIdx := 0; sIdx < nSlugs; sIdx++ {
			for v := 1; v <= nVersions; v++ {
				insertJobs <- job{b, sIdx, v}
			}
		}
	}
	close(insertJobs)

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for g := 0; g < concurrency; g++ {
		go func() {
			defer wg.Done()
			for j := range insertJobs {
				_, err := s.PutPromptTemplate(ctx, &spec.PutPromptTemplateRequest{
					BundleID:     fmt.Sprintf("b%d", j.b),
					TemplateSlug: fmt.Sprintf("tpl-%d", j.s),
					Body: &spec.PutPromptTemplateRequestBody{
						DisplayName: "T",
						IsEnabled:   true,
						Version:     fmt.Sprintf("v%d", j.v),
						Blocks: []spec.MessageBlock{{
							ID:      "1",
							Role:    spec.User,
							Content: "hi",
						}},
					},
				})
				if err != nil {
					errCh <- err
				}
				time.Sleep(500 * time.Microsecond) // keep ModifiedAt monotonic per slug
			}
		}()
	}
	wg.Wait()

	// 3. QUICK LIST – should return latest (enabled) version per slug.
	all, err := collectTemplates(ctx, s, spec.ListPromptTemplatesRequest{}, 100)
	if err != nil {
		errCh <- err
	} else if got, want := uniqCntActive(all), nBundles*nSlugs; got != want {
		errCh <- fmt.Errorf("active list unique-slug count mismatch got=%d want=%d", got, want)
	}

	// 4. Disable latest version of half the slugs, delete v1 of every 3rd.
	modJobs := make(chan job, nBundles*nSlugs)
	for b := 0; b < nBundles; b++ {
		for sIdx := 0; sIdx < nSlugs; sIdx++ {
			if sIdx%2 == 0 { // disable v4
				modJobs <- job{b, sIdx, 4}
			}
			if sIdx%3 == 0 { // delete v1
				modJobs <- job{b, sIdx, 1}
			}
		}
	}
	close(modJobs)

	wg.Add(concurrency)
	for g := 0; g < concurrency; g++ {
		go func() {
			defer wg.Done()
			for j := range modJobs {
				bID := fmt.Sprintf("b%d", j.b)
				slug := fmt.Sprintf("tpl-%d", j.s)

				if j.v == 4 { // PATCH (disable)
					_, err := s.PatchPromptTemplate(ctx, &spec.PatchPromptTemplateRequest{
						BundleID:     bID,
						TemplateSlug: slug,
						Body: &spec.PatchPromptTemplateRequestBody{
							Version:   "v4",
							IsEnabled: false,
						},
					})
					if err != nil {
						errCh <- err
					}
				} else { // DELETE v1
					_, err := s.DeletePromptTemplate(ctx, &spec.DeletePromptTemplateRequest{
						BundleID:     bID,
						TemplateSlug: slug,
						Version:      "v1",
					})
					if err != nil {
						errCh <- err
					}
				}
			}
		}()
	}
	wg.Wait()

	// 5. LIST with allVersions+includeDisabled.
	base := spec.ListPromptTemplatesRequest{
		AllVersions:     true,
		IncludeDisabled: true,
	}
	all, err = collectTemplates(ctx, s, base, 150)
	if err != nil {
		errCh <- err
	} else {
		totalMade := nBundles * nSlugs * nVersions
		deleted := nBundles * (nSlugs / 3) // every 3rd slug v1 deleted
		want := totalMade - deleted
		if got := uniqCntAll(all); got != want {
			errCh <- fmt.Errorf("allVersions unique count mismatch got=%d want=%d", got, want)
		}
	}

	// 6. Pagination sanity (page size 70).
	base = spec.ListPromptTemplatesRequest{PageSize: 70}
	all, err = collectTemplates(ctx, s, base, 70)
	if err != nil {
		errCh <- err
	} else if got, want := uniqCntActive(all), nBundles*nSlugs; got != want {
		errCh <- fmt.Errorf("pagination unique count mismatch got=%d want=%d", got, want)
	}

	// 7. Drain errors.
	close(errCh)
	drainErrors(t, errCh)
}
