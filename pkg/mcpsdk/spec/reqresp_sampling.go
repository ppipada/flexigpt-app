package spec

import jsonrpcReqResp "github.com/flexigpt/flexiui/pkg/jsonrpc/reqresp"

// Describes a message issued to or received from an LLM API.
type SamplingMessage struct {
	Content Content `json:"content"`
	Role    Role    `json:"role"`
}

// A request from the server to sample an LLM via the client. The client has full
// discretion over which model to select. The client should also inform the user
// before beginning sampling, to allow them to inspect the request (human in the
// loop) and decide whether to approve it.
type CreateMessageRequest jsonrpcReqResp.Request[CreateMessageRequestParams]

type CreateMessageRequestParams struct {
	// This property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their requests.
	Meta map[string]any `json:"_meta,omitempty"`
	_    struct{}       `json:"-"               additionalProperties:"true"`

	Messages []SamplingMessage `json:"messages"`

	// The server's preferences for which model to select. The client MAY ignore these
	// preferences.
	ModelPreferences *ModelPreferences `json:"modelPreferences,omitempty"`

	// An optional system prompt the server wants to use for sampling. The client MAY
	// modify or omit this prompt.
	SystemPrompt *string `json:"systemPrompt,omitempty"`

	// A request to include context from one or more MCP servers (including the
	// caller), to be attached to the prompt. The client MAY ignore this request.
	IncludeContext *IncludeContext `json:"includeContext,omitempty"`

	// Temperature corresponds to the JSON schema field "temperature".
	Temperature *float64 `json:"temperature,omitempty"`

	// The maximum number of tokens to sample, as requested by the server. The client
	// MAY choose to sample fewer tokens than requested.
	MaxTokens int `json:"maxTokens"`

	StopSequences []string `json:"stopSequences,omitempty"`

	// Optional metadata to pass through to the LLM provider. The format of this
	// metadata is provider-specific.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// The client's response to a sampling/create_message request from the server. The
// client should inform the user before returning the sampled message, to allow
// them to inspect the response (human in the loop) and decide whether to allow the
// server to see it.
type (
	CreateMessageResponse jsonrpcReqResp.Response[*CreateMessageResult]
	CreateMessageResult   struct {
		// This result property is reserved by the protocol to allow clients and servers
		// to attach additional metadata to their responses.
		Meta map[string]any `json:"_meta,omitempty"`
		_    struct{}       `json:"-"               additionalProperties:"true"`

		Content Content `json:"content"`
		Role    Role    `json:"role"`

		// The name of the model that generated the message.
		Model string `json:"model"`

		// The reason why sampling stopped, if known.
		StopReason *string `json:"stopReason,omitempty"`
	}
)

// The server's preferences for model selection, requested of the client during
// sampling.
//
// Because LLMs can vary along multiple dimensions, choosing the "best" model is
// rarely straightforward.  Different models excel in different areas—some are
// faster but less capable, others are more capable but more expensive, and so
// on. This interface allows servers to express their priorities across multiple
// dimensions to help clients make an appropriate selection for their use case.
//
// These preferences are always advisory. The client MAY ignore them. It is also
// up to the client to decide how to interpret these preferences and how to
// balance them against other considerations.
type ModelPreferences struct {
	// How much to prioritize cost when selecting a model. A value of 0 means cost
	// is not important, while a value of 1 means cost is the most important
	// factor.
	CostPriority *float64 `json:"costPriority,omitempty"`

	// Optional hints to use for model selection.
	//
	// If multiple hints are specified, the client MUST evaluate them in order
	// (such that the first match is taken).
	//
	// The client SHOULD prioritize these hints over the numeric priorities, but
	// MAY still use the priorities to select from ambiguous matches.
	Hints []ModelHint `json:"hints,omitempty"`

	// How much to prioritize intelligence and capabilities when selecting a
	// model. A value of 0 means intelligence is not important, while a value of 1
	// means intelligence is the most important factor.
	IntelligencePriority *float64 `json:"intelligencePriority,omitempty"`

	// How much to prioritize sampling speed (latency) when selecting a model. A
	// value of 0 means speed is not important, while a value of 1 means speed is
	// the most important factor.
	SpeedPriority *float64 `json:"speedPriority,omitempty"`
}

// Hints to use for model selection.
//
// Keys not declared here are currently left unspecified by the spec and are up
// to the client to interpret.
type ModelHint struct {
	// A hint for a model name.
	//
	// The client SHOULD treat this as a substring of a model name; for example:
	//  - `claude-3-5-sonnet` should match `claude-3-5-sonnet-20241022`
	//  - `sonnet` should match `claude-3-5-sonnet-20241022`,
	// `claude-3-sonnet-20240229`, etc.
	//  - `claude` should match any Claude model
	//
	// The client MAY also map the string to a different provider's model name or a
	// different model family, as long as it fills a similar niche; for example:
	//  - `gemini-1.5-flash` could match `claude-3-haiku-20240307`
	Name *string `json:"name,omitempty"`
}
