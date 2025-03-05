package reqresp

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// MyData is a sample data structure for testing.
type MyData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// Test unmarshalBatchItem with Request[json.RawMessage].
func TestUnmarshalBatchItem_Request(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		wantIsBatch bool
		wantItems   []Request[json.RawMessage]
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name:       "Empty input data",
			data:       []byte{},
			wantErr:    true,
			wantErrMsg: "Received empty data",
		},
		{
			name:        "Valid single request",
			data:        []byte(`{"jsonrpc": "2.0", "method": "sum", "params": [1,2,3], "id":1}`),
			wantIsBatch: false,
			wantItems: []Request[json.RawMessage]{
				{
					JSONRPC: "2.0",
					Method:  "sum",
					Params:  json.RawMessage(`[1,2,3]`),
					ID:      RequestID{Value: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid batch requests",
			data: []byte(
				`[{"jsonrpc": "2.0", "method": "sum", "params": [1,2,3], "id":1}, {"jsonrpc": "2.0", "method": "subtract", "params": [42,23], "id":2}]`,
			),
			wantIsBatch: true,
			wantItems: []Request[json.RawMessage]{
				{
					JSONRPC: "2.0",
					Method:  "sum",
					Params:  json.RawMessage(`[1,2,3]`),
					ID:      RequestID{Value: 1},
				},
				{
					JSONRPC: "2.0",
					Method:  "subtract",
					Params:  json.RawMessage(`[42,23]`),
					ID:      RequestID{Value: 2},
				},
			},
			wantErr: false,
		},
		{
			name:       "Invalid JSON",
			data:       []byte(`{this is not valid JSON}`),
			wantErr:    true,
			wantErrMsg: "Failed to unmarshal single item",
		},
		{
			name:        "Empty batch",
			data:        []byte(`[]`),
			wantErr:     false,
			wantIsBatch: true,
			wantItems:   []Request[json.RawMessage]{},
		},
		{
			name:       "Null input",
			data:       []byte(`null`),
			wantErr:    true,
			wantErrMsg: "Received null data",
		},
		{
			name:       "No input",
			data:       []byte(``),
			wantErr:    true,
			wantErrMsg: "Received empty data",
		},
		{
			name:       "Garbage data",
			data:       []byte(`garbage data`),
			wantErr:    true,
			wantErrMsg: "Failed to unmarshal single item",
		},
		{
			name:       "Whitespace input",
			data:       []byte("   "),
			wantErr:    true,
			wantErrMsg: "Received empty data",
		},
		{
			name:       "Only null byte",
			data:       []byte("\x00"),
			wantErr:    true,
			wantErrMsg: "Failed to unmarshal single item",
		},
		{
			name:       "Array with null",
			data:       []byte(`[null]`),
			wantErr:    true,
			wantErrMsg: "Received null data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var isBatch bool
			var items []Request[json.RawMessage]
			err := unmarshalBatchItem(tt.data, &isBatch, &items)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unmarshalBatchItem() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf(
						"unmarshalBatchItem() error message = %v, want %v",
						err.Error(),
						tt.wantErrMsg,
					)
				}
				return
			}
			if isBatch != tt.wantIsBatch {
				t.Errorf("unmarshalBatchItem() isBatch = %v, want %v", isBatch, tt.wantIsBatch)
			}
			if !compareRequestSlices(items, tt.wantItems) {
				t.Errorf("unmarshalBatchItem() items = %+v, want %+v", items, tt.wantItems)
			}
		})
	}
}

// Test marshalBatchItem with Request[json.RawMessage].
func TestMarshalBatchItem_Request(t *testing.T) {
	tests := []struct {
		name       string
		isBatch    bool
		items      []Request[json.RawMessage]
		wantData   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Single item",
			isBatch: false,
			items: []Request[json.RawMessage]{
				{
					JSONRPC: "2.0",
					Method:  "subtract",
					Params:  json.RawMessage(`[42,23]`),
					ID:      RequestID{Value: 1},
				},
			},
			wantData: `{"jsonrpc":"2.0","method":"subtract","params":[42,23],"id":1}`,
			wantErr:  false,
		},
		{
			name:    "Batch items",
			isBatch: true,
			items: []Request[json.RawMessage]{
				{
					JSONRPC: "2.0",
					Method:  "sum",
					Params:  json.RawMessage(`[1,2,3]`),
					ID:      RequestID{Value: 1},
				},
				{
					JSONRPC: "2.0",
					Method:  "subtract",
					Params:  json.RawMessage(`[42,23]`),
					ID:      RequestID{Value: 2},
				},
			},
			wantData: `[{"jsonrpc":"2.0","method":"sum","params":[1,2,3],"id":1},{"jsonrpc":"2.0","method":"subtract","params":[42,23],"id":2}]`,
			wantErr:  false,
		},
		{
			name:     "Empty items with isBatch=false",
			isBatch:  false,
			items:    []Request[json.RawMessage]{},
			wantData: `null`,
			wantErr:  false,
		},
		{
			name:     "Empty items with isBatch=true",
			isBatch:  true,
			items:    []Request[json.RawMessage]{},
			wantData: `[]`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := marshalBatchItem(tt.isBatch, tt.items)
			if (err != nil) != tt.wantErr {
				t.Fatalf("marshalBatchItem() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf(
						"marshalBatchItem() error message = %v, want %v",
						err.Error(),
						tt.wantErrMsg,
					)
				}
				return
			}
			if !jsonStringsEqual(string(data), tt.wantData) {
				t.Errorf("marshalBatchItem() data = %s, want %s", string(data), tt.wantData)
			}
		})
	}
}

// Test BatchItem[T] UnmarshalJSON and MarshalJSON with MyData.
func TestBatchItem_MyData(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		wantIsBatch bool
		wantItems   []MyData
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name:        "Single item",
			jsonData:    `{"name": "Item1", "value": 100}`,
			wantIsBatch: false,
			wantItems: []MyData{
				{Name: "Item1", Value: 100},
			},
			wantErr: false,
		},
		{
			name:        "Batch items",
			jsonData:    `[{"name": "Item1", "value": 100}, {"name": "Item2", "value": 200}]`,
			wantIsBatch: true,
			wantItems: []MyData{
				{Name: "Item1", Value: 100},
				{Name: "Item2", Value: 200},
			},
			wantErr: false,
		},
		{
			name:       "Invalid JSON",
			jsonData:   `{"name": "Item1", "value": 100`,
			wantErr:    true,
			wantErrMsg: "unexpected end of JSON input",
		},
		{
			name:       "Empty input",
			jsonData:   `  `,
			wantErr:    true,
			wantErrMsg: "unexpected end of JSON input",
		},
		{
			name:       "Invalid field type",
			jsonData:   `{"name": "Item1", "value": "one hundred"}`,
			wantErr:    true,
			wantErrMsg: "Failed to unmarshal single item",
		},
		{
			name:        "Empty batch",
			jsonData:    `[]`,
			wantIsBatch: true,
			wantItems:   []MyData{},
			wantErr:     false,
		},
		{
			name:        "Valid empty object",
			jsonData:    `{}`,
			wantIsBatch: false,
			wantItems:   []MyData{{}},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var meta BatchItem[MyData]
			err := json.Unmarshal([]byte(tt.jsonData), &meta)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BatchItem.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf(
						"BatchItem.UnmarshalJSON() error message = %v, want %v",
						err.Error(),
						tt.wantErrMsg,
					)
				}
				return
			}
			if meta.IsBatch != tt.wantIsBatch {
				t.Errorf(
					"BatchItem.UnmarshalJSON() IsBatch = %v, want %v",
					meta.IsBatch,
					tt.wantIsBatch,
				)
			}
			if !reflect.DeepEqual(meta.Items, tt.wantItems) {
				t.Errorf(
					"BatchItem.UnmarshalJSON() Items = %#v, want %#v",
					meta.Items,
					tt.wantItems,
				)
			}
		})
	}
}

func TestBatchItem_MarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		meta       BatchItem[MyData]
		wantData   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "Single item",
			meta: BatchItem[MyData]{
				IsBatch: false,
				Items: []MyData{
					{Name: "Item1", Value: 100},
				},
			},
			wantData: `{"name":"Item1","value":100}`,
			wantErr:  false,
		},
		{
			name: "Batch items",
			meta: BatchItem[MyData]{
				IsBatch: true,
				Items: []MyData{
					{Name: "Item1", Value: 100},
					{Name: "Item2", Value: 200},
				},
			},
			wantData: `[{"name":"Item1","value":100},{"name":"Item2","value":200}]`,
			wantErr:  false,
		},
		{
			name: "Empty items with IsBatch=false",
			meta: BatchItem[MyData]{
				IsBatch: false,
				Items:   []MyData{},
			},
			wantData: `null`,
			wantErr:  false,
		},
		{
			name: "Empty items with IsBatch=true",
			meta: BatchItem[MyData]{
				IsBatch: true,
				Items:   []MyData{},
			},
			wantData: `[]`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(&tt.meta)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BatchItem.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf(
						"BatchItem.MarshalJSON() error message = %v, want %v",
						err.Error(),
						tt.wantErrMsg,
					)
				}
				return
			}
			if !jsonStringsEqual(string(data), tt.wantData) {
				t.Errorf("BatchItem.MarshalJSON() data = %s, want %s", string(data), tt.wantData)
			}
		})
	}
}

// Test BatchRequest UnmarshalJSON and MarshalJSON.
func TestBatchRequest_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		wantIsBatch bool
		wantItems   []UnionRequest
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name:        "Valid single request",
			jsonData:    `{"jsonrpc":"2.0","method":"sum","params":[1,2,3],"id":1}`,
			wantIsBatch: false,
			wantItems: []UnionRequest{
				{
					JSONRPC: "2.0",
					Method:  stringToPointer("sum"),
					Params:  json.RawMessage(`[1,2,3]`),
					ID:      &RequestID{Value: 1},
				},
			},
			wantErr: false,
		},
		{
			name:        "Valid batch requests",
			jsonData:    `[{"jsonrpc":"2.0","method":"sum","params":[1,2,3],"id":1},{"jsonrpc":"2.0","method":"subtract","params":[42,23],"id":2}]`,
			wantIsBatch: true,
			wantItems: []UnionRequest{
				{
					JSONRPC: "2.0",
					Method:  stringToPointer("sum"),
					Params:  json.RawMessage(`[1,2,3]`),
					ID:      &RequestID{Value: 1},
				},
				{
					JSONRPC: "2.0",
					Method:  stringToPointer("subtract"),
					Params:  json.RawMessage(`[42,23]`),
					ID:      &RequestID{Value: 2},
				},
			},
			wantErr: false,
		},
		{
			name:       "Empty input",
			jsonData:   ``,
			wantErr:    true,
			wantErrMsg: "unexpected end of JSON input",
		},
		{
			name:       "Null input",
			jsonData:   `null`,
			wantErr:    true,
			wantErrMsg: "Received null data",
		},
		{
			name:        "Empty batch",
			jsonData:    `[]`,
			wantIsBatch: true,
			wantItems:   []UnionRequest{},
			wantErr:     false,
		},
		{
			name:       "Invalid JSON",
			jsonData:   `{this is not valid JSON}`,
			wantErr:    true,
			wantErrMsg: "invalid character 't' looking for beginning of object key string",
		},
		{
			name:       "Array with null",
			jsonData:   `[null]`,
			wantErr:    true,
			wantErrMsg: "Received null data",
		},
		{
			name:       "Whitespace input",
			jsonData:   "   ",
			wantErr:    true,
			wantErrMsg: "unexpected end of JSON input",
		},
		{
			name:       "Garbage data",
			jsonData:   `garbage data`,
			wantErr:    true,
			wantErrMsg: "invalid character 'g' looking for beginning of value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metaRequest BatchRequest
			metaRequest.Body = &BatchItem[UnionRequest]{}
			err := json.Unmarshal([]byte(tt.jsonData), metaRequest.Body)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BatchRequest.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf(
						"BatchRequest.UnmarshalJSON() error message = %v, want %v",
						err.Error(),
						tt.wantErrMsg,
					)
				}
				return
			}
			if metaRequest.Body.IsBatch != tt.wantIsBatch {
				t.Errorf(
					"BatchRequest.UnmarshalJSON() IsBatch = %v, want %v",
					metaRequest.Body.IsBatch,
					tt.wantIsBatch,
				)
			}
			if !compareUnionRequestSlices(metaRequest.Body.Items, tt.wantItems) {
				t.Errorf(
					"BatchRequest.UnmarshalJSON() Items = %+v, want %+v",
					metaRequest.Body.Items,
					tt.wantItems,
				)
			}
		})
	}
}

func TestBatchRequest_MarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		meta       BatchRequest
		wantData   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "Single request",
			meta: BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{
						{
							JSONRPC: "2.0",
							Method:  stringToPointer("subtract"),
							Params:  json.RawMessage(`[42,23]`),
							ID:      &RequestID{Value: 1},
						},
					},
				},
			},
			wantData: `{"jsonrpc":"2.0","method":"subtract","params":[42,23],"id":1}`,
			wantErr:  false,
		},
		{
			name: "Batch requests",
			meta: BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: true,
					Items: []UnionRequest{
						{
							JSONRPC: "2.0",
							Method:  stringToPointer("sum"),
							Params:  json.RawMessage(`[1,2,3]`),
							ID:      &RequestID{Value: 1},
						},
						{
							JSONRPC: "2.0",
							Method:  stringToPointer("subtract"),
							Params:  json.RawMessage(`[42,23]`),
							ID:      &RequestID{Value: 2},
						},
					},
				},
			},
			wantData: `[{"jsonrpc":"2.0","method":"sum","params":[1,2,3],"id":1},{"jsonrpc":"2.0","method":"subtract","params":[42,23],"id":2}]`,
			wantErr:  false,
		},
		{
			name: "Empty items with IsBatch=false",
			meta: BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items:   []UnionRequest{},
				},
			},
			wantErr:  false,
			wantData: `null`,
		},
		{
			name: "Empty items with IsBatch=true",
			meta: BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: true,
					Items:   []UnionRequest{},
				},
			},
			wantData: `[]`,
			wantErr:  false,
		},
		{
			name: "Nil Body",
			meta: BatchRequest{
				Body: nil,
			},
			wantErr:  false,
			wantData: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.meta.Body)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BatchRequest.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf(
						"BatchRequest.MarshalJSON() error message = %v, want %v",
						err.Error(),
						tt.wantErrMsg,
					)
				}
				return
			}

			if !jsonStringsEqual(string(data), tt.wantData) {
				t.Errorf("BatchRequest.MarshalJSON() data = %s, want %s", string(data), tt.wantData)
			}
		})
	}
}

// Test BatchResponse UnmarshalJSON and MarshalJSON.
func TestBatchResponse_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		wantIsBatch bool
		wantItems   []Response[json.RawMessage]
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name:        "Valid single response",
			jsonData:    `{"jsonrpc":"2.0","result":7,"id":1}`,
			wantIsBatch: false,
			wantItems: []Response[json.RawMessage]{
				{
					JSONRPC: "2.0",
					Result:  json.RawMessage(`7`),
					ID:      &RequestID{Value: 1},
				},
			},
			wantErr: false,
		},
		{
			name:        "Valid batch responses",
			jsonData:    `[{"jsonrpc":"2.0","result":7,"id":1},{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found"},"id":2}]`,
			wantIsBatch: true,
			wantItems: []Response[json.RawMessage]{
				{
					JSONRPC: "2.0",
					Result:  json.RawMessage(`7`),
					ID:      &RequestID{Value: 1},
				},
				{
					JSONRPC: "2.0",
					Error: &JSONRPCError{
						Code:    -32601,
						Message: "Method not found",
					},
					ID: &RequestID{Value: 2},
				},
			},
			wantErr: false,
		},
		{
			name:       "Empty input",
			jsonData:   ``,
			wantErr:    true,
			wantErrMsg: "unexpected end of JSON input",
		},
		{
			name:       "Null input",
			jsonData:   `null`,
			wantErr:    true,
			wantErrMsg: "Received null data",
		},
		{
			name:        "Empty batch",
			jsonData:    `[]`,
			wantIsBatch: true,
			wantItems:   []Response[json.RawMessage]{},
			wantErr:     false,
		},
		{
			name:       "Invalid JSON",
			jsonData:   `{this is not valid JSON}`,
			wantErr:    true,
			wantErrMsg: "invalid character 't' looking for beginning of object key string",
		},
		{
			name:       "Array with null",
			jsonData:   `[null]`,
			wantErr:    true,
			wantErrMsg: "Received null data",
		},
		{
			name:       "Whitespace input",
			jsonData:   "   ",
			wantErr:    true,
			wantErrMsg: "unexpected end of JSON input",
		},
		{
			name:       "Garbage data",
			jsonData:   `garbage data`,
			wantErr:    true,
			wantErrMsg: "invalid character 'g' looking for beginning of value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metaResponse BatchResponse
			metaResponse.Body = &BatchItem[Response[json.RawMessage]]{}
			err := json.Unmarshal([]byte(tt.jsonData), metaResponse.Body)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BatchResponse.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf(
						"BatchResponse.UnmarshalJSON() error message = %v, want %v",
						err.Error(),
						tt.wantErrMsg,
					)
				}
				return
			}
			if metaResponse.Body.IsBatch != tt.wantIsBatch {
				t.Errorf(
					"BatchResponse.UnmarshalJSON() IsBatch = %v, want %v",
					metaResponse.Body.IsBatch,
					tt.wantIsBatch,
				)
			}
			eq, err := jsonStructEqual(metaResponse.Body.Items, tt.wantItems)
			if err != nil || !eq {
				t.Errorf(
					"BatchResponse.UnmarshalJSON() Items = %+v, want %+v",
					metaResponse.Body.Items,
					tt.wantItems,
				)
			}
		})
	}
}

func TestBatchResponse_MarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		meta       BatchResponse
		wantData   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "Single response with result",
			meta: BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{
						{
							JSONRPC: "2.0",
							Result:  json.RawMessage(`7`),
							ID:      &RequestID{Value: 1},
						},
					},
				},
			},
			wantData: `{"jsonrpc":"2.0","result":7,"id":1}`,
			wantErr:  false,
		},
		{
			name: "Single response with error",
			meta: BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{
						{
							JSONRPC: "2.0",
							Error: &JSONRPCError{
								Code:    -32601,
								Message: "Method not found",
							},
							ID: &RequestID{Value: 2},
						},
					},
				},
			},
			wantData: `{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found"},"id":2}`,
			wantErr:  false,
		},
		{
			name: "Batch responses",
			meta: BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: true,
					Items: []Response[json.RawMessage]{
						{
							JSONRPC: "2.0",
							Result:  json.RawMessage(`7`),
							ID:      &RequestID{Value: 1},
						},
						{
							JSONRPC: "2.0",
							Error: &JSONRPCError{
								Code:    -32601,
								Message: "Method not found",
							},
							ID: &RequestID{Value: 2},
						},
					},
				},
			},
			wantData: `[{"jsonrpc":"2.0","result":7,"id":1},{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found"},"id":2}]`,
			wantErr:  false,
		},
		{
			name: "Empty items with IsBatch=false",
			meta: BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items:   []Response[json.RawMessage]{},
				},
			},
			wantErr:  false,
			wantData: `null`,
		},
		{
			name: "Empty items with IsBatch=true",
			meta: BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: true,
					Items:   []Response[json.RawMessage]{},
				},
			},
			wantData: `[]`,
			wantErr:  false,
		},
		{
			name: "Nil Body",
			meta: BatchResponse{
				Body: nil,
			},
			wantData: "null",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.meta.Body)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BatchResponse.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf(
						"BatchResponse.MarshalJSON() error message = %v, want %v",
						err.Error(),
						tt.wantErrMsg,
					)
				}
				return
			}

			if !jsonStringsEqual(string(data), tt.wantData) {
				t.Errorf(
					"BatchResponse.MarshalJSON() data = %s, want %s",
					string(data),
					tt.wantData,
				)
			}
		})
	}
}
