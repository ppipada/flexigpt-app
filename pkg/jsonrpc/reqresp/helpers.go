package reqresp

import (
	"context"
	"encoding/json"
)

type contextKey string

const (
	ctxKeyRequestID   contextKey = "jsonrpcRequestID"
	ctxKeyMethodName  contextKey = "jsonrpcMethodName"
	ctxKeyMessageType contextKey = "jsonrpcMessageType"
)

// GetRequestID retrieves the RequestID from the context.
func GetRequestID(ctx context.Context) (RequestID, bool) {
	id, ok := ctx.Value(ctxKeyRequestID).(RequestID)
	return id, ok
}

// GetMethodName retrieves the MethodName from the context.
func GetMethodName(ctx context.Context) (string, bool) {
	method, ok := ctx.Value(ctxKeyMethodName).(string)
	return method, ok
}

// IsNotification checks if the request is a notification.
func GetMessageType(ctx context.Context) (MessageType, bool) {
	msgType, ok := ctx.Value(ctxKeyMessageType).(MessageType)
	return msgType, ok
}

// Helper function to create context with request information.
func contextWithRequestInfo(
	parentCtx context.Context,
	methodName string,
	msgType MessageType,
	requestID *RequestID,
) context.Context {
	ctx := context.WithValue(parentCtx, ctxKeyMethodName, methodName)
	ctx = context.WithValue(ctx, ctxKeyMessageType, msgType)
	if requestID != nil {
		ctx = context.WithValue(ctx, ctxKeyRequestID, *requestID)
	}
	return ctx
}

// Helper function to unmarshal generic data.
func unmarshalData[I any](data json.RawMessage) (I, error) {
	var p I
	if data == nil {
		return p, nil
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return p, err
	}
	return p, nil
}

// Helper function to create an InvalidParamsError response.
func invalidParamsResponse(req Request[json.RawMessage], err error) Response[json.RawMessage] {
	return Response[json.RawMessage]{
		JSONRPC: JSONRPCVersion,
		ID:      &req.ID,
		Error: &JSONRPCError{
			Code:    InvalidParamsError,
			Message: GetDefaultErrorMessage(InvalidParamsError) + ": " + err.Error(),
		},
	}
}
