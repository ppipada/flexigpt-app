package reqresp

import (
	"context"
	"encoding/json"
	"fmt"
)

// Now, we can define BatchRequest and BatchResponse using BatchItem[T].
type BatchRequest struct {
	Body *BatchItem[UnionRequest]
}

type BatchResponse struct {
	Body *BatchItem[Response[json.RawMessage]]
}

func (b BatchResponse) MarshalJSON() ([]byte, error) {
	if b.Body == nil {
		return []byte("null"), nil
	}
	return json.Marshal(struct {
		Body *BatchItem[Response[json.RawMessage]] `json:"Body"`
	}{
		Body: b.Body,
	})
}

type BatchRequestHandler struct {
	methodMap             map[string]IMethodHandler
	notificationMap       map[string]INotificationHandler
	responseMap           map[string]IResponseHandler
	responseHandlerMapper func(context.Context, Response[json.RawMessage]) (string, error)
}

type HandlerOption func(*BatchRequestHandler)

func WithMethodMap(methodMap map[string]IMethodHandler) HandlerOption {
	return func(h *BatchRequestHandler) {
		h.methodMap = methodMap
	}
}

func WithNotificationMap(notificationMap map[string]INotificationHandler) HandlerOption {
	return func(h *BatchRequestHandler) {
		h.notificationMap = notificationMap
	}
}

func WithResponseMap(
	responseMap map[string]IResponseHandler,
	mapper func(context.Context, Response[json.RawMessage]) (string, error),
) HandlerOption {
	return func(h *BatchRequestHandler) {
		h.responseMap = responseMap
		h.responseHandlerMapper = mapper
	}
}

func NewBatchRequestHandler(
	opts ...HandlerOption,
) *BatchRequestHandler {
	handler := &BatchRequestHandler{
		methodMap:       make(map[string]IMethodHandler),
		notificationMap: make(map[string]INotificationHandler),
		responseMap:     make(map[string]IResponseHandler),
		// Or provide a default mapper if applicable.
		responseHandlerMapper: nil,
	}

	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

func (brh *BatchRequestHandler) Handle(
	ctx context.Context,
	metaReq *BatchRequest,
) (*BatchResponse, error) {
	if metaReq == nil || metaReq.Body == nil || len(metaReq.Body.Items) == 0 {
		item := Response[json.RawMessage]{
			JSONRPC: JSONRPCVersion,
			ID:      nil,
			Error: &JSONRPCError{
				Code:    ParseError,
				Message: GetDefaultErrorMessage(ParseError) + ": No input received",
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
		msgType, jerr := brh.detectMessageType(request)
		if jerr != nil {
			resp.Body.Items = append(resp.Body.Items, Response[json.RawMessage]{
				JSONRPC: JSONRPCVersion,
				ID:      request.ID,
				Error:   jerr,
			})
			continue
		}

		switch {
		case msgType == MessageTypeNotification && brh.notificationMap != nil:
			_ = handleNotification(ctx, request, brh.notificationMap)
			// Cannot return error; possibly log internally
			// Even if notification was not found, you cannot send anything back.
			continue

		case msgType == MessageTypeMethod && brh.methodMap != nil:
			response := handleMethod(ctx, request, brh.methodMap)
			resp.Body.Items = append(resp.Body.Items, response)
			continue

		case msgType == MessageTypeResponse && brh.responseMap != nil && brh.responseHandlerMapper != nil:
			_ = handleResponse(ctx, request, brh.responseMap, brh.responseHandlerMapper)
			continue

		default:
			// Possibly log this.
			continue
		}
	}

	// If there are no responses to return, return empty response.
	if len(resp.Body.Items) == 0 {
		return &BatchResponse{}, nil
	}
	// Log.Printf("%#v", resp.Body).
	return &resp, nil
}

func (brh *BatchRequestHandler) detectMessageType(u UnionRequest) (MessageType, *JSONRPCError) {
	switch {
	case u.JSONRPC != "2.0":
		msg := fmt.Sprintf(
			": Invalid JSON-RPC version: '%s'",
			u.JSONRPC,
		)
		return MessageTypeInvalid, &JSONRPCError{
			Code:    InvalidRequestError,
			Message: GetDefaultErrorMessage(InvalidRequestError) + msg,
		}

	case u.Method != nil:
		// Invalid if both method and result/error are present.
		if u.Result != nil || u.Error != nil {
			return MessageTypeInvalid, &JSONRPCError{
				Code: InvalidRequestError,
				Message: GetDefaultErrorMessage(
					InvalidRequestError,
				) + ": Invalid message: 'method' cannot coexist with 'result' or 'error'",
			}
		}
		// It's a Request or Notification.
		if u.ID != nil {
			return MessageTypeMethod, nil
		}
		return MessageTypeNotification, nil

	case u.Result != nil || u.Error != nil:
		// Invalid if both result and error are present.
		if u.Result != nil && u.Error != nil {
			return MessageTypeInvalid, &JSONRPCError{
				Code: InternalError,
				Message: GetDefaultErrorMessage(
					InternalError,
				) + ": Invalid message: 'result' and 'error' cannot coexist",
			}
		}

		// Response must have an ID.
		if u.ID != nil {
			return MessageTypeResponse, nil
		}
		return MessageTypeInvalid, &JSONRPCError{
			Code:    InternalError,
			Message: GetDefaultErrorMessage(InternalError) + ": Invalid response: missing 'id'",
		}

	default:
		// Invalid message.
		return MessageTypeInvalid, &JSONRPCError{
			Code: InvalidRequestError,
			Message: GetDefaultErrorMessage(
				InvalidRequestError,
			) + ": Unknown message type: missing both 'method' and 'result'/'error'",
		}
	}
}
