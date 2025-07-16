package booloverlay

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

type BundleID string

func (BundleID) Group() GroupID { return "bundles" }
func (b BundleID) ID() KeyID    { return KeyID(b) }

type TemplateID string

func (TemplateID) Group() GroupID { return "templates" }
func (t TemplateID) ID() KeyID    { return KeyID(t) }

type OtherID string

func (OtherID) Group() GroupID { return "other" }
func (o OtherID) ID() KeyID    { return KeyID(o) }

type CompositeKey struct{ A, B string }

func (CompositeKey) Group() GroupID { return "composite" }
func (c CompositeKey) ID() KeyID    { return KeyID(c.A + "::" + c.B) }

type DuplicateBundleID string

func (DuplicateBundleID) Group() GroupID { return "bundles" }
func (d DuplicateBundleID) ID() KeyID    { return KeyID(d) }

func tmpStore(t *testing.T, opts ...Option) (s *Store, path string) {
	t.Helper()
	dir := t.TempDir()
	path = filepath.Join(dir, "overlay.json")
	s, err := NewStore(path, opts...)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	return s, path
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
	return err != nil && substr != "" && (func() bool {
		return (len(err.Error()) >= len(substr)) && (func() bool {
			for i := range err.Error() {
				if len(err.Error())-i < len(substr) {
					return false
				}
				if err.Error()[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		})()
	})()
}

func getEnabled(st *Store, k Key, def bool) (bool, error) {
	f, ok, err := st.GetFlag(k)
	if err != nil {
		return def, err
	}
	if !ok {
		return def, nil
	}
	return f.Enabled, nil
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
	if _, err := st.SetFlag(BundleID("foo"), true); err != nil {
		t.Fatalf("SetFlag: %v", err)
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
		set  []bool
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
			got, err := getEnabled(st, tc.key, tc.def)
			if err != nil {
				t.Fatalf("GetFlag 1: %v", err)
			}
			if got != tc.def {
				t.Fatalf("initial get = %v, want %v", got, tc.def)
			}

			for i, v := range tc.set {
				if _, err := st.SetFlag(tc.key, v); err != nil {
					t.Fatalf("SetFlag #%d: %v", i, err)
				}
				got, err := getEnabled(st, tc.key, !v)
				if err != nil {
					t.Fatalf("GetFlag #%d: %v", i, err)
				}
				if got != v {
					t.Fatalf("after set #%d get = %v, want %v", i, got, v)
				}
			}

			if err := st.Delete(tc.key); err != nil {
				t.Fatalf("Delete: %v", err)
			}
			got, err = getEnabled(st, tc.key, tc.def)
			if err != nil {
				t.Fatalf("GetFlag after delete: %v", err)
			}
			if got != tc.def {
				t.Fatalf("after delete get = %v, want default %v", got, tc.def)
			}
		})
	}
}

func TestUnregisteredKeyErrors(t *testing.T) {
	st, _ := tmpStore(t, WithKeyType[BundleID]())
	wantErr := "not registered"
	if _, err := st.SetFlag(TemplateID("x"), true); err == nil || !contains(err, wantErr) {
		t.Fatalf("SetFlag with unregistered key should fail, got %v", err)
	}
	_, _, err := st.GetFlag(TemplateID("x"))
	if err == nil || !contains(err, wantErr) {
		t.Fatalf("GetFlag with unregistered key should fail, got %v", err)
	}
	if err := st.Delete(TemplateID("x")); err == nil || !contains(err, wantErr) {
		t.Fatalf("Delete with unregistered key should fail, got %v", err)
	}
}

func TestGroupMissingFromFile(t *testing.T) {
	_, path := tmpStore(t, WithKeyType[BundleID]())
	obj := readFileJSON(t, path)
	delete(obj, "templates")
	data, _ := json.Marshal(obj)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("inject: %v", err)
	}
	st2, err := NewStore(path, WithKeyType[BundleID](), WithKeyType[TemplateID]())
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if _, err := st2.SetFlag(TemplateID("foo"), true); err != nil {
		t.Fatalf("SetFlag: %v", err)
	}
	obj2 := readFileJSON(t, path)
	if _, ok := obj2["templates"]; !ok {
		t.Fatalf("expected group %q to exist in file after SetFlag", "templates")
	}
}

func TestPersistenceAcrossReopen(t *testing.T) {
	st, path := tmpStore(t, WithKeyType[BundleID]())
	if _, err := st.SetFlag(BundleID("persist"), true); err != nil {
		t.Fatalf("set: %v", err)
	}

	st2, err := NewStore(path, WithKeyType[BundleID]())
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	flag, ok, err := st2.GetFlag(BundleID("persist"))
	if err != nil {
		t.Fatalf("get after reopen: %v", err)
	}
	if !ok || !flag.Enabled {
		t.Fatal("value lost after reopen")
	}
}

func TestUnknownGroupIsNotPreserved(t *testing.T) {
	st, path := tmpStore(t, WithKeyType[BundleID]())
	// Inject a foreign group with mismatching schema.
	obj := readFileJSON(t, path)
	obj["foreign"] = map[string]any{"foo": map[string]any{"enabled": true}}
	data, _ := json.Marshal(obj)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("inject: %v", err)
	}
	if _, err := st.SetFlag(BundleID("core"), true); err != nil {
		t.Fatalf("set: %v", err)
	}
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
			if _, err := st.SetFlag(k, val); err != nil {
				t.Errorf("set: %v", err)
			}
			got, err := getEnabled(st, k, !val)
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
	st, err := NewStore(path, WithKeyType[BundleID]())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if _, err := st.SetFlag(BundleID("foo"), true); err != nil {
		t.Fatalf("SetFlag: %v", err)
	}
	ok, err := getEnabled(st, BundleID("foo"), false)
	if err != nil || !ok {
		t.Fatalf("GetFlag: %v %v", ok, err)
	}
}

func TestFileInitializationPrepopulated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "overlay.json")
	obj := map[string]any{
		"bundles": map[string]any{"foo": map[string]any{"enabled": true}},
		"foreign": map[string]any{"bar": map[string]any{"enabled": false}},
	}
	data, _ := json.Marshal(obj)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	st, err := NewStore(path, WithKeyType[BundleID]())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ok, err := getEnabled(st, BundleID("foo"), false)
	if err != nil || !ok {
		t.Fatalf("GetFlag: %v %v", ok, err)
	}
}

func TestDefensiveEmptyKey(t *testing.T) {
	st, _ := tmpStore(t, WithKeyType[BundleID]())
	if _, err := st.SetFlag(BundleID(""), true); err != nil {
		t.Fatalf("SetFlag: %v", err)
	}
	ok, err := getEnabled(st, BundleID(""), false)
	if err != nil || !ok {
		t.Fatalf("GetFlag: %v %v", ok, err)
	}
	if err := st.Delete(BundleID("")); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	ok, err = getEnabled(st, BundleID(""), false)
	if err != nil || ok {
		t.Fatalf("GetFlag after delete: %v %v", ok, err)
	}
}

func TestToggleBackAndForth(t *testing.T) {
	st, _ := tmpStore(t, WithKeyType[BundleID]())
	k := BundleID("toggle")
	for i := range 10 {
		val := i%2 == 0
		if _, err := st.SetFlag(k, val); err != nil {
			t.Fatalf("SetFlag: %v", err)
		}
		got, err := getEnabled(st, k, !val)
		if err != nil {
			t.Fatalf("GetFlag: %v", err)
		}
		if got != val {
			t.Fatalf("toggle #%d: got %v want %v", i, got, val)
		}
	}
}

func TestLargeKey(t *testing.T) {
	st, _ := tmpStore(t, WithKeyType[BundleID]())
	large := BundleID(string(make([]byte, 4096)))
	if _, err := st.SetFlag(large, true); err != nil {
		t.Fatalf("SetFlag: %v", err)
	}
	ok, err := getEnabled(st, large, false)
	if err != nil || !ok {
		t.Fatalf("GetFlag: %v %v", ok, err)
	}
}

func TestFilePermissionError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "overlay.json")
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

func TestSetAfterDelete(t *testing.T) {
	st, _ := tmpStore(t, WithKeyType[BundleID]())
	k := BundleID("foo")
	if _, err := st.SetFlag(k, true); err != nil {
		t.Fatalf("SetFlag: %v", err)
	}
	if err := st.Delete(k); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := st.SetFlag(k, false); err != nil {
		t.Fatalf("SetFlag after delete: %v", err)
	}
	ok, err := getEnabled(st, k, true)
	if err != nil || ok {
		t.Fatalf("GetFlag after re-set: %v %v", ok, err)
	}
}

func TestSetFlagTimestamps(t *testing.T) {
	st, _ := tmpStore(t, WithKeyType[BundleID]())
	k := BundleID("ts")
	f1, err := st.SetFlag(k, true)
	if err != nil {
		t.Fatalf("SetFlag: %v", err)
	}
	if !f1.CreatedAt.Equal(f1.ModifiedAt) {
		t.Fatalf("on create timestamps differ: %+v", f1)
	}
	time.Sleep(10 * time.Millisecond)
	f2, err := st.SetFlag(k, false)
	if err != nil {
		t.Fatalf("SetFlag: %v", err)
	}
	if !f2.ModifiedAt.After(f2.CreatedAt) {
		t.Fatalf("modified_at not updated: %+v", f2)
	}
}
