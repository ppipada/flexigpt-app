package reqresp

import (
	"context"
	"encoding/json"
	"reflect"
)

// INotificationHandler is an interface for handlers that process notifications (no response expected).
type INotificationHandler interface {
	// Even though there is a error return allowed this is mainly present for any debugging logs etc
	// The client will never receive any error.
	Handle(ctx context.Context, req Notification[json.RawMessage]) error
	GetTypes() reflect.Type
}

// NotificationHandler is a RPC handler for methods that do not expect a response.
//
// Usage Scenarios:
//
//  1. Compulsory Parameters:
//     Use concrete types for I when input is required.
//
//  2. Optional Input Parameters:
//     Use a pointer type for I to allow passing nil when no input is provided.
//
//  3. No Input Parameters:
//     Use struct{} for I when the handler does not require any input.
//
// Example:
//
//	// Handler with no input
//	handler := NotificationHandler[struct{}]{
//	    Endpoint: func(ctx context.Context, _ struct{}) error {
//	        // Implementation
//	        return nil
//	    },
//	}
type NotificationHandler[I any] struct {
	Endpoint func(ctx context.Context, params I) error
}

// Handle processes a notification (no response expected).
func (n *NotificationHandler[I]) Handle(
	ctx context.Context,
	req Notification[json.RawMessage],
) error {
	params, err := unmarshalData[I](req.Params)
	if err != nil {
		// Cannot send error to client in notification; possibly log internally.
		return err
	}

	// Call the endpoint.
	return n.Endpoint(ctx, params)
}

// GetTypes returns the reflect.Type of the input.
func (m *NotificationHandler[I]) GetTypes() reflect.Type {
	return reflect.TypeOf((*I)(nil)).Elem()
}

func handleNotification(
	ctx context.Context,
	request UnionRequest,
	notificationMap map[string]INotificationHandler,
) error {
	handler, ok := notificationMap[*request.Method]
	if !ok {
		return &JSONRPCError{
			Code: MethodNotFoundError,
			Message: GetDefaultErrorMessage(
				MethodNotFoundError,
			) + ": Notification" + *request.Method,
		}
	}
	subCtx := contextWithRequestInfo(ctx, *request.Method, MessageTypeNotification, nil)
	return handler.Handle(subCtx, Notification[json.RawMessage]{
		JSONRPC: request.JSONRPC,
		Method:  *request.Method,
		Params:  request.Params,
	})
}
