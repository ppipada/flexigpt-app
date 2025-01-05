package jsonrpc

import (
	"reflect"

	"github.com/danielgtaylor/huma/v2"
)

// http://www.jsonrpc.org/specification
const JSONRPC_VERSION = "2.0"

// RequestID can be a int or a string
// Do a type alias as we want marshal/unmarshal etc to be available
type RequestID = IntString

type Request[T any] struct {
	// Support JSON RPC v2.
	JSONRPC string     `json:"jsonrpc"          enum:"2.0" doc:"JSON-RPC version, must be '2.0'"`
	ID      *RequestID `json:"id,omitempty"                doc:"RequestID is int or string for methods and absent for notifications"`
	Method  string     `json:"method"                      doc:"Method to invoke"`
	Params  T          `json:"params,omitempty"            doc:"Method parameters"`
}

type Response[T any] struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      *RequestID    `json:"id,omitempty"`
	Result  T             `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

func (rs Response[T]) Schema(r huma.Registry) *huma.Schema {

	errorObjectSchema := r.Schema(reflect.TypeOf(rs.Error), true, "")

	responseObjectSchema := &huma.Schema{
		Type:     huma.TypeObject,
		Required: []string{"jsonrpc"},
		Properties: map[string]*huma.Schema{
			"jsonrpc": {
				Type:        huma.TypeString,
				Enum:        []any{"2.0"},
				Description: "JSON-RPC version, must be '2.0'",
			},
			"id": {
				Description: "Request identifier. Compulsory for method responses. This MUST be null to the client in case of parse errors etc.",
				OneOf: []*huma.Schema{
					{Type: huma.TypeInteger},
					{Type: huma.TypeString},
				},
			},
			// "result": {},
			// "error":  errorObjectSchema,
		},
		OneOf: []*huma.Schema{
			{
				Required: []string{"result"},
				Properties: map[string]*huma.Schema{
					"result": {
						Description: "Result of the method call",
					},
				},
			},
			{
				Required: []string{"error"},
				Properties: map[string]*huma.Schema{
					"error": errorObjectSchema,
				},
			},
		},
	}

	return responseObjectSchema
}

// A notification which does not expect a response.
type Notification[T any] struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  T      `json:"params,omitempty"`
}
