package spec

type (
	ModelName         string
	ModelDisplayName  string
	ModelShortCommand string
	ReasoningLevel    string
	ReasoningType     string
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

// ModelParams represents input information about a model to a completion.
type ModelParams struct {
	Name                 ModelName        `json:"name"`
	Stream               bool             `json:"stream"`
	MaxPromptLength      int              `json:"maxPromptLength"`
	MaxOutputLength      int              `json:"maxOutputLength"`
	Temperature          *float64         `json:"temperature,omitempty"`
	Reasoning            *ReasoningParams `json:"reasoning,omitempty"`
	SystemPrompt         string           `json:"systemPrompt"`
	Timeout              int              `json:"timeout"`
	AdditionalParameters map[string]any   `json:"additionalParameters"`
}

// ModelPreset is the entire “model + default knobs” bundle the user can save.
// Anything not present in the preset is considered to be taken as default from any global defaults or inbuilt model defaults.
type ModelPreset struct {
	Name         ModelName         `json:"name"         required:"true"`
	DisplayName  ModelDisplayName  `json:"displayName"  required:"true"`
	ShortCommand ModelShortCommand `json:"shortCommand" required:"true"`
	IsEnabled    bool              `json:"isEnabled"    required:"true"`

	Stream               *bool            `json:"stream,omitempty"`
	MaxPromptLength      *int             `json:"maxPromptLength,omitempty"`
	MaxOutputLength      *int             `json:"maxOutputLength,omitempty"`
	Temperature          *float64         `json:"temperature,omitempty"`
	Reasoning            *ReasoningParams `json:"reasoning,omitempty"`
	SystemPrompt         *string          `json:"systemPrompt,omitempty"`
	Timeout              *int             `json:"timeout,omitempty"`
	AdditionalParameters map[string]any   `json:"additionalParameters,omitempty"`
}
