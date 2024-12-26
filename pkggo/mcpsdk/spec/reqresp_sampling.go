package spec

// Describes a message issued to or received from an LLM API.
type SamplingMessage struct {
	// Content corresponds to the JSON schema field "content".
	Content Content `json:"content" yaml:"content" mapstructure:"content"`

	// Role corresponds to the JSON schema field "role".
	Role Role `json:"role" yaml:"role" mapstructure:"role"`
}

// A request from the server to sample an LLM via the client. The client has full
// discretion over which model to select. The client should also inform the user
// before beginning sampling, to allow them to inspect the request (human in the
// loop) and decide whether to approve it.
type CreateMessageRequest struct {
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params CreateMessageRequestParams `json:"params" yaml:"params" mapstructure:"params"`
}

type CreateMessageRequestParams struct {
	// A request to include context from one or more MCP servers (including the
	// caller), to be attached to the prompt. The client MAY ignore this request.
	IncludeContext *IncludeContext `json:"includeContext,omitempty" yaml:"includeContext,omitempty" mapstructure:"includeContext,omitempty"`

	// The maximum number of tokens to sample, as requested by the server. The client
	// MAY choose to sample fewer tokens than requested.
	MaxTokens int `json:"maxTokens" yaml:"maxTokens" mapstructure:"maxTokens"`

	// Messages corresponds to the JSON schema field "messages".
	Messages []SamplingMessage `json:"messages" yaml:"messages" mapstructure:"messages"`

	// Optional metadata to pass through to the LLM provider. The format of this
	// metadata is provider-specific.
	Metadata CreateMessageRequestParamsMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty" mapstructure:"metadata,omitempty"`

	// The server's preferences for which model to select. The client MAY ignore these
	// preferences.
	ModelPreferences *ModelPreferences `json:"modelPreferences,omitempty" yaml:"modelPreferences,omitempty" mapstructure:"modelPreferences,omitempty"`

	// StopSequences corresponds to the JSON schema field "stopSequences".
	StopSequences []string `json:"stopSequences,omitempty" yaml:"stopSequences,omitempty" mapstructure:"stopSequences,omitempty"`

	// An optional system prompt the server wants to use for sampling. The client MAY
	// modify or omit this prompt.
	SystemPrompt *string `json:"systemPrompt,omitempty" yaml:"systemPrompt,omitempty" mapstructure:"systemPrompt,omitempty"`

	// Temperature corresponds to the JSON schema field "temperature".
	Temperature *float64 `json:"temperature,omitempty" yaml:"temperature,omitempty" mapstructure:"temperature,omitempty"`
}

// Optional metadata to pass through to the LLM provider. The format of this
// metadata is provider-specific.
type CreateMessageRequestParamsMetadata map[string]interface{}

// The client's response to a sampling/create_message request from the server. The
// client should inform the user before returning the sampled message, to allow
// them to inspect the response (human in the loop) and decide whether to allow the
// server to see it.
type CreateMessageResult struct {
	SamplingMessage

	_ struct{} `json:"-" additionalProperties:"true"`

	// This result property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their responses.
	Meta map[string]interface{} `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`

	// The name of the model that generated the message.
	Model string `json:"model" yaml:"model" mapstructure:"model"`

	// The reason why sampling stopped, if known.
	StopReason *string `json:"stopReason,omitempty" yaml:"stopReason,omitempty" mapstructure:"stopReason,omitempty"`
}

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
	CostPriority *float64 `json:"costPriority,omitempty" yaml:"costPriority,omitempty" mapstructure:"costPriority,omitempty"`

	// Optional hints to use for model selection.
	//
	// If multiple hints are specified, the client MUST evaluate them in order
	// (such that the first match is taken).
	//
	// The client SHOULD prioritize these hints over the numeric priorities, but
	// MAY still use the priorities to select from ambiguous matches.
	Hints []ModelHint `json:"hints,omitempty" yaml:"hints,omitempty" mapstructure:"hints,omitempty"`

	// How much to prioritize intelligence and capabilities when selecting a
	// model. A value of 0 means intelligence is not important, while a value of 1
	// means intelligence is the most important factor.
	IntelligencePriority *float64 `json:"intelligencePriority,omitempty" yaml:"intelligencePriority,omitempty" mapstructure:"intelligencePriority,omitempty"`

	// How much to prioritize sampling speed (latency) when selecting a model. A
	// value of 0 means speed is not important, while a value of 1 means speed is
	// the most important factor.
	SpeedPriority *float64 `json:"speedPriority,omitempty" yaml:"speedPriority,omitempty" mapstructure:"speedPriority,omitempty"`
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
	Name *string `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name,omitempty"`
}

// A request from the client to the server, to ask for completion options.
type CompleteRequest struct {
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params CompleteRequestParams `json:"params" yaml:"params" mapstructure:"params"`
}

type CompleteRequestParams struct {
	// The argument's information
	Argument CompleteRequestParamsArgument `json:"argument" yaml:"argument" mapstructure:"argument"`

	// Ref corresponds to the JSON schema field "ref".
	Ref CompletionReference `json:"ref" yaml:"ref" mapstructure:"ref"`
}

// The argument's information
type CompleteRequestParamsArgument struct {
	// The name of the argument
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// The value of the argument to use for completion matching.
	Value string `json:"value" yaml:"value" mapstructure:"value"`
}

// Identifies a completion.
type CompletionReference struct {
	// Type corresponds to the JSON schema field "type".
	Type Ref `json:"type" yaml:"type" mapstructure:"type"`
	// The name of the prompt or prompt template
	Name *string `json:"name" yaml:"name" mapstructure:"name"`
	// The URI or URI template of the resource.
	Uri *string `json:"uri"  yaml:"uri"  mapstructure:"uri"`
}

// The server's response to a completion/complete request
type CompleteResult struct {
	// This result property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their responses.
	Meta map[string]interface{} `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`

	// Completion corresponds to the JSON schema field "completion".
	Completion CompleteResultCompletion `json:"completion" yaml:"completion" mapstructure:"completion"`
}

type CompleteResultCompletion struct {
	// Indicates whether there are additional completion options beyond those provided
	// in the current response, even if the exact total is unknown.
	HasMore *bool `json:"hasMore,omitempty" yaml:"hasMore,omitempty" mapstructure:"hasMore,omitempty"`

	// The total number of completion options available. This can exceed the number of
	// values actually sent in the response.
	Total *int `json:"total,omitempty" yaml:"total,omitempty" mapstructure:"total,omitempty"`

	// An array of completion values. Must not exceed 100 items.
	Values []string `json:"values" yaml:"values" mapstructure:"values"`
}
