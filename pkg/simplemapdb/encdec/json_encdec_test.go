package encdec

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestJSONEncoderDecoder_Encode(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    string
		wantErr bool
	}{
		{
			name:    "encode simple map",
			value:   map[string]string{"key": "value"},
			want:    "{\n  \"key\": \"value\"\n}\n",
			wantErr: false,
		},
		{
			name:    "encode nil value",
			value:   nil,
			want:    "null\n",
			wantErr: false,
		},
		{
			name:    "encode unsupported type",
			value:   make(chan int),
			want:    "",
			wantErr: true,
		},
		{
			name:    "encode struct",
			value:   struct{ Name string }{Name: "Alice"},
			want:    "{\n  \"Name\": \"Alice\"\n}\n",
			wantErr: false,
		},
		{
			name:    "encode slice",
			value:   []int{1, 2, 3},
			want:    "[\n  1,\n  2,\n  3\n]\n",
			wantErr: false,
		},
		{
			name: "complex nested structure",
			value: map[string]any{
				"nested": map[string]any{
					"key": "value",
				},
			},
			want:    "{\n  \"nested\": {\n    \"key\": \"value\"\n  }\n}\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			encoder := JSONEncoderDecoder{}
			err := encoder.Encode(&buf, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONEncoderDecoder_Decode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		value   any
		want    any
		wantErr bool
	}{
		{
			name:    "decode simple map",
			input:   "{\n  \"key\": \"value\"\n}",
			value:   &map[string]string{},
			want:    &map[string]string{"key": "value"},
			wantErr: false,
		},
		{
			name:    "decode invalid JSON",
			input:   "{key: value}",
			value:   &map[string]string{},
			want:    &map[string]string{},
			wantErr: true,
		},
		{
			name:    "decode into nil",
			input:   "{}",
			value:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "decode empty JSON",
			input:   "",
			value:   &map[string]string{},
			want:    &map[string]string{},
			wantErr: true,
		},
		{
			name:    "decode into struct",
			input:   "{\n  \"Name\": \"Alice\"\n}",
			value:   &struct{ Name string }{},
			want:    &struct{ Name string }{Name: "Alice"},
			wantErr: false,
		},
		{
			name:    "decode into slice",
			input:   "[1, 2, 3]",
			value:   &[]int{},
			want:    &[]int{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "nil reader",
			input:   "",
			value:   &map[string]string{},
			want:    &map[string]string{},
			wantErr: true,
		},
		{
			name:    "non-pointer value",
			input:   "{\n  \"key\": \"value\"\n}",
			value:   map[string]string{},
			want:    map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := JSONEncoderDecoder{}
			err := decoder.Decode(strings.NewReader(tt.input), tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.value, tt.want) {
				t.Errorf("Decode() = %v, want %v", tt.value, tt.want)
			}
		})
	}
}

func TestStructWithJSONTagsToMap(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email,omitempty"`
	}

	tests := []struct {
		name    string
		input   any
		want    map[string]any
		wantErr bool
	}{
		{
			name: "happy path",
			input: TestStruct{
				Name:  "John Doe",
				Age:   30,
				Email: "john.doe@example.com",
			},
			want: map[string]any{
				"name":  "John Doe",
				"age":   float64(30),
				"email": "john.doe@example.com",
			},
			wantErr: false,
		},
		{
			name: "empty struct",
			input: TestStruct{
				Name:  "",
				Age:   0,
				Email: "",
			},
			want: map[string]any{
				"name": "",
				"age":  float64(0),
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name: "struct with nested struct",
			input: struct {
				Person TestStruct `json:"person"`
			}{
				Person: TestStruct{
					Name:  "Jane Doe",
					Age:   25,
					Email: "jane.doe@example.com",
				},
			},
			want: map[string]any{
				"person": map[string]any{
					"name":  "Jane Doe",
					"age":   float64(25),
					"email": "jane.doe@example.com",
				},
			},
			wantErr: false,
		},
		{
			name:    "unsupported type",
			input:   func() {},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StructWithJSONTagsToMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("StructWithJSONTagsToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StructWithJSONTagsToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapToStructWithJSONTags(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email,omitempty"`
	}

	tests := []struct {
		name    string
		input   map[string]any
		output  any
		want    TestStruct
		wantErr bool
	}{
		{
			name: "happy path",
			input: map[string]any{
				"name":  "John Doe",
				"age":   float64(30),
				"email": "john.doe@example.com",
			},
			output: &TestStruct{},
			want: TestStruct{
				Name:  "John Doe",
				Age:   30,
				Email: "john.doe@example.com",
			},
			wantErr: false,
		},
		{
			name: "missing optional field",
			input: map[string]any{
				"name": "Jane Doe",
				"age":  float64(25),
			},
			output: &TestStruct{},
			want: TestStruct{
				Name:  "Jane Doe",
				Age:   25,
				Email: "",
			},
			wantErr: false,
		},
		{
			name:    "nil map",
			input:   nil,
			output:  &TestStruct{},
			want:    TestStruct{},
			wantErr: true,
		},
		{
			name: "non-pointer output",
			input: map[string]any{
				"name": "John Doe",
				"age":  float64(30),
			},
			output:  TestStruct{},
			want:    TestStruct{},
			wantErr: true,
		},
		{
			name: "nil pointer output",
			input: map[string]any{
				"name": "John Doe",
				"age":  float64(30),
			},
			output:  (*TestStruct)(nil),
			want:    TestStruct{},
			wantErr: true,
		},
		{
			name: "extra fields in map",
			input: map[string]any{
				"name":    "John Doe",
				"age":     float64(30),
				"email":   "john.doe@example.com",
				"address": "123 Main St",
			},
			output: &TestStruct{},
			want: TestStruct{
				Name:  "John Doe",
				Age:   30,
				Email: "john.doe@example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid JSON structure",
			input: map[string]any{
				"name": "John Doe",
				"age":  "not a number",
			},
			output:  &TestStruct{},
			want:    TestStruct{},
			wantErr: true,
		},
		{
			name: "incompatible types",
			input: map[string]any{
				"name": 123,
				"age":  "thirty",
			},
			output:  &TestStruct{},
			want:    TestStruct{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MapToStructWithJSONTags(tt.input, tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapToStructWithJSONTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(tt.output, &tt.want) {
				t.Errorf("MapToStructWithJSONTags() = %v, want %v", tt.output, &tt.want)
			}
		})
	}
}
