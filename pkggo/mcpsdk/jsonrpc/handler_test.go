package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestGetMetaRequestHandler(t *testing.T) {
	// Define method maps
	methodMap := map[string]IMethodHandler{
		"add": &MethodHandler[AddParams, AddResult]{Endpoint: AddEndpoint},
		"addErrorSimple": &MethodHandler[AddParams, AddResult]{
			Endpoint: func(ctx context.Context, params AddParams) (AddResult, error) {
				return AddResult{}, errors.New("intentional error")
			},
		},
		"addErrorJSONRPC": &MethodHandler[AddParams, AddResult]{
			Endpoint: func(ctx context.Context, params AddParams) (AddResult, error) {
				return AddResult{}, &JSONRPCError{
					Code:    1234,
					Message: "Custom error",
				}
			},
		},
		"concat": &MethodHandler[ConcatParams, string]{Endpoint: ConcatEndpoint},
	}

	notificationMap := map[string]INotificationHandler{
		"ping": &NotificationHandler[PingParams]{Endpoint: PingEndpoint},
		"notify": &NotificationHandler[NotifyParams]{
			Endpoint: NotifyEndpoint,
		},
		"errornotify": &NotificationHandler[NotifyParams]{
			Endpoint: func(ctx context.Context, params NotifyParams) error {
				return errors.New("processing error")
			},
		},
	}

	// Define test cases
	tests := []struct {
		name         string
		metaReq      *MetaRequest
		expectedResp *MetaResponse
	}{
		{
			name:    "Nil MetaRequest",
			metaReq: nil,
			expectedResp: &MetaResponse{
				Body: &Meta[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPC_VERSION,
						ID:      nil,
						Error: &JSONRPCError{
							Code:    ParseError,
							Message: "No input received for",
						},
					}},
				},
			},
		},
		{
			name: "Empty Body Items",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items:   []Request[json.RawMessage]{},
				},
			},
			expectedResp: &MetaResponse{
				Body: &Meta[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPC_VERSION,
						ID:      nil,
						Error: &JSONRPCError{
							Code:    ParseError,
							Message: "No input received for",
						},
					}},
				},
			},
		},
		{
			name: "Invalid JSON-RPC version",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{
						{
							JSONRPC: "1.0",
							Method:  "add",
							Params:  json.RawMessage(`{"a":1,"b":2}`),
							ID:      &RequestID{Value: 1},
						},
					},
				},
			},
			expectedResp: &MetaResponse{
				Body: &Meta[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPC_VERSION,
						ID:      &RequestID{Value: 1},
						Error: &JSONRPCError{
							Code:    InvalidRequestError,
							Message: "Invalid JSON-RPC version: '1.0'",
						},
					}},
				},
			},
		},
		{
			name: "Invalid notification method",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{{
						JSONRPC: JSONRPC_VERSION,
						Method:  "unknown_notification",
						Params:  json.RawMessage(`{}`),
						ID:      nil,
					}},
				},
			},
			expectedResp: nil,
		},
		{
			name: "Valid notification",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{{
						JSONRPC: JSONRPC_VERSION,
						Method:  "ping",
						Params:  json.RawMessage(`{"message":"hello"}`),
						ID:      nil,
					}},
				},
			},
			expectedResp: nil, // Notifications do not produce a response
		},
		{
			name: "Processing single notification",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "notify",
							Params:  json.RawMessage(`{"message":"Hello"}`),
							ID:      nil, // Notification
						},
					},
				},
			},
			expectedResp: nil,
		},
		{
			name: "Invalid parameters in notification (unmarshaling fails)",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "notify",
							Params:  json.RawMessage(`{"message":123}`),
							ID:      nil, // Notification
						},
					},
				},
			},
			expectedResp: nil,
		},
		{
			name: "Notify Endpoint returns an error",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "errornotify",
							Params:  json.RawMessage(`{"message":"Hello"}`),
							ID:      nil, // Notification
						},
					},
				},
			},
			expectedResp: nil,
		},
		{
			name: "Processing batch of requests and notifications",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: true,
					Items: []Request[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "add",
							Params:  json.RawMessage(`{"a":1,"b":2}`),
							ID:      &RequestID{Value: 1},
						},
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "notify",
							Params:  json.RawMessage(`{"message":"Hello"}`),
							ID:      nil,
						},
					},
				},
			},
			expectedResp: &MetaResponse{
				Body: &Meta[Response[json.RawMessage]]{
					IsBatch: true,
					Items: []Response[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							ID:      &RequestID{Value: 1},
							Result:  json.RawMessage(`{"sum":3}`),
						},
						// No response for notification
					},
				},
			},
		},
		{
			name: "Valid request to 'add' method",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{{
						JSONRPC: JSONRPC_VERSION,
						Method:  "add",
						Params:  json.RawMessage(`{"a":2,"b":3}`),
						ID:      &RequestID{Value: 1},
					}},
				},
			},
			expectedResp: &MetaResponse{
				Body: &Meta[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPC_VERSION,
						ID:      &RequestID{Value: 1},
						Result:  json.RawMessage(`{"sum":5}`),
					}},
				},
			},
		},
		{
			name: "Method with missing method name",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "",
							Params:  json.RawMessage(`{"a":1,"b":2}`),
							ID:      &RequestID{Value: 1},
						},
					},
				},
			},
			expectedResp: &MetaResponse{
				Body: &Meta[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPC_VERSION,
						ID:      &RequestID{Value: 1},
						Error: &JSONRPCError{
							Code:    InvalidRequestError,
							Message: "Method name missing",
						},
					}},
				},
			},
		},
		{
			name: "Method not found",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{{
						JSONRPC: JSONRPC_VERSION,
						Method:  "subtract",
						Params:  json.RawMessage(`{"a":5,"b":2}`),
						ID:      &RequestID{Value: 2},
					}},
				},
			},
			expectedResp: &MetaResponse{
				Body: &Meta[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPC_VERSION,
						ID:      &RequestID{Value: 2},
						Error: &JSONRPCError{
							Code:    MethodNotFoundError,
							Message: "Method 'subtract' not found",
						},
					}},
				},
			},
		},
		{
			name: "Method with invalid ID",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "add",
							Params:  json.RawMessage(`{"a":1,"b":2}`),
							ID:      nil,
						},
					},
				},
			},
			expectedResp: &MetaResponse{
				Body: &Meta[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPC_VERSION,
						ID:      nil,
						Error: &JSONRPCError{
							Code:    InvalidRequestError,
							Message: "Received no requestID for method: 'add'",
						},
					}},
				},
			},
		},
		{
			name: "Batch request with mixed valid and invalid methods",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: true,
					Items: []Request[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "add",
							Params:  json.RawMessage(`{"a":1,"b":2}`),
							ID:      &RequestID{Value: 1},
						},
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "concat",
							Params:  json.RawMessage(`{"s1":"hello","s2":"world"}`),
							ID:      &RequestID{Value: 2},
						},
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "subtract",
							Params:  json.RawMessage(`{"a":5,"b":3}`),
							ID:      &RequestID{Value: 3},
						},
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "ping",
							Params:  json.RawMessage(`{"message":"ping"}`),
							ID:      nil,
						},
					},
				},
			},
			expectedResp: &MetaResponse{
				Body: &Meta[Response[json.RawMessage]]{
					IsBatch: true,
					Items: []Response[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							ID:      &RequestID{Value: 1},
							Result:  json.RawMessage(`{"sum":3}`),
						},
						{
							JSONRPC: JSONRPC_VERSION,
							ID:      &RequestID{Value: 2},
							Result:  json.RawMessage(`"helloworld"`),
						},
						{
							JSONRPC: JSONRPC_VERSION,
							ID:      &RequestID{Value: 3},
							Error: &JSONRPCError{
								Code:    MethodNotFoundError,
								Message: "Method 'subtract' not found",
							},
						},
					},
				},
			},
		},
		{
			name: "Method request with invalid parameters",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "add",
							Params:  json.RawMessage(`{"a":"one","b":2}`),
							ID:      &RequestID{Value: 1},
						},
					},
				},
			},
			expectedResp: &MetaResponse{
				Body: &Meta[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							ID:      &RequestID{Value: 1},
							Error: &JSONRPCError{
								Code:    InvalidParamsError,
								Message: "Invalid parameters: json: cannot unmarshal string into Go struct field AddParams.a of type int",
							},
						},
					},
				},
			},
		},
		{
			name: "Method endpoint returns simple error",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "addErrorSimple",
							Params:  json.RawMessage(`{"a":1,"b":2}`),
							ID:      &RequestID{Value: 1},
						},
					},
				},
			},
			expectedResp: &MetaResponse{
				Body: &Meta[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							ID:      &RequestID{Value: 1},
							Error: &JSONRPCError{
								Code:    InternalError,
								Message: "intentional error",
							},
						},
					},
				},
			},
		},
		{
			name: "Method Endpoint returns a *jsonrpc.Error",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							Method:  "addErrorJSONRPC",
							Params:  json.RawMessage(`{"a":1,"b":2}`),
							ID:      &RequestID{Value: 1},
						},
					},
				},
			},
			expectedResp: &MetaResponse{
				Body: &Meta[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{
						{
							JSONRPC: JSONRPC_VERSION,
							ID:      &RequestID{Value: 1},
							Error: &JSONRPCError{
								Code:    1234,
								Message: "Custom error",
							},
						},
					},
				},
			},
		},
		{
			name: "Handler returns an error",
			metaReq: &MetaRequest{
				Body: &Meta[Request[json.RawMessage]]{
					IsBatch: false,
					Items: []Request[json.RawMessage]{{
						JSONRPC: JSONRPC_VERSION,
						Method:  "add",
						Params:  json.RawMessage(`invalid`),
						ID:      &RequestID{Value: 4},
					}},
				},
			},
			expectedResp: &MetaResponse{
				Body: &Meta[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPC_VERSION,
						ID:      &RequestID{Value: 4},
						Error: &JSONRPCError{
							Code:    InvalidParamsError,
							Message: "Invalid parameters: invalid character 'i' looking for beginning of value",
						},
					}},
				},
			},
		},
	}

	handlerFunc := GetMetaRequestHandler(methodMap, notificationMap)
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handlerFunc(ctx, tt.metaReq)
			if err != nil {
				t.Errorf("handlerFunc returned error: %v", err)
			}
			eq, err := jsonStructEqual(tt.expectedResp, resp)
			if err != nil {
				t.Fatalf("Could not compare struct")
			}
			if !eq {
				vals, err := getJSONStrings(tt.expectedResp, resp)
				if err != nil {
					t.Fatalf("Could not encode json")
				}
				t.Errorf("Expected response %#v, got %#v", vals[0], vals[1])
			}
		})
	}
}
