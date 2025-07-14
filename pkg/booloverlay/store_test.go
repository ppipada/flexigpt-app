package booloverlay

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

type BundleID string

func (BundleID) Group() string { return "bundles" }
func (b BundleID) ID() string  { return string(b) }

type TemplateID string

func (TemplateID) Group() string { return "templates" }
func (t TemplateID) ID() string  { return string(t) }

type OtherID string

func (OtherID) Group() string { return "other" }
func (o OtherID) ID() string  { return string(o) }

type CompositeKey struct{ A, B string }

func (CompositeKey) Group() string { return "composite" }
func (c CompositeKey) ID() string  { return c.A + "::" + c.B }

type DuplicateBundleID string

func (DuplicateBundleID) Group() string { return "bundles" }
func (d DuplicateBundleID) ID() string  { return string(d) }

func tmpStore(t *testing.T, opts ...Option) (st *Store, path string) {
	t.Helper()
	dir := t.TempDir()
	path = filepath.Join(dir, "overlay.json")
	st, err := NewStore(path, opts...)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	return st, path
}

func readFileJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	return obj
}

func contains(err error, substr string) bool {
	return err != nil && substr != "" && (len(err.Error()) >= len(substr)) && (func() bool {
		for i := range err.Error() {
			if len(err.Error())-i < len(substr) {
				return false
			}
			j := 0
			for ; j < len(substr) && err.Error()[i+j] == substr[j]; j++ {
			}
			if j == len(substr) {
				return true
			}
		}
		return false
	})()
}

func TestRegistrationAndGroupCreation(t *testing.T) {
	st, path := tmpStore(t, WithKeyType[BundleID](), WithKeyType[TemplateID]())
	obj := readFileJSON(t, path)
	if _, ok := obj["bundles"]; !ok {
		t.Fatalf("expected group %q to exist in file", "bundles")
	}
	if _, ok := obj["templates"]; !ok {
		t.Fatalf("expected group %q to exist in file", "templates")
	}
	if _, ok := obj["other"]; ok {
		t.Fatalf("group %q must NOT exist (was never registered)", "other")
	}
	// Should not panic or error if group already present.
	if err := st.SetEnabled(BundleID("foo"), true); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}
}

func TestDuplicateRegistrationFails(t *testing.T) {
	_, err := NewStore(t.TempDir()+"/o.json",
		WithKeyType[BundleID](),
		WithKeyType[DuplicateBundleID](),
	)
	if err == nil {
		t.Fatal("expected duplicate registration to error")
	}
}

func TestNoRegistrationMeansNoGroups(t *testing.T) {
	_, path := tmpStore(t /* no opts */)
	obj := readFileJSON(t, path)
	if len(obj) != 0 {
		t.Fatalf("expected no groups, got %v", obj)
	}
}

func TestCRUDAndBoundaries(t *testing.T) {
	st, _ := tmpStore(
		t,
		WithKeyType[BundleID](),
		WithKeyType[TemplateID](),
		WithKeyType[CompositeKey](),
	)

	type testCase struct {
		name string
		key  Key
		def  bool
		set  []bool // sequence of values to set
	}
	tests := []testCase{
		{"bundle/normal", BundleID("core"), false, []bool{true, false, true}},
		{"bundle/empty", BundleID(""), true, []bool{false, true}},
		{"template/empty", TemplateID(""), false, []bool{true}},
		{"template/zero", TemplateID(""), false, []bool{}},
		{"composite/zero", CompositeKey{}, true, []bool{false}},
		{"composite/long", CompositeKey{"a", string(make([]byte, 1000))}, false, []bool{true}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Entry absent → get default.
			got, err := st.IsEnabled(tc.key, tc.def)
			if err != nil {
				t.Fatalf("IsEnabled 1: %v", err)
			}
			if got != tc.def {
				t.Fatalf("initial IsEnabled = %v, want %v", got, tc.def)
			}

			// Set values in sequence, check after each.
			for i, v := range tc.set {
				if err := st.SetEnabled(tc.key, v); err != nil {
					t.Fatalf("SetEnabled #%d: %v", i, err)
				}
				got, err := st.IsEnabled(tc.key, !v)
				if err != nil {
					t.Fatalf("IsEnabled #%d: %v", i, err)
				}
				if got != v {
					t.Fatalf("after set #%d IsEnabled = %v, want %v", i, got, v)
				}
			}

			if err := st.Delete(tc.key); err != nil {
				t.Fatalf("Delete: %v", err)
			}
			got, err = st.IsEnabled(tc.key, tc.def)
			if err != nil {
				t.Fatalf("IsEnabled after delete: %v", err)
			}
			if got != tc.def {
				t.Fatalf("after delete IsEnabled = %v, want default %v", got, tc.def)
			}
		})
	}
}

func TestUnregisteredKeyErrors(t *testing.T) {
	st, _ := tmpStore(t, WithKeyType[BundleID]())
	wantErr := "not registered"
	if err := st.SetEnabled(TemplateID("x"), true); err == nil || !contains(err, wantErr) {
		t.Fatalf("SetEnabled with unregistered key should fail, got %v", err)
	}
	_, err := st.IsEnabled(TemplateID("x"), false)
	if err == nil || !contains(err, wantErr) {
		t.Fatalf("IsEnabled with unregistered key should fail, got %v", err)
	}
	if err := st.Delete(TemplateID("x")); err == nil || !contains(err, wantErr) {
		t.Fatalf("Delete with unregistered key should fail, got %v", err)
	}
}

func TestGroupMissingFromFile(t *testing.T) {
	// Simulate a file with only "bundles" group, but register "templates" too.
	_, path := tmpStore(t, WithKeyType[BundleID]())
	// Remove "templates" group if present.
	obj := readFileJSON(t, path)
	delete(obj, "templates")
	data, _ := json.Marshal(obj)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("inject: %v", err)
	}
	// Now register "templates" and ensure it is created.
	st2, err := NewStore(path, WithKeyType[BundleID](), WithKeyType[TemplateID]())
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if err := st2.SetEnabled(TemplateID("foo"), true); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}
	obj2 := readFileJSON(t, path)
	if _, ok := obj2["templates"]; !ok {
		t.Fatalf("expected group %q to exist in file after SetEnabled", "templates")
	}
}

func TestPersistenceAcrossReopen(t *testing.T) {
	st, path := tmpStore(t, WithKeyType[BundleID]())
	if err := st.SetEnabled(BundleID("persist"), true); err != nil {
		t.Fatalf("set: %v", err)
	}

	st2, err := NewStore(path, WithKeyType[BundleID]())
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	ok, err := st2.IsEnabled(BundleID("persist"), false)
	if err != nil {
		t.Fatalf("get after reopen: %v", err)
	}
	if !ok {
		t.Fatal("value lost after reopen")
	}
}

func TestUnknownGroupIsNotPreserved(t *testing.T) {
	st, path := tmpStore(t, WithKeyType[BundleID]())
	// Inject a foreign group directly into the JSON file.
	obj := readFileJSON(t, path)
	obj["foreign"] = map[string]any{"foo": true}
	data, _ := json.Marshal(obj)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("inject: %v", err)
	}
	// Normal operation on a registered group.
	if err := st.SetEnabled(BundleID("core"), true); err != nil {
		t.Fatalf("set: %v", err)
	}
	// Foreign group must still be there, unchanged.
	obj2 := readFileJSON(t, path)
	if _, ok := obj2["foreign"].(map[string]any); ok {
		t.Fatal("foreign group was not deleted")
	}
}

func TestConcurrentAccess(t *testing.T) {
	const n = 100
	st, _ := tmpStore(t, WithKeyType[BundleID]())
	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func() {
			defer wg.Done()
			k := BundleID("id-" + BundleID(strconv.Itoa(i)).ID())
			val := i%2 == 0
			if err := st.SetEnabled(k, val); err != nil {
				t.Errorf("set: %v", err)
			}
			got, err := st.IsEnabled(k, !val)
			if err != nil {
				t.Errorf("get: %v", err)
			}
			if got != val {
				t.Errorf("concurrent mismatch: key=%v want=%v got=%v", k, val, got)
			}
		}()
	}
	wg.Wait()
}

func TestFileInitializationEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "overlay.json")
	// File does not exist yet.
	st, err := NewStore(path, WithKeyType[BundleID]())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	// Should be able to set/get.
	if err := st.SetEnabled(BundleID("foo"), true); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}
	ok, err := st.IsEnabled(BundleID("foo"), false)
	if err != nil || !ok {
		t.Fatalf("IsEnabled: %v %v", ok, err)
	}
}

func TestFileInitializationPrepopulated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "overlay.json")
	// Pre-populate with some JSON.
	obj := map[string]any{
		"bundles": map[string]any{"foo": true},
		"foreign": map[string]any{"bar": false},
	}
	data, _ := json.Marshal(obj)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	st, err := NewStore(path, WithKeyType[BundleID]())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ok, err := st.IsEnabled(BundleID("foo"), false)
	if err != nil || !ok {
		t.Fatalf("IsEnabled: %v %v", ok, err)
	}
}

func TestDefensiveEmptyKey(t *testing.T) {
	st, _ := tmpStore(t, WithKeyType[BundleID]())
	// Empty key is just BundleID("").
	if err := st.SetEnabled(BundleID(""), true); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}
	ok, err := st.IsEnabled(BundleID(""), false)
	if err != nil || !ok {
		t.Fatalf("IsEnabled: %v %v", ok, err)
	}
	if err := st.Delete(BundleID("")); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	ok, err = st.IsEnabled(BundleID(""), false)
	if err != nil || ok {
		t.Fatalf("IsEnabled after delete: %v %v", ok, err)
	}
}

func TestDefensiveNilStore(t *testing.T) {
	// Not possible to call methods on nil Store (would panic), so skip.
}

func TestToggleBackAndForth(t *testing.T) {
	st, _ := tmpStore(t, WithKeyType[BundleID]())
	k := BundleID("toggle")
	for i := range 10 {
		val := i%2 == 0
		if err := st.SetEnabled(k, val); err != nil {
			t.Fatalf("SetEnabled: %v", err)
		}
		got, err := st.IsEnabled(k, !val)
		if err != nil {
			t.Fatalf("IsEnabled: %v", err)
		}
		if got != val {
			t.Fatalf("toggle #%d: got %v want %v", i, got, val)
		}
	}
}

func TestLargeKey(t *testing.T) {
	st, _ := tmpStore(t, WithKeyType[BundleID]())
	large := BundleID(string(make([]byte, 4096)))
	if err := st.SetEnabled(large, true); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}
	ok, err := st.IsEnabled(large, false)
	if err != nil || !ok {
		t.Fatalf("IsEnabled: %v %v", ok, err)
	}
}

func TestFilePermissionError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "overlay.json")
	// Create a file with no write permission.
	if err := os.WriteFile(path, []byte("{}"), 0o400); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := NewStore(path, WithKeyType[BundleID]())
	if err == nil {
		t.Fatalf("NewStore: did not get error in new store creation")
	}
}

func TestCorruptedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "overlay.json")
	// Write invalid JSON.
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := NewStore(path, WithKeyType[BundleID]())
	if err == nil {
		t.Fatal("expected NewStore to fail on corrupted file")
	}
}

func TestMissingDirectory(t *testing.T) {
	path := filepath.Join("nonexistent", "overlay.json")
	_, err := NewStore(path, WithKeyType[BundleID]())
	if err == nil {
		t.Fatal("expected NewStore to fail on missing directory")
	}
}

func TestNilKeyImpossible(t *testing.T) {
	// Not possible in Go, as Key is an interface and nil cannot be passed as a value of a concrete type.
	// This test is a placeholder for completeness.
}

func TestSetAfterDelete(t *testing.T) {
	st, _ := tmpStore(t, WithKeyType[BundleID]())
	k := BundleID("foo")
	if err := st.SetEnabled(k, true); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}
	if err := st.Delete(k); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := st.SetEnabled(k, false); err != nil {
		t.Fatalf("SetEnabled after delete: %v", err)
	}
	ok, err := st.IsEnabled(k, true)
	if err != nil || ok {
		t.Fatalf("IsEnabled after re-set: %v %v", ok, err)
	}
}
