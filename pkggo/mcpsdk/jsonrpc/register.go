package jsonrpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// GetDefaultOperation gets the conventional values for jsonrpc as a single operation
func GetDefaultOperation() huma.Operation {

	return huma.Operation{
		Method:        http.MethodPost,
		Path:          "/jsonrpc",
		DefaultStatus: 200,

		Tags:        []string{"JSONRPC"},
		Summary:     "JSONRPC endpoint",
		Description: "Serve all jsonrpc methods",
		OperationID: "jsonrpc",
	}
}

// GetStatusError converts any errors returned into a JSONRPC error response object that implements the huma StatusError interface
// IF the JSONRPC handler is invoked, it should never throw an error, but should return a error response object.
// JSONRPC requires a error case to be covered via the specifications error response object
func GetStatusError(status int, message string, errs ...error) huma.StatusError {
	var foundJSONRPCError *JSONRPCError
	code := InternalError
	if status >= 400 && status < 500 {
		code = InvalidRequestError
	}

	details := make([]string, len(errs))
	for i, err := range errs {
		details[i] = err.Error()
		if converted, ok := err.(huma.ErrorDetailer); ok {
			d := converted.ErrorDetail()
			// See if this is parse error
			if strings.Contains(d.Message, "unmarshal") ||
				strings.Contains(d.Message, "invalid character") ||
				strings.Contains(d.Message, "unexpected end") {
				code = ParseError
			}
		} else if jsonRPCError, ok := err.(JSONRPCError); ok {
			// Check if the error is of type JSONRPCError
			foundJSONRPCError = &jsonRPCError
		}
	}
	// add the http status to details and set status sent back as 200
	details = append(details, fmt.Sprintf("HTTP Status:%d", status))
	status = 200
	// If a JSONRPCError was found, update the message and append JSON-encoded details
	if foundJSONRPCError != nil {
		details = append(details, fmt.Sprintf("Message:%s", message))
		message = foundJSONRPCError.Message
		code = foundJSONRPCError.Code

		// JSON encode the Data field of the found JSONRPCError
		if jsonData, err := json.Marshal(foundJSONRPCError.Data); err == nil {
			details = append(details, string(jsonData))
		}
	}

	return &ResponseStatusError{
		status: status,
		Response: Response[any]{
			JSONRPC: JSONRPC_VERSION,
			ID:      nil,
			Error: &JSONRPCError{
				Code:    code,
				Message: message,
				Data:    details,
			},
		},
	}
}

// Register a new JSONRPC operation.
// The `methodMap` maps from method name to request handlers. Request clients expect a response object
// The `notificationMap` maps from method name to notification handlers. Notification clients do not expect a response
//
// These maps can be instantiated as
//
//	methodMap := map[string]jsonrpc.IMethodHandler{
//		"add": &jsonrpc.MethodHandler[AddParams, int]{Endpoint: AddEndpoint},
//	}
//
//	notificationMap := map[string]jsonrpc.INotificationHandler{
//		"log": &jsonrpc.NotificationHandler[LogParams]{Endpoint: LogEndpoint},
//	}
func Register(
	api huma.API,
	op huma.Operation,
	methodMap map[string]IMethodHandler,
	notificationMap map[string]INotificationHandler,
) {

	reqHandler := GetMetaRequestHandler(methodMap, notificationMap)
	huma.Register(api, op, reqHandler)
}

func intPtr(i int) *int {
	return &i
}
