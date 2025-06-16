package spec

import "time"

type (
	ModelName      string
	ReasoningLevel string
	ReasoningType  string
)

const (
	ReasoningTypeHybridWithTokens ReasoningType = "hybridWithTokens"
	ReasoningTypeSingleWithLevels ReasoningType = "singleWithLevels"
)

const (
	ReasoningLevelLow    ReasoningLevel = "low"
	ReasoningLevelMedium ReasoningLevel = "medium"
	ReasoningLevelHigh   ReasoningLevel = "high"
)

type ReasoningParams struct {
	Type   ReasoningType  `json:"type"`
	Level  ReasoningLevel `json:"level"`
	Tokens int            `json:"tokens"`
}

type ModelDefaults struct {
	DisplayName string `json:"displayName"`
	IsEnabled   bool   `json:"isEnabled"`
}

// ModelParams represents input information about a model to a completion.
type ModelParams struct {
	Name                 ModelName        `json:"name"`
	Stream               bool             `json:"stream"`
	MaxPromptLength      int              `json:"maxPromptLength"`
	MaxOutputLength      int              `json:"maxOutputLength"`
	Temperature          *float64         `json:"temperature,omitempty"`
	Reasoning            *ReasoningParams `json:"reasoning"`
	SystemPrompt         string           `json:"systemPrompt"`
	Timeout              int              `json:"timeout"`
	AdditionalParameters map[string]any   `json:"additionalParameters"`
}

// Entire “model + default knobs” bundle the user can pick.
type ModelPreset struct {
	ID string `json:"id"`
	// Human label shown in UI (e.g. "Creative-4o").
	Name string `json:"name"`

	// So provider + engine live here as well.
	Provider string `json:"provider"`
	Engine   string `json:"engine"`

	// Hard limits (often published by the provider).
	MaxPromptLength int `json:"maxPromptLength"`
	MaxOutputLength int `json:"maxOutputLength"`

	// Most-used sampling knobs — pointers mean “unspecified”.
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	TopK             *int     `json:"top_k,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`

	StopSequences []string         `json:"stop,omitempty"`
	LogitBias     map[int]int      `json:"logit_bias,omitempty"`
	Reasoning     *ReasoningParams `json:"reasoning,omitempty"`

	// System prompt frequently shipped along with every call.
	SystemPrompt string `json:"systemPrompt,omitempty"`

	Stream  bool `json:"stream"`
	Timeout int  `json:"timeout"`

	// Anything provider-specific or experimental lands here.
	AdditionalParameters map[string]any `json:"additionalParameters,omitempty"`

	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}
