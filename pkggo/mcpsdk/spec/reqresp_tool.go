package spec

// Sent from the client to request a list of tools the server has.
type ListToolsRequest PaginatedRequest

// The server's response to a tools/list request from the client.
type ListToolsResult struct {
	PaginatedResult

	// Tools corresponds to the JSON schema field "tools".
	Tools []Tool `json:"tools" yaml:"tools" mapstructure:"tools"`
}

// Definition for a tool the client can call.
type Tool struct {
	// A human-readable description of the tool.
	Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`

	// A JSON Schema object defining the expected parameters for the tool.
	InputSchema ToolInputSchema `json:"inputSchema" yaml:"inputSchema" mapstructure:"inputSchema"`

	// The name of the tool.
	Name string `json:"name" yaml:"name" mapstructure:"name"`
}

// A JSON Schema object defining the expected parameters for the tool.
type ToolInputSchema struct {
	// Properties corresponds to the JSON schema field "properties".
	Properties ToolInputSchemaProperties `json:"properties,omitempty" yaml:"properties,omitempty" mapstructure:"properties,omitempty"`

	// Required corresponds to the JSON schema field "required".
	Required []string `json:"required,omitempty" yaml:"required,omitempty" mapstructure:"required,omitempty"`

	// Type corresponds to the JSON schema field "type".
	Type string `json:"type" yaml:"type" mapstructure:"type"`
}

type ToolInputSchemaProperties map[string]map[string]interface{}

// Used by the client to invoke a tool provided by the server.
type CallToolRequest struct {
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params CallToolRequestParams `json:"params" yaml:"params" mapstructure:"params"`
}

type CallToolRequestParams struct {
	// Arguments corresponds to the JSON schema field "arguments".
	Arguments CallToolRequestParamsArguments `json:"arguments,omitempty" yaml:"arguments,omitempty" mapstructure:"arguments,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`
}

type CallToolRequestParamsArguments map[string]interface{}

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
type CallToolResult struct {
	// This result property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their responses.
	Meta CallToolResultMeta `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`

	// Content corresponds to the JSON schema field "content".
	Content []Content `json:"content" yaml:"content" mapstructure:"content"`

	// Whether the tool call ended in an error.
	//
	// If not set, this is assumed to be false (the call was successful).
	IsError *bool `json:"isError,omitempty" yaml:"isError,omitempty" mapstructure:"isError,omitempty"`
}

// This result property is reserved by the protocol to allow clients and servers to
// attach additional metadata to their responses.
type CallToolResultMeta map[string]interface{}

// An optional notification from the server to the client, informing it that the
// list of tools it offers has changed. This may be issued by servers without any
// previous subscription from the client.
type ToolListChangedNotification Notification
