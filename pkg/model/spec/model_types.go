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
	Name                        ModelName        `json:"name"`
	Stream                      bool             `json:"stream"`
	MaxPromptLength             int              `json:"maxPromptLength"`
	MaxOutputLength             int              `json:"maxOutputLength"`
	Temperature                 *float64         `json:"temperature,omitempty"`
	Reasoning                   *ReasoningParams `json:"reasoning,omitempty"`
	SystemPrompt                string           `json:"systemPrompt"`
	Timeout                     int              `json:"timeout"`
	AdditionalParametersRawJSON *string          `json:"additionalParametersRawJSON"`
}
