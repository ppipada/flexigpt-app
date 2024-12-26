package spec

// Sent from the client to request a list of prompts and prompt templates the
// server has.
type ListPromptsRequest PaginatedRequest

// The server's response to a prompts/list request from the client.
type ListPromptsResult struct {
	PaginatedResult

	// Prompts corresponds to the JSON schema field "prompts".
	Prompts []Prompt `json:"prompts" yaml:"prompts" mapstructure:"prompts"`
}

// A prompt or prompt template that the server offers.
type Prompt struct {
	// A list of arguments to use for templating the prompt.
	Arguments []PromptArgument `json:"arguments,omitempty" yaml:"arguments,omitempty" mapstructure:"arguments,omitempty"`

	// An optional description of what this prompt provides
	Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`

	// The name of the prompt or prompt template.
	Name string `json:"name" yaml:"name" mapstructure:"name"`
}

// Describes an argument that a prompt can accept.
type PromptArgument struct {
	// A human-readable description of the argument.
	Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`

	// The name of the argument.
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// Whether this argument must be provided.
	Required *bool `json:"required,omitempty" yaml:"required,omitempty" mapstructure:"required,omitempty"`
}

// Used by the client to get a prompt provided by the server.
type GetPromptRequest struct {
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params GetPromptRequestParams `json:"params" yaml:"params" mapstructure:"params"`
}

type GetPromptRequestParams struct {
	// Arguments to use for templating the prompt.
	Arguments GetPromptRequestParamsArguments `json:"arguments,omitempty" yaml:"arguments,omitempty" mapstructure:"arguments,omitempty"`

	// The name of the prompt or prompt template.
	Name string `json:"name" yaml:"name" mapstructure:"name"`
}

// Arguments to use for templating the prompt.
type GetPromptRequestParamsArguments map[string]string

// The server's response to a prompts/get request from the client.
type GetPromptResult struct {
	// This result property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their responses.
	Meta map[string]interface{} `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`

	// An optional description for the prompt.
	Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`

	// Messages corresponds to the JSON schema field "messages".
	Messages []PromptMessage `json:"messages" yaml:"messages" mapstructure:"messages"`
}

// Describes a message returned as part of a prompt.
//
// This is similar to `SamplingMessage`, but also supports the embedding of
// resources from the MCP server.
type PromptMessage struct {
	// Content corresponds to the JSON schema field "content".
	Content Content `json:"content" yaml:"content" mapstructure:"content"`

	// Role corresponds to the JSON schema field "role".
	Role Role `json:"role" yaml:"role" mapstructure:"role"`
}

// An optional notification from the server to the client, informing it that the
// list of prompts it offers has changed. This may be issued by servers without any
// previous subscription from the client.
type PromptListChangedNotification Notification
