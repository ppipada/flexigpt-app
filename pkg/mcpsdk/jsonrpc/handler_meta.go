package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
)

type UnionRequest struct {
	JSONRPC string          `json:"jsonrpc"          enum:"2.0" doc:"JSON-RPC version, must be '2.0'" required:"true"`
	ID      *RequestID      `json:"id,omitempty"`
	Method  *string         `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// Now, we can define BatchRequest and BatchResponse using BatchItem[T].
type BatchRequest struct {
	Body *BatchItem[UnionRequest]
}

type BatchResponse struct {
	Body *BatchItem[Response[json.RawMessage]]
}

func detectMessageType(u UnionRequest) (MessageType, *JSONRPCError) {
	switch {
	case u.JSONRPC != "2.0":
		msg := fmt.Sprintf(
			"Invalid JSON-RPC version: '%s'",
			u.JSONRPC,
		)
		return MessageTypeInvalid, &JSONRPCError{
			Code:    InvalidRequestError,
			Message: msg,
		}

	case u.Method != nil:
		// Invalid if both method and result/error are present
		if u.Result != nil || u.Error != nil {
			return MessageTypeInvalid, &JSONRPCError{
				Code:    InvalidRequestError,
				Message: "Invalid message: 'method' cannot coexist with 'result' or 'error'",
			}
		}
		// It's a Request or Notification
		if u.ID != nil {
			return MessageTypeMethod, nil
		}
		return MessageTypeNotification, nil

	case u.Result != nil || u.Error != nil:
		// Invalid if both result and error are present
		if u.Result != nil && u.Error != nil {
			return MessageTypeInvalid, &JSONRPCError{
				Code:    InternalError,
				Message: "Invalid message: 'result' and 'error' cannot coexist",
			}
		}

		// Response must have an ID
		if u.ID != nil {
			return MessageTypeResponse, nil
		}
		return MessageTypeInvalid, &JSONRPCError{
			Code:    InternalError,
			Message: "Invalid response: missing 'id'",
		}

	default:
		// Invalid message
		return MessageTypeInvalid, &JSONRPCError{
			Code:    InternalError,
			Message: "Unknown message type: missing both 'method' and 'result'/'error'",
		}
	}
}

func handleNotification(
	ctx context.Context,
	request UnionRequest,
	notificationMap map[string]INotificationHandler,
) error {
	handler, ok := notificationMap[*request.Method]
	if !ok {
		return &JSONRPCError{
			Code:    MethodNotFoundError,
			Message: fmt.Sprintf("Notification '%s' not found", *request.Method),
		}
	}
	subCtx := contextWithRequestInfo(ctx, *request.Method, MessageTypeNotification, nil)
	return handler.Handle(subCtx, Notification[json.RawMessage]{
		JSONRPC: request.JSONRPC,
		Method:  *request.Method,
		Params:  request.Params,
	})
}

func handleResponse(
	ctx context.Context,
	request UnionRequest,
	responseMap map[string]IResponseHandler,
	responseHandlerMapper func(context.Context, Response[json.RawMessage]) (string, error),
) error {
	// Create context with request info
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
			Message: err.Error(),
		}
	}
	subCtx := contextWithRequestInfo(ctx, method, MessageTypeResponse, request.ID)
	handler, ok := responseMap[method]
	if !ok {
		return &JSONRPCError{
			Code:    MethodNotFoundError,
			Message: fmt.Sprintf("Method '%s' not found", method),
		}
	}

	return handler.Handle(subCtx, resp)
}

func handleMethod(
	ctx context.Context,
	request UnionRequest,
	methodMap map[string]IMethodHandler,
) Response[json.RawMessage] {
	handler, ok := methodMap[*request.Method]
	if !ok {
		return Response[json.RawMessage]{
			JSONRPC: JSONRPCVersion,
			ID:      request.ID,
			Error: &JSONRPCError{
				Code:    MethodNotFoundError,
				Message: fmt.Sprintf("Method '%s' not found", *request.Method),
			},
		}
	}
	if request.ID == nil {
		return Response[json.RawMessage]{
			JSONRPC: JSONRPCVersion,
			ID:      request.ID,
			Error: &JSONRPCError{
				Code:    InvalidRequestError,
				Message: fmt.Sprintf("Received no requestID for method: '%s'", *request.Method),
			},
		}
	}
	subCtx := contextWithRequestInfo(ctx, *request.Method, MessageTypeMethod, request.ID)
	return handler.Handle(subCtx, Request[json.RawMessage]{
		JSONRPC: request.JSONRPC,
		ID:      *request.ID,
		Method:  *request.Method,
		Params:  request.Params,
	})
}

// GetBatchRequestHandler creates a handler function that processes BatchRequests.
func GetBatchRequestHandler(
	methodMap map[string]IMethodHandler,
	notificationMap map[string]INotificationHandler,
	responseMap map[string]IResponseHandler,
	responseHandlerMapper func(context.Context, Response[json.RawMessage]) (string, error),
) func(context.Context, *BatchRequest) (*BatchResponse, error) {
	return func(ctx context.Context, metaReq *BatchRequest) (*BatchResponse, error) {
		if metaReq == nil || metaReq.Body == nil || len(metaReq.Body.Items) == 0 {
			item := Response[json.RawMessage]{
				JSONRPC: JSONRPCVersion,
				ID:      nil,
				Error: &JSONRPCError{
					Code:    ParseError,
					Message: "No input received for",
				},
			}
			// Return single error if invalid batch or even a single item cannot be found.
			ret := BatchResponse{
				Body: &BatchItem[Response[json.RawMessage]]{
					IsBatch: false,
					Items:   []Response[json.RawMessage]{item},
				},
			}
			return &ret, nil
		}

		resp := BatchResponse{
			Body: &BatchItem[Response[json.RawMessage]]{
				IsBatch: metaReq.Body.IsBatch,
				Items:   []Response[json.RawMessage]{},
			},
		}

		for _, request := range metaReq.Body.Items {
			msgType, jerr := detectMessageType(request)
			if jerr != nil {
				resp.Body.Items = append(resp.Body.Items, Response[json.RawMessage]{
					JSONRPC: JSONRPCVersion,
					ID:      request.ID,
					Error:   jerr,
				})
				continue
			}

			switch {
			case msgType == MessageTypeNotification && notificationMap != nil:
				_ = handleNotification(ctx, request, notificationMap)
				// Cannot return error; possibly log internally
				// Even if notification was not found, you cannot send anything back.
				continue

			case msgType == MessageTypeMethod && methodMap != nil:
				response := handleMethod(ctx, request, methodMap)
				resp.Body.Items = append(resp.Body.Items, response)
				continue

			case msgType == MessageTypeResponse && responseMap != nil && responseHandlerMapper != nil:
				_ = handleResponse(ctx, request, responseMap, responseHandlerMapper)
				continue

			default:
				// Possibly log this.
				continue
			}
		}

		// If there are no responses to return, return nil response.
		if len(resp.Body.Items) == 0 {
			return nil, nil
		}

		return &resp, nil
	}
}
