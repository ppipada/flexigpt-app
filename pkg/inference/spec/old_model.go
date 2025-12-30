package spec

import toolSpec "github.com/ppipada/flexigpt-app/pkg/tool/spec"

type ReasoningContentType string

const (
	ReasoningContentTypeOpenAIResponses   ReasoningContentType = "reasoningOpenAIResponses"
	ReasoningContentTypeAnthropicMessages ReasoningContentType = "reasoningAnthropicMessages"
)

type ReasoningContentOpenAIResponses struct {
	ID      string   `json:"id"`
	Summary []string `json:"summary,omitzero"`
	Content []string `json:"content,omitzero"`
	Status  string   `json:"status,omitzero"`
	// The encrypted content of the reasoning item - populated when a response is
	// generated with `reasoning.encrypted_content` in the `include` parameter.
	EncryptedContent string `json:"encryptedContent,omitzero"`
}

type ReasoningContentAnthropicMessages struct {
	Signature        string `json:"signature"`
	Thinking         string `json:"thinking"`
	RedactedThinking string `json:"redactedThinking"`
}

type ReasoningContent struct {
	Type                     ReasoningContentType               `json:"type"`
	ContentOpenAIResponses   *ReasoningContentOpenAIResponses   `json:"contentOpenAIResponses,omitempty"`
	ContentAnthropicMessages *ReasoningContentAnthropicMessages `json:"contentAnthropicMessages,omitempty"`
}

type CitationKind string

const (
	CitationKindURLOpenAIResponses   CitationKind = "openAIResponsesURLCitation"
	CitationKindURLAnthropicMessages CitationKind = "anthropicMessagesURLCitation"
)

type URLCitationOpenAIResponses struct {
	URL        string `json:"url"`
	Title      string `json:"title"`
	StartIndex int64  `json:"startIndex"`
	EndIndex   int64  `json:"endIndex"`
}

type URLCitationAnthropicMessages struct {
	URL            string `json:"url"`
	Title          string `json:"title"`
	EncryptedIndex string `json:"encryptedIndex"`
	CitedText      string `json:"citedText"`
}

type Citation struct {
	Kind CitationKind `json:"kind"`

	// Exactly one of the following should be non-nil, depending on Kind.
	URLCitationOpenAIResponses   *URLCitationOpenAIResponses   `json:"urlCitationOpenAIResponses,omitempty"`
	URLCitationAnthropicMessages *URLCitationAnthropicMessages `json:"urlCitationAnthropicMessages,omitempty"`
}

// ToolCall captures model-requested (or model-returned) tool invocations
// so the frontend can decide how to proceed (execute locally, display results,
// etc.).
type ToolCall struct {
	ID              string                    `json:"id"`
	CallID          string                    `json:"callID"`
	Name            string                    `json:"name"`
	Arguments       string                    `json:"arguments"`
	Type            string                    `json:"type"`
	Status          string                    `json:"status,omitempty"`
	ToolStoreChoice *toolSpec.ToolStoreChoice `json:"toolStoreChoice,omitempty"`
}

type ToolOutput struct {
	ID              string                    `json:"id"`
	CallID          string                    `json:"callID"`
	Name            string                    `json:"name"`
	Summary         string                    `json:"summary,omitempty"`
	RawOutput       string                    `json:"rawOutput,omitempty"`
	ToolStoreChoice *toolSpec.ToolStoreChoice `json:"toolStoreChoice,omitempty"`
}
