package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

func InitJSONRPChandlers(api huma.API) {
	// Define method maps
	methodMap := map[string]IMethodHandler{
		"add": &MethodHandler[AddParams, AddResult]{Endpoint: AddEndpoint},
		"addpositional": &MethodHandler[[]int, AddResult]{
			Endpoint: func(ctx context.Context, params []int) (AddResult, error) {
				res := 0
				for _, v := range params {
					res += v
				}
				return AddResult{Sum: res}, nil
			},
		},
		"concat": &MethodHandler[ConcatParams, string]{Endpoint: ConcatEndpoint},
		"echo": &MethodHandler[struct{}, *struct{}]{
			Endpoint: func(ctx context.Context, _ struct{}) (*struct{}, error) {
				return nil, nil
			},
		},
		"echooptional": &MethodHandler[*string, *string]{
			Endpoint: func(ctx context.Context, e *string) (*string, error) {
				return e, nil
			},
		},
	}

	notificationMap := map[string]INotificationHandler{
		"ping": &NotificationHandler[PingParams]{Endpoint: PingEndpoint},
		"notify": &NotificationHandler[NotifyParams]{
			Endpoint: NotifyEndpoint,
		},
	}

	// Get default operation
	op := GetDefaultOperation()

	// Register the methods
	Register(api, op, methodMap, notificationMap)

}

func loggingMiddleware(ctx huma.Context, next func(huma.Context)) {
	// log.Printf("Received request: %v %v", ctx.URL().RawPath, ctx.Operation().Path)
	next(ctx)
	// log.Printf("Responded to request: %v %v", ctx.URL().RawPath, ctx.Operation().Path)
}

var DefaultJSONFormat = huma.Format{
	Marshal: func(w io.Writer, v any) error {
		return json.NewEncoder(w).Encode(v)
	},
	Unmarshal: func(data []byte, v any) error {
		// log.Printf("Trying to unmarshal %v", string(data))
		err := json.Unmarshal(data, v)
		// log.Printf("err %v", err)
		return err
	},
}

func getRouter() *http.ServeMux {
	router := http.NewServeMux()
	config := huma.DefaultConfig("Example JSONRPC API", "1.0.0")
	config.Formats = map[string]huma.Format{
		"application/json": DefaultJSONFormat,
		"json":             DefaultJSONFormat,
	}
	huma.NewError = GetStatusError

	api := humago.New(router, config)
	api.UseMiddleware(loggingMiddleware)

	InitJSONRPChandlers(api)
	return router
}

func setupTestServer(t *testing.T) (*httptest.Server, *http.Client, string) {
	router := getRouter()
	server := httptest.NewUnstartedServer(router)
	server.Start()
	t.Cleanup(server.Close) // Ensure server closes after test
	client := server.Client()
	url := server.URL + "/jsonrpc"
	return server, client, url
}

func sendJSONRPCRequest(t *testing.T, client *http.Client, url string, request interface{}) []byte {
	var reqBytes []byte
	var err error
	if b, ok := request.([]byte); ok {
		reqBytes = b
	} else {
		reqBytes, err = json.Marshal(request)
		if err != nil {
			t.Fatalf("JSONRPCError marshaling request: %v", err)
		}
	}
	t.Logf("Sending req %s", string(reqBytes))
	resp, err := client.Post(url, "application/json", bytes.NewReader(reqBytes))
	if err != nil {
		t.Fatalf("JSONRPCError sending request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("JSONRPCError reading response: %v", err)
	}

	if len(respBody) == 0 {
		t.Log("Got Empty response")
		return nil
	}
	var o interface{}
	err = json.Unmarshal(respBody, &o)
	if err == nil {
		r, err := json.Marshal(o)
		if err == nil {
			t.Logf("Json resp %s", string(r))
		}
	}

	return respBody
}

func TestValidSingleRequests(t *testing.T) {
	_, client, url := setupTestServer(t)

	tests := []struct {
		name           string
		request        interface{}
		expectedResult interface{}
	}{
		{
			name: "Add method with named parameters",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "add",
				"params":  map[string]interface{}{"a": 2, "b": 3},
				"id":      1,
			},
			expectedResult: map[string]float64{"sum": 5},
		},
		{
			name: "Add method with positional parameters",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "addpositional",
				"params":  []interface{}{2, 3},
				"id":      2,
			},
			expectedResult: map[string]float64{"sum": 5},
		},
		{
			name: "Echo method with no parameters",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "echo",
				"id":      3,
			},
			expectedResult: nil,
		},
		{
			name: "Echo method with optional parameters",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "echooptional",
				"id":      "1",
				"params":  "foo",
			},
			expectedResult: "foo",
		},
		{
			name: "Echo method with optional parameters nil input",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "echooptional",
				"id":      "2",
			},
			expectedResult: nil,
		},
		{
			name: "Concat method",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "concat",
				"params":  map[string]interface{}{"s1": "Hello, ", "s2": "World!"},
				"id":      2,
			},
			expectedResult: "Hello, World!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			respBody := sendJSONRPCRequest(t, client, url, tc.request)

			var response struct {
				JSONRPC      string        `json:"jsonrpc"`
				Result       interface{}   `json:"result"`
				JSONRPCError *JSONRPCError `json:"error"`
				ID           interface{}   `json:"id"`
			}

			err := json.Unmarshal(respBody, &response)
			if err != nil {
				t.Fatalf("Error unmarshaling response: %v", err)
			}

			if response.JSONRPCError != nil {
				t.Errorf("Expected no error, but got: %+v", response.JSONRPCError)
			} else {
				eq, err := jsonStructEqual(response.Result, tc.expectedResult)
				if err != nil || !eq {
					t.Errorf("Expected result %#v, got %#v", tc.expectedResult, response.Result)
				}
			}
		})
	}
}

func TestInvalidSingleRequests(t *testing.T) {
	_, client, url := setupTestServer(t)

	tests := []struct {
		name          string
		request       interface{}
		rawRequest    []byte
		expectedError *JSONRPCError
	}{
		{
			name:       "Invalid JSON request",
			rawRequest: []byte(`{ this is invalid json }`),
			expectedError: &JSONRPCError{
				Code:    -32700,
				Message: "validation failed",
			},
		},
		{
			name: "Method not found",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "unknown_method",
				"id":      1,
			},
			expectedError: &JSONRPCError{
				Code:    -32601,
				Message: "Method 'unknown_method' not found",
			},
		},
		{
			name: "Invalid parameters",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "add",
				"params":  map[string]interface{}{"a": "two", "b": 3},
				"id":      2,
			},
			expectedError: &JSONRPCError{
				Code:    -32602,
				Message: "Invalid parameters",
			},
		},
		{
			name: "Missing jsonrpc field",
			request: map[string]interface{}{
				"method": "add",
				"params": map[string]interface{}{"a": 2, "b": 3},
				"id":     3,
			},
			expectedError: &JSONRPCError{
				Code:    -32600,
				Message: "validation failed",
			},
		},
		{
			name: "Invalid jsonrpc version",
			request: map[string]interface{}{
				"jsonrpc": "1.0",
				"method":  "add",
				"params":  map[string]interface{}{"a": 2, "b": 3},
				"id":      4,
			},
			expectedError: &JSONRPCError{
				Code:    -32600,
				Message: "validation failed",
			},
		},
		{
			name: "Missing method field",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"params":  map[string]interface{}{"a": 2, "b": 3},
				"id":      5,
			},
			expectedError: &JSONRPCError{
				Code:    -32600,
				Message: "validation failed",
			},
		},
		{
			name: "Invalid id field (array)",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "add",
				"params":  map[string]interface{}{"a": 2, "b": 3},
				"id":      []int{1, 2, 3},
			},
			expectedError: &JSONRPCError{
				Code:    -32600,
				Message: "validation failed",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.request
			if tc.rawRequest != nil {
				req = tc.rawRequest
			}

			respBody := sendJSONRPCRequest(t, client, url, req)

			var response struct {
				JSONRPC      string        `json:"jsonrpc"`
				Result       interface{}   `json:"result"`
				JSONRPCError *JSONRPCError `json:"error"`
				ID           interface{}   `json:"id"`
			}

			err := json.Unmarshal(respBody, &response)
			if err != nil {
				t.Fatalf("Error unmarshaling response: %v", err)
			}

			if response.JSONRPCError == nil {
				t.Errorf("Expected error but got none")
			} else {
				if response.JSONRPCError.Code != tc.expectedError.Code {
					t.Errorf("Expected error code %d, got %d", tc.expectedError.Code, response.JSONRPCError.Code)
				}
				if !strings.Contains(response.JSONRPCError.Message, tc.expectedError.Message) {
					t.Errorf("Expected error message '%s', got '%s'", tc.expectedError.Message, response.JSONRPCError.Message)
				}
			}
		})
	}
}

func TestNotifications(t *testing.T) {
	_, client, url := setupTestServer(t)

	tests := []struct {
		name    string
		request interface{}
	}{
		{
			name: "Valid notification",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "notify",
				"params":  map[string]interface{}{"message": "Hello"},
			},
		},
		{
			name: "Notification with invalid method",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "unknown_method",
				"params":  map[string]interface{}{"message": "Hello"},
			},
		},
		{
			name: "Ping notification",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "ping",
				"params":  map[string]interface{}{"message": "Test Ping"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			respBody := sendJSONRPCRequest(t, client, url, tc.request)

			if len(respBody) != 0 {
				t.Errorf("Expected no response, but got: %s", string(respBody))
			}
		})
	}
}

func TestBatchRequests(t *testing.T) {
	_, client, url := setupTestServer(t)

	tests := []struct {
		name               string
		batchRequest       []interface{}
		expectedResponses  int
		expectedErrorCodes []int
		expectedResults    map[interface{}]interface{}
	}{
		{
			name: "Valid batch with multiple requests",
			batchRequest: []interface{}{
				map[string]interface{}{
					"jsonrpc": "2.0",
					"method":  "add",
					"params":  map[string]interface{}{"a": 1, "b": 2},
					"id":      1,
				},
				map[string]interface{}{
					"jsonrpc": "2.0",
					"method":  "echooptional",
					"params":  "foo",
					"id":      2,
				},
			},
			expectedResponses:  2,
			expectedErrorCodes: []int{},
		},
		{
			name: "Batch with mixed valid requests and notifications",
			batchRequest: []interface{}{
				map[string]interface{}{
					"jsonrpc": "2.0",
					"method":  "add",
					"params":  map[string]interface{}{"a": 1, "b": 2},
					"id":      1,
				},
				map[string]interface{}{
					"jsonrpc": "2.0",
					"method":  "notify",
					"params":  map[string]interface{}{"message": "Hello"},
				},
			},
			expectedResponses:  1,
			expectedErrorCodes: []int{},
		},
		{
			name: "Batch with invalid JSON in one request",
			batchRequest: []interface{}{[]byte(`[{
							"jsonrpc": "2.0",
							"method": "add",
							"params": {"a":1,"b":2},
							"id":1
					}, {
							"jsonrpc": "2.0",
							"method": "invalid_method",
							"params": {},
							"id":2
					}`)}, // Incomplete closing square bracket
			expectedResponses:  1,
			expectedErrorCodes: []int{-32700},
		},
		{
			name: "Batch of notifications",
			batchRequest: []interface{}{
				map[string]interface{}{
					"jsonrpc": "2.0",
					"method":  "notify",
					"params":  map[string]interface{}{"message": "Hello"},
				},
				map[string]interface{}{
					"jsonrpc": "2.0",
					"method":  "notify",
					"params":  map[string]interface{}{"message": "World"},
				},
			},
			expectedResponses:  0,
			expectedErrorCodes: []int{},
		},
		{
			name:               "Empty batch array",
			batchRequest:       []interface{}{},
			expectedResponses:  1,
			expectedErrorCodes: []int{-32600},
		},
		{
			name: "Batch with valid and invalid methods",
			batchRequest: []interface{}{
				map[string]interface{}{
					"jsonrpc": "2.0",
					"method":  "add",
					"params":  map[string]interface{}{"a": 1, "b": 2},
					"id":      1,
				},
				map[string]interface{}{
					"jsonrpc": "2.0",
					"method":  "concat",
					"params":  map[string]interface{}{"s1": "foo", "s2": "bar"},
					"id":      2,
				},
				map[string]interface{}{
					"jsonrpc": "2.0",
					"method":  "unknownMethod",
					"id":      3,
				},
			},
			expectedResponses:  3,
			expectedErrorCodes: []int{-32601},
			expectedResults: map[interface{}]interface{}{
				float64(1): map[string]interface{}{"sum": float64(3)},
				float64(2): "foobar",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var batch interface{}
			if len(tc.batchRequest) > 0 {
				if b, ok := tc.batchRequest[0].([]byte); ok {
					batch = b
				} else {
					batch = tc.batchRequest
				}
			}
			respBody := sendJSONRPCRequest(t, client, url, batch)

			if tc.expectedResponses == 0 {
				if len(respBody) != 0 {
					t.Errorf("Expected no response, but got: %s", string(respBody))
					return
				} else {
					return
				}
			}

			var responses []struct {
				JSONRPC      string        `json:"jsonrpc"`
				Result       interface{}   `json:"result"`
				JSONRPCError *JSONRPCError `json:"error"`
				ID           interface{}   `json:"id"`
			}

			if err := json.Unmarshal(respBody, &responses); err != nil {
				var singleResponse struct {
					JSONRPC      string        `json:"jsonrpc"`
					Result       interface{}   `json:"result"`
					JSONRPCError *JSONRPCError `json:"error"`
					ID           interface{}   `json:"id"`
				}
				if err := json.Unmarshal(respBody, &singleResponse); err != nil {
					t.Fatalf("Error unmarshaling response: %v", err)
				}
				responses = []struct {
					JSONRPC      string        `json:"jsonrpc"`
					Result       interface{}   `json:"result"`
					JSONRPCError *JSONRPCError `json:"error"`
					ID           interface{}   `json:"id"`
				}{singleResponse}
			}

			if len(responses) != tc.expectedResponses {
				t.Errorf("Expected %d responses, got %d", tc.expectedResponses, len(responses))
			}

			var gotErrorCodes []int
			for _, response := range responses {
				if response.JSONRPCError != nil {
					gotErrorCodes = append(gotErrorCodes, response.JSONRPCError.Code)
				}
			}
			if !arraysAreSimilar(gotErrorCodes, tc.expectedErrorCodes) {
				t.Errorf(
					"Mismatched error codes. Got: %#v, Expected: %#v",
					gotErrorCodes,
					tc.expectedErrorCodes,
				)
			}

			if tc.expectedResults != nil {
				for _, response := range responses {
					id := response.ID
					expectedResult, ok := tc.expectedResults[id]
					if ok {
						if response.JSONRPCError != nil {
							t.Errorf(
								"Expected result for id %v, but got error: %+v",
								id,
								response.JSONRPCError,
							)
						} else {
							eq, err := jsonStructEqual(response.Result, expectedResult)
							if err != nil {
								t.Errorf("Error comparing result for id %v: %v", id, err)
							} else if !eq {
								t.Errorf("Mismatched result for id %v. Got: %+v, Expected: %+v", id, response.Result, expectedResult)
							}
						}
					}
				}
			}
		})
	}
}
