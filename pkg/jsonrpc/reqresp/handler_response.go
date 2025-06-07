package reqresp

import (
	"context"
	"encoding/json"
	"reflect"
)

// IResponseHandler is an interface for handlers that processes responses (no call response expected).
type IResponseHandler interface {
	// Even though there is a error return allowed this is mainly present for any debugging logs etc
	// The client will never receive any error.
	Handle(ctx context.Context, req Response[json.RawMessage]) error
	GetTypes() reflect.Type
}

// ResponseHandler is a handler for processing JSON-RPC responses.
//
// The result type T should be the actual type you expect, but it will be
// passed as *T to the endpoint function, allowing for nil when there's an error.
//
// Usage Scenarios:
//
//  1. Expected Result:
//     Use concrete types for T, and the endpoint will receive *T.
//
//  2. No Result Expected:
//     Use struct{} for T when no result is expected (only checking for errors).
//
// Example:
//
//	// Handler for a response with a specific result type
//	handler := ResponseHandler[ResultType]{
//	    Endpoint: func(ctx context.Context, result *ResultType, err *JSONRPCError) error {
//	        if err != nil {
//	            // Handle the error
//	            return fmt.Errorf("RPC error: %v", err.Message)
//	        }
//	        if result == nil {
//	            return fmt.Errorf("expected result but got nil")
//	        }
//	        // Process the result
//	        return nil
//	    },
//	}
type ResponseHandler[T any] struct {
	// Endpoint is called with the unmarshaled result (as pointer) and/or error.
	Endpoint func(ctx context.Context, result *T, err *JSONRPCError) error
}

// Handle processes a JSON-RPC response.
func (r *ResponseHandler[T]) Handle(ctx context.Context, resp Response[json.RawMessage]) error {
	// If there's an error, call the endpoint with nil result and the error.
	if resp.Error != nil {
		return r.Endpoint(ctx, nil, resp.Error)
	}

	// Unmarshal the result if present.
	result, err := unmarshalData[T](resp.Result)
	if err != nil {
		return err
	}

	// Call the endpoint with the result and nil error.
	return r.Endpoint(ctx, &result, nil)
}

// GetTypes returns the reflect.Type of the expected result.
func (r *ResponseHandler[T]) GetTypes() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func handleResponse(
	ctx context.Context,
	request UnionRequest,
	responseMap map[string]IResponseHandler,
	responseHandlerMapper func(context.Context, Response[json.RawMessage]) (string, error),
) error {
	// Create context with request info.
	resp := Response[json.RawMessage]{
		JSONRPC: request.JSONRPC,
		ID:      request.ID,
		Result:  request.Result,
		Error:   request.Error,
	}
	method, err := responseHandlerMapper(ctx, resp)
	if err != nil {
		return &JSONRPCError{
			Code:    InternalError,
			Message: GetDefaultErrorMessage(InternalError) + ": " + err.Error(),
		}
	}
	subCtx := contextWithRequestInfo(ctx, method, MessageTypeResponse, request.ID)
	handler, ok := responseMap[method]
	if !ok {
		return &JSONRPCError{
			Code:    MethodNotFoundError,
			Message: GetDefaultErrorMessage(MethodNotFoundError) + ": " + method,
		}
	}

	return handler.Handle(subCtx, resp)
}
