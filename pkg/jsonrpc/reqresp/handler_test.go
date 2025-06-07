package reqresp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// AddParams defines the parameters for the "add" method.
type AddParams struct {
	A int `json:"a"`
	B int `json:"b"`
}

type AddResult struct {
	Sum int `json:"sum"`
}

type NotifyParams struct {
	Message string `json:"message"`
}

// ConcatParams defines the parameters for the "concat" method.
type ConcatParams struct {
	S1 string `json:"s1"`
	S2 string `json:"s2"`
}

// PingParams defines the parameters for the "ping" notification.
type PingParams struct {
	Message string `json:"message"`
}

func AddEndpoint(ctx context.Context, params AddParams) (AddResult, error) {
	res := params.A + params.B
	return AddResult{Sum: res}, nil
}

func AddResponseEndpoint(ctx context.Context, result *AddResult, err *JSONRPCError) error {
	return nil
}

// ConcatEndpoint is the handler for the "concat" method.
func ConcatEndpoint(ctx context.Context, params ConcatParams) (string, error) {
	return params.S1 + params.S2, nil
}

// PingEndpoint is the handler for the "ping" notification.
func PingEndpoint(ctx context.Context, params PingParams) error {
	return nil
}

func NotifyEndpoint(ctx context.Context, params NotifyParams) error {
	// Process notification.
	return nil
}

func TestBatchRequestHandler(t *testing.T) {
	// Define method maps.
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

	responseMap := map[string]IResponseHandler{
		"add": &ResponseHandler[AddResult]{Endpoint: AddResponseEndpoint},
	}

	responseHandlerMapper := func(context.Context, Response[json.RawMessage]) (string, error) {
		return "add", nil
	}

	// Define test cases.
	tests := []struct {
		name         string
		metaReq      *BatchRequest
		expectedResp *BatchResponse
	}{
		{
			name:    "Nil BatchRequest",
			metaReq: nil,
			expectedResp: &BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPCVersion,
						ID:      nil,
						Error: &JSONRPCError{
							Code:    ParseError,
							Message: GetDefaultErrorMessage(ParseError) + ": No input received",
						},
					}},
				},
			},
		},
		{
			name: "Empty Body Items",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items:   []UnionRequest{},
				},
			},
			expectedResp: &BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPCVersion,
						ID:      nil,
						Error: &JSONRPCError{
							Code:    ParseError,
							Message: GetDefaultErrorMessage(ParseError) + ": No input received",
						},
					}},
				},
			},
		},
		{
			name: "Invalid JSON-RPC version",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{
						{
							JSONRPC: "1.0",
							Method:  stringToPointer("add"),
							Params:  json.RawMessage(`{"a":1,"b":2}`),
							ID:      &RequestID{Value: 1},
						},
					},
				},
			},
			expectedResp: &BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPCVersion,
						ID:      &RequestID{Value: 1},
						Error: &JSONRPCError{
							Code: InvalidRequestError,
							Message: GetDefaultErrorMessage(
								InvalidRequestError,
							) + ": Invalid JSON-RPC version: '1.0'",
						},
					}},
				},
			},
		},
		{
			name: "Invalid notification method",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{{
						JSONRPC: JSONRPCVersion,
						Method:  stringToPointer("unknown_notification"),
						Params:  json.RawMessage(`{}`),
						ID:      nil,
					}},
				},
			},
			expectedResp: nil,
		},
		{
			name: "Valid notification",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{{
						JSONRPC: JSONRPCVersion,
						Method:  stringToPointer("ping"),
						Params:  json.RawMessage(`{"message":"hello"}`),
						ID:      nil,
					}},
				},
			},
			// Notifications do not produce a response.
			expectedResp: nil,
		},
		{
			name: "Processing single notification",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{
						{
							JSONRPC: JSONRPCVersion,
							Method:  stringToPointer("notify"),
							Params:  json.RawMessage(`{"message":"Hello"}`),
							ID:      nil,
						},
					},
				},
			},
			expectedResp: nil,
		},
		{
			name: "Invalid parameters in notification (unmarshaling fails)",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{
						{
							JSONRPC: JSONRPCVersion,
							Method:  stringToPointer("notify"),
							Params:  json.RawMessage(`{"message":123}`),
							ID:      nil,
						},
					},
				},
			},
			expectedResp: nil,
		},
		{
			name: "Notify Endpoint returns an error",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{
						{
							JSONRPC: JSONRPCVersion,
							Method:  stringToPointer("errornotify"),
							Params:  json.RawMessage(`{"message":"Hello"}`),
							ID:      nil,
						},
					},
				},
			},
			expectedResp: nil,
		},
		{
			name: "Processing batch of requests and notifications",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: true,
					Items: []UnionRequest{
						{
							JSONRPC: JSONRPCVersion,
							Method:  stringToPointer("add"),
							Params:  json.RawMessage(`{"a":1,"b":2}`),
							ID:      &RequestID{Value: 1},
						},
						{
							JSONRPC: JSONRPCVersion,
							Method:  stringToPointer("notify"),
							Params:  json.RawMessage(`{"message":"Hello"}`),
							ID:      nil,
						},
					},
				},
			},
			expectedResp: &BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: true,
					Items: []Response[json.RawMessage]{
						{
							JSONRPC: JSONRPCVersion,
							ID:      &RequestID{Value: 1},
							Result:  json.RawMessage(`{"sum":3}`),
						},
						// No response for notification.
					},
				},
			},
		},
		{
			name: "Valid request to 'add' method",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{{
						JSONRPC: JSONRPCVersion,
						Method:  stringToPointer("add"),
						Params:  json.RawMessage(`{"a":2,"b":3}`),
						ID:      &RequestID{Value: 1},
					}},
				},
			},
			expectedResp: &BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPCVersion,
						ID:      &RequestID{Value: 1},
						Result:  json.RawMessage(`{"sum":5}`),
					}},
				},
			},
		},
		{
			name: "Method with missing method name",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{
						{
							JSONRPC: JSONRPCVersion,
							Method:  stringToPointer(""),
							Params:  json.RawMessage(`{"a":1,"b":2}`),
							ID:      &RequestID{Value: 1},
						},
					},
				},
			},
			expectedResp: &BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPCVersion,
						ID:      &RequestID{Value: 1},
						Error: &JSONRPCError{
							Code:    MethodNotFoundError,
							Message: GetDefaultErrorMessage(MethodNotFoundError) + ": ",
						},
					}},
				},
			},
		},
		{
			name: "Method not found",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{{
						JSONRPC: JSONRPCVersion,
						Method:  stringToPointer("subtract"),
						Params:  json.RawMessage(`{"a":5,"b":2}`),
						ID:      &RequestID{Value: 2},
					}},
				},
			},
			expectedResp: &BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPCVersion,
						ID:      &RequestID{Value: 2},
						Error: &JSONRPCError{
							Code:    MethodNotFoundError,
							Message: GetDefaultErrorMessage(MethodNotFoundError) + ": subtract",
						},
					}},
				},
			},
		},
		{
			name: "Batch request with mixed valid and invalid methods",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: true,
					Items: []UnionRequest{
						{
							JSONRPC: JSONRPCVersion,
							Method:  stringToPointer("add"),
							Params:  json.RawMessage(`{"a":1,"b":2}`),
							ID:      &RequestID{Value: 1},
						},
						{
							JSONRPC: JSONRPCVersion,
							Method:  stringToPointer("concat"),
							Params:  json.RawMessage(`{"s1":"hello","s2":"world"}`),
							ID:      &RequestID{Value: 2},
						},
						{
							JSONRPC: JSONRPCVersion,
							Method:  stringToPointer("subtract"),
							Params:  json.RawMessage(`{"a":5,"b":3}`),
							ID:      &RequestID{Value: 3},
						},
						{
							JSONRPC: JSONRPCVersion,
							Method:  stringToPointer("ping"),
							Params:  json.RawMessage(`{"message":"ping"}`),
							ID:      nil,
						},
					},
				},
			},
			expectedResp: &BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: true,
					Items: []Response[json.RawMessage]{
						{
							JSONRPC: JSONRPCVersion,
							ID:      &RequestID{Value: 1},
							Result:  json.RawMessage(`{"sum":3}`),
						},
						{
							JSONRPC: JSONRPCVersion,
							ID:      &RequestID{Value: 2},
							Result:  json.RawMessage(`"helloworld"`),
						},
						{
							JSONRPC: JSONRPCVersion,
							ID:      &RequestID{Value: 3},
							Error: &JSONRPCError{
								Code:    MethodNotFoundError,
								Message: GetDefaultErrorMessage(MethodNotFoundError) + ": subtract",
							},
						},
					},
				},
			},
		},
		{
			name: "Method request with invalid parameters",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{
						{
							JSONRPC: JSONRPCVersion,
							Method:  stringToPointer("add"),
							Params:  json.RawMessage(`{"a":"one","b":2}`),
							ID:      &RequestID{Value: 1},
						},
					},
				},
			},
			expectedResp: &BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{
						{
							JSONRPC: JSONRPCVersion,
							ID:      &RequestID{Value: 1},
							Error: &JSONRPCError{
								Code: InvalidParamsError,
								Message: GetDefaultErrorMessage(
									InvalidParamsError,
								) + ": json: cannot unmarshal string into Go struct field AddParams.a of type int",
							},
						},
					},
				},
			},
		},
		{
			name: "Method endpoint returns simple error",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{
						{
							JSONRPC: JSONRPCVersion,
							Method:  stringToPointer("addErrorSimple"),
							Params:  json.RawMessage(`{"a":1,"b":2}`),
							ID:      &RequestID{Value: 1},
						},
					},
				},
			},
			expectedResp: &BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{
						{
							JSONRPC: JSONRPCVersion,
							ID:      &RequestID{Value: 1},
							Error: &JSONRPCError{
								Code: InternalError,
								Message: GetDefaultErrorMessage(
									InternalError,
								) + ": intentional error",
							},
						},
					},
				},
			},
		},
		{
			name: "Method Endpoint returns a *jsonrpc.Error",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{
						{
							JSONRPC: JSONRPCVersion,
							Method:  stringToPointer("addErrorJSONRPC"),
							Params:  json.RawMessage(`{"a":1,"b":2}`),
							ID:      &RequestID{Value: 1},
						},
					},
				},
			},
			expectedResp: &BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{
						{
							JSONRPC: JSONRPCVersion,
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
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{{
						JSONRPC: JSONRPCVersion,
						Method:  stringToPointer("add"),
						Params:  json.RawMessage(`invalid`),
						ID:      &RequestID{Value: 4},
					}},
				},
			},
			expectedResp: &BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items: []Response[json.RawMessage]{{
						JSONRPC: JSONRPCVersion,
						ID:      &RequestID{Value: 4},
						Error: &JSONRPCError{
							Code: InvalidParamsError,
							Message: GetDefaultErrorMessage(
								InvalidParamsError,
							) + ": invalid character 'i' looking for beginning of value",
						},
					}},
				},
			},
		},
		{
			name: "Valid response from 'add' method",
			metaReq: &BatchRequest{
				Body: &BatchItem[UnionRequest]{
					IsBatch: false,
					Items: []UnionRequest{{
						JSONRPC: JSONRPCVersion,
						Result:  json.RawMessage(`{"a":2,"b":3}`),
						ID:      &RequestID{Value: 1},
					}},
				},
			},
			expectedResp: nil,
		},
	}

	brh := NewBatchRequestHandler(WithMethodMap(methodMap),
		WithNotificationMap(notificationMap),
		WithResponseMap(responseMap, responseHandlerMapper),
	)
	ctx := t.Context()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := brh.Handle(ctx, tt.metaReq)
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
