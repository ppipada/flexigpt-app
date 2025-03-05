package reqresp

import "encoding/json"

// http://www.jsonrpc.org/specification
const JSONRPCVersion = "2.0"

type MessageType string

// Constants for message types.
const (
	MessageTypeInvalid      MessageType = "Invalid"
	MessageTypeMethod       MessageType = "Request"
	MessageTypeNotification MessageType = "Notification"
	MessageTypeResponse     MessageType = "Response"
)

// RequestID can be a int or a string.
// Do a type alias as we want marshal/unmarshal etc to be available.
type RequestID = IntString

type Request[T any] struct {
	// Support JSON RPC v2.
	JSONRPC string    `json:"jsonrpc"          enum:"2.0" doc:"JSON-RPC version, must be '2.0'" required:"true"`
	ID      RequestID `json:"id"                          doc:"RequestID is int or string"      required:"true"`
	Method  string    `json:"method"                      doc:"Method to invoke"                required:"true"`
	Params  T         `json:"params,omitempty"            doc:"Method parameters"`
}

// A notification which does not expect a response.
type Notification[T any] struct {
	JSONRPC string `json:"jsonrpc"          enum:"2.0" doc:"JSON-RPC version, must be '2.0'" required:"true"`
	Method  string `json:"method"                      doc:"Method to invoke"                required:"true"`
	Params  T      `json:"params,omitempty"            doc:"Notification parameters"`
}

type Response[T any] struct {
	JSONRPC string        `json:"jsonrpc"          enum:"2.0" doc:"JSON-RPC version, must be '2.0'"                                                   required:"true"`
	ID      *RequestID    `json:"id,omitempty"                doc:"RequestID is int or string. This will be nil only for requests that fail decoding"`
	Result  T             `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type UnionRequest struct {
	JSONRPC string          `json:"jsonrpc"          enum:"2.0" doc:"JSON-RPC version, must be '2.0'" required:"true"`
	ID      *RequestID      `json:"id,omitempty"`
	Method  *string         `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}
