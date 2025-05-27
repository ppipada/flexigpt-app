package spec

import jsonrpcReqResp "github.com/ppipada/flexigpt-app/pkg/jsonrpc/reqresp"

// Sent from the client to request a list of prompts and prompt templates the
// server has.
type ListPromptsRequest jsonrpcReqResp.Request[*PaginatedRequestParams]

// The server's response to a prompts/list request from the client.
type ListPromptsResponse jsonrpcReqResp.Response[*ListPromptsResult]

type ListPromptsResult struct {
	_ struct{} `json:"-"       additionalProperties:"true"`
	PaginatedResultParams
	Prompts []Prompt `json:"prompts"`
}

// A prompt or prompt template that the server offers.
type Prompt struct {
	// The name of the prompt or prompt template.
	Name string `json:"name"`

	// An optional description of what this prompt provides
	Description *string `json:"description,omitempty"`

	// A list of arguments to use for templating the prompt.
	Arguments []PromptArgument `json:"arguments,omitempty"`
}

// Describes an argument that a prompt can accept.
type PromptArgument struct {
	// The name of the argument.
	Name string `json:"name"`

	// A human-readable description of the argument.
	Description *string `json:"description,omitempty"`

	// Whether this argument must be provided.
	Required *bool `json:"required,omitempty"`
}

// Used by the client to get a prompt provided by the server.
type GetPromptRequest jsonrpcReqResp.Request[GetPromptRequestParams]

type GetPromptRequestParams struct {
	// This property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their requests.
	Meta map[string]any `json:"_meta,omitempty"`
	_    struct{}       `json:"-"               additionalProperties:"true"`

	Name string `json:"name"`

	// Arguments to use for templating the prompt.
	Arguments map[string]string `json:"arguments,omitempty"`
}

// The server's response to a prompts/get request from the client.
type (
	GetPromptResponse jsonrpcReqResp.Response[*GetPromptResult]
	GetPromptResult   struct {
		// This result property is reserved by the protocol to allow clients and servers
		// to attach additional metadata to their responses.
		Meta        map[string]any  `json:"_meta,omitempty"`
		Description *string         `json:"description,omitempty"`
		Messages    []PromptMessage `json:"messages"`
	}
)

// Describes a message returned as part of a prompt.
//
// This is similar to `SamplingMessage`, but also supports the embedding of
// resources from the MCP server.
type PromptMessage struct {
	Role    Role    `json:"role"`
	Content Content `json:"content"`
}

// An optional notification from the server to the client, informing it that the
// list of prompts it offers has changed. This may be issued by servers without any
// previous subscription from the client.
type PromptListChangedNotification jsonrpcReqResp.Notification[*AdditionalParams]
