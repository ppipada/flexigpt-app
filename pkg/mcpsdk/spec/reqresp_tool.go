package spec

import (
	"github.com/flexigpt/flexiui/pkg/mcpsdk/jsonrpc"
)

// Sent from the client to request a list of tools the server has.
type ListToolsRequest jsonrpc.Request[*PaginatedRequestParams]

// The server's response to a tools/list request from the client.
type (
	ListToolsResponse jsonrpc.Response[*ListToolsResult]
	ListToolsResult   struct {
		_ struct{} `json:"-"     additionalProperties:"true"`
		PaginatedResultParams
		Tools []Tool `json:"tools"`
	}
)

// Definition for a tool the client can call.
type Tool struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`

	// A JSON Schema object defining the expected parameters for the tool.
	InputSchema ToolInputSchema `json:"inputSchema"`
}

// A JSON Schema object defining the expected parameters for the tool.
type ToolInputSchema struct {
	Type       string                            `json:"type"                 enum:"object"`
	Properties map[string]map[string]interface{} `json:"properties,omitempty"`
	Required   []string                          `json:"required,omitempty"`
}

// Used by the client to invoke a tool provided by the server.
type (
	CallToolRequest       jsonrpc.Request[CallToolRequestParams]
	CallToolRequestParams struct {
		// This property is reserved by the protocol to allow clients and servers
		// to attach additional metadata to their requests.
		Meta      map[string]interface{} `json:"_meta,omitempty"`
		_         struct{}               `json:"-"                   additionalProperties:"true"`
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
	}
)

// The server's response to a tool call.
//
// Any errors that originate from the tool SHOULD be reported inside the result
// object, with `isError` set to true, _not_ as an MCP protocol-level error
// response. Otherwise, the LLM would not be able to see that an error occurred
// and self-correct.
//
// However, any errors in _finding_ the tool, an error indicating that the
// server does not support tool calls, or any other exceptional conditions,
// should be reported as an MCP error response.
type (
	CallToolResponse jsonrpc.Response[*CallToolResult]
	CallToolResult   struct {
		// This result property is reserved by the protocol to allow clients and servers
		// to attach additional metadata to their responses.
		Meta map[string]interface{} `json:"_meta,omitempty"`

		// Content corresponds to the JSON schema field "content".
		Content []Content `json:"content"`

		// Whether the tool call ended in an error.
		//
		// If not set, this is assumed to be false (the call was successful).
		IsError *bool `json:"isError,omitempty"`
	}
)

// An optional notification from the server to the client, informing it that the
// list of tools it offers has changed. This may be issued by servers without any
// previous subscription from the client.
type ToolListChangedNotification jsonrpc.Notification[*AdditionalParams]
