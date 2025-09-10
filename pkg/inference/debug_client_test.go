package inference

import (
	"reflect"
	"testing"
)

// TestFilterSensitiveInfo_NilAndEmptyInput verifies behavior for nil and empty inputs.
func TestFilterSensitiveInfo_NilAndEmptyInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]any
		want map[string]any
	}{
		{
			name: "NilInputReturnsNil.",
			in:   nil,
			want: nil,
		},
		{
			name: "EmptyMapReturnsEmptyMap.",
			in:   map[string]any{},
			want: map[string]any{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := filterSensitiveInfo(tc.in)

			// Nil case: want nil, ensure got is nil.
			if tc.in == nil {
				if got != nil {
					t.Fatalf("got = %#v, want = nil.", got)
				}
				return
			}

			// Empty map case: ensure equality and non-nil map.
			if got == nil {
				t.Fatalf("got = nil, want = %#v.", tc.want)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got = %#v, want = %#v.", got, tc.want)
			}
		})
	}
}

// TestFilterSensitiveInfo_BasicRedaction verifies redaction of sensitive keys at the top level.
func TestFilterSensitiveInfo_BasicRedaction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]any
		want map[string]any
	}{
		{
			name: "NoSensitiveKeysRemainUnchanged.",
			in: map[string]any{
				"Make":  "Honda",
				"Model": "Civic",
				"Year":  2020,
				"New":   true,
			},
			want: map[string]any{
				"Make":  "Honda",
				"Model": "Civic",
				"Year":  2020,
				"New":   true,
			},
		},
		{
			name: "SensitiveKeysAreMaskedCaseInsensitiveAndSubstring.",
			in: map[string]any{
				"Authorization": "Bearer secret",
				"apiKey":        "abc123",
				"monkey":        "banana",
				"turKey":        "sandwich",
				"Key":           "value",
				"VIN":           "JH4DA9350LS000001",
				"Spare":         nil,
			},
			want: map[string]any{
				"Authorization": "***",
				"apiKey":        "***",
				"monkey":        "***",
				"turKey":        "***",
				"Key":           "***",
				"VIN":           "JH4DA9350LS000001",
				"Spare":         nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := filterSensitiveInfo(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got = %#v, want = %#v.", got, tc.want)
			}
		})
	}
}

// TestFilterSensitiveInfo_NestedStructures verifies redaction inside nested maps and slices.
func TestFilterSensitiveInfo_NestedStructures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]any
		want map[string]any
	}{
		{
			name: "NestedMapsAndSlicesAreProcessedRecursively.",
			in: map[string]any{
				"Car": map[string]any{
					"Make":   "Toyota",
					"Model":  "Supra",
					"APIKey": "supersecret",
					"Parts": []any{
						map[string]any{
							"Name":      "Turbo",
							"dealerKey": "should-redact",
						},
						[]any{
							map[string]any{
								"authorization": "abc",
								"Model":         "86",
							},
						},
					},
				},
				"garageKey": 12345,
			},
			want: map[string]any{
				"Car": map[string]any{
					"Make":   "Toyota",
					"Model":  "Supra",
					"APIKey": "***",
					"Parts": []any{
						map[string]any{
							"Name":      "Turbo",
							"dealerKey": "***",
						},
						[]any{
							map[string]any{
								"authorization": "***",
								"Model":         "86",
							},
						},
					},
				},
				"garageKey": "***",
			},
		},
		{
			name: "UnknownTypesArePreservedAsIs.",
			in: map[string]any{
				"Make": "Nissan",
				"Spec": struct {
					Trim string
				}{Trim: "Nismo"},
				"key": "noop",
			},
			want: map[string]any{
				"Make": "Nissan",
				"Spec": struct {
					Trim string
				}{Trim: "Nismo"},
				"key": "***",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := filterSensitiveInfo(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got = %#v, want = %#v.", got, tc.want)
			}
		})
	}
}

// TestFilterSensitiveInfo_Cycles verifies that cyclic references are handled safely.
func TestFilterSensitiveInfo_Cycles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   func() map[string]any
		want map[string]any
	}{
		{
			name: "MapSelfCycleYieldsCycleToken.",
			in: func() map[string]any {
				vehicle := map[string]any{
					"Make": "Toyota",
					"VIN":  "JT2JA82J1R0000001",
				}
				vehicle["Self"] = vehicle
				return map[string]any{"Vehicle": vehicle}
			},
			want: map[string]any{
				"Vehicle": map[string]any{
					"Make": "Toyota",
					"VIN":  "JT2JA82J1R0000001",
					"Self": "<cycle>",
				},
			},
		},
		{
			name: "SliceSelfCycleYieldsCycleToken.",
			in: func() map[string]any {
				garage := []any{"BMW"}
				garage = append(garage, nil)
				garage[1] = garage // Self-reference.
				return map[string]any{"Garage": garage}
			},
			want: map[string]any{
				"Garage": []any{"BMW", "<cycle>"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := filterSensitiveInfo(tc.in())
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got = %#v, want = %#v.", got, tc.want)
			}
		})
	}
}

// TestFilterSensitiveInfo_Immutability verifies that input is not mutated and output is a deep copy.
func TestFilterSensitiveInfo_Immutability(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   func() map[string]any
	}{
		{
			name: "DeepCopyPreventsMutationLeaks.",
			in: func() map[string]any {
				return map[string]any{
					"Car": map[string]any{
						"Make":  "Honda",
						"Model": "Civic",
						"Key":   "topSecret",
					},
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orig := tc.in()
			clean := filterSensitiveInfo(orig)

			// Mutate original after sanitization.
			origCar := orig["Car"].(map[string]any)
			origCar["Model"] = "Accord"
			origCar["Key"] = "changed"

			// Ensure the sanitized copy did not change.
			cleanCar := clean["Car"].(map[string]any)
			if got, want := cleanCar["Model"], "Civic"; got != want {
				t.Fatalf("sanitized copy mutated: Model got = %v, want = %v.", got, want)
			}
			if got, want := cleanCar["Key"], any("***"); got != want {
				t.Fatalf("sanitized copy mutated: Key got = %v, want = %v.", got, want)
			}

			// Mutate the sanitized copy and ensure the original did not change.
			cleanCar["Model"] = "Integra"
			if got, want := origCar["Model"], "Accord"; got != want {
				t.Fatalf(
					"original mutated by changing sanitized copy: got = %v, want = %v.",
					got,
					want,
				)
			}
		})
	}
}
