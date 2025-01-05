package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Method is a RPC handler method with input type I and output type O.
// It expects a response to be returned to the client.
type MethodHandler[I any, O any] struct {
	Endpoint func(ctx context.Context, params I) (O, error)
}

// Handle processes a request expecting a response.
func (m *MethodHandler[I, O]) Handle(
	ctx context.Context,
	req Request[json.RawMessage],
) (Response[json.RawMessage], error) {
	params, err := unmarshalParams[I](req)
	if err != nil {
		// Return InvalidParamsError
		return invalidParamsResponse(req, err), nil
	}

	// Call the handler
	result, err := m.Endpoint(ctx, params)
	if err != nil {
		// Check if err is a *jsonrpc.Error (JSON-RPC error)
		var jsonrpcErr *JSONRPCError
		if errors.As(err, &jsonrpcErr) {
			// Handler returned a JSON-RPC error
			return Response[json.RawMessage]{
				JSONRPC: JSONRPC_VERSION,
				ID:      req.ID,
				Error:   jsonrpcErr,
			}, nil
		}
		// Handler returned a standard error
		return Response[json.RawMessage]{
			JSONRPC: JSONRPC_VERSION,
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    InternalError,
				Message: err.Error(),
			},
		}, nil
	}

	// Marshal the result.
	resultData, err := json.Marshal(result)
	if err != nil {
		return Response[json.RawMessage]{
			JSONRPC: JSONRPC_VERSION,
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    InternalError,
				Message: fmt.Sprintf("Error marshaling result: %v", err),
			},
		}, nil
	}

	// Return the response with the marshaled result
	return Response[json.RawMessage]{
		JSONRPC: JSONRPC_VERSION,
		ID:      req.ID,
		Result:  json.RawMessage(resultData),
	}, nil
}
