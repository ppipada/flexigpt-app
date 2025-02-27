package humaadapter

import (
	"reflect"

	"github.com/danielgtaylor/huma/v2"
	jsonrpcReqResp "github.com/flexigpt/flexiui/pkg/jsonrpc/reqresp"
)

type ResponseStatusError struct {
	jsonrpcReqResp.Response[any]
	status int `json:"-"`
}

func (e *ResponseStatusError) Error() string {
	if e.Response.Error != nil {
		return e.Response.Error.Message
	}
	return ""
}

func (e *ResponseStatusError) GetStatus() int {
	return e.status
}

func (e ResponseStatusError) Schema(r huma.Registry) *huma.Schema {
	errorObjectSchema := r.Schema(reflect.TypeOf(e.Response.Error), true, "")

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
			"error": errorObjectSchema,
		},
	}

	return responseObjectSchema
}
