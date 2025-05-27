package spec

import (
	jsonrpcReqResp "github.com/ppipada/flexigpt-app/pkg/jsonrpc/reqresp"
)

// A request from the client to the server, to enable or adjust logging.
type SetLevelRequest jsonrpcReqResp.Request[SetLevelRequestParams]

type SetLevelRequestParams struct {
	// This property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their requests.
	Meta map[string]any `json:"_meta,omitempty"`
	_    struct{}       `json:"-"               additionalProperties:"true"`
	// The level of logging that the client wants to receive from the server. The
	// server should send all logs at this level and higher (i.e., more severe) to the
	// client as notifications/logging/message.
	Level LoggingLevel `json:"level"`
}

// Notification of a log message passed from server to client. If no
// logging/setLevel request has been sent from the client, the server MAY decide
// which messages to send automatically.
type LoggingMessageNotification jsonrpcReqResp.Notification[LoggingMessageNotificationParams]

type LoggingMessageNotificationParams struct {
	// This property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their requests.
	Meta map[string]any `json:"_meta,omitempty"`
	_    struct{}       `json:"-"                additionalProperties:"true"`
	// An optional name of the logger issuing this message.
	Logger *string `json:"logger,omitempty"`

	// The data to be logged, such as a string message or an object. Any JSON
	// serializable type is allowed here.
	Data  any          `json:"data"`
	Level LoggingLevel `json:"level"`
}
