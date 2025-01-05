package jsonrpc

import (
	"context"
	"encoding/json"
)

// NotificationHandler is a RPC handler for methods that do not expect a response.
type NotificationHandler[I any] struct {
	Endpoint func(ctx context.Context, params I) error
}

// Handle processes a notification (no response expected).
func (n *NotificationHandler[I]) Handle(ctx context.Context, req Request[json.RawMessage]) error {
	params, err := unmarshalParams[I](req)
	if err != nil {
		// Cannot send error to client in notification; possibly log internally
		return err
	}

	// Call the endpoint
	return n.Endpoint(ctx, params)
}
