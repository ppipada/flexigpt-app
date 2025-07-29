package spec

import "time"

const SchemaVersion = "2025-07-01"

type (
	ModelName         string
	ModelDisplayName  string
	ModelShortCommand string
	ModelPresetID     string
	ReasoningLevel    string
	ReasoningType     string

	ProviderName        string
	ProviderAPIType     string
	ProviderDisplayName string
)

const (
	InbuiltAnthropicCompatible   ProviderAPIType = "inbuiltAnthropicCompatible"
	InbuiltHuggingFaceCompatible ProviderAPIType = "inbuiltHuggingFaceCompatible"
	InbuiltOpenAICompatible      ProviderAPIType = "inbuiltOpenAICompatible"
	CustomOpenAICompatible       ProviderAPIType = "customOpenAICompatible"
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

// ModelPreset is the entire "model + default knobs" bundle the user can save.
// Anything not present in the preset is considered to be taken as default from any global defaults or inbuilt model defaults.
type ModelPreset struct {
	SchemaVersion string            `json:"schemaVersion" required:"true"`
	ID            ModelPresetID     `json:"id"            required:"true"`
	Name          ModelName         `json:"name"          required:"true"`
	DisplayName   ModelDisplayName  `json:"displayName"   required:"true"`
	ShortCommand  ModelShortCommand `json:"shortCommand"  required:"true"`
	IsEnabled     bool              `json:"isEnabled"     required:"true"`

	Stream                      *bool            `json:"stream,omitempty"`
	MaxPromptLength             *int             `json:"maxPromptLength,omitempty"`
	MaxOutputLength             *int             `json:"maxOutputLength,omitempty"`
	Temperature                 *float64         `json:"temperature,omitempty"`
	Reasoning                   *ReasoningParams `json:"reasoning,omitempty"`
	SystemPrompt                *string          `json:"systemPrompt,omitempty"`
	Timeout                     *int             `json:"timeout,omitempty"`
	AdditionalParametersRawJSON *string          `json:"additionalParametersRawJSON,omitempty"`

	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
	IsBuiltIn  bool      `json:"isBuiltIn"`
}

type ProviderPreset struct {
	SchemaVersion string              `json:"schemaVersion" required:"true"`
	Name          ProviderName        `json:"name"          required:"true"`
	DisplayName   ProviderDisplayName `json:"displayName"   required:"true"`
	APIType       ProviderAPIType     `json:"apiType"`
	IsEnabled     bool                `json:"isEnabled"     required:"true"`

	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
	IsBuiltIn  bool      `json:"isBuiltIn"`

	Origin                   string            `json:"origin"                   required:"true"`
	ChatCompletionPathPrefix string            `json:"chatCompletionPathPrefix" required:"true"`
	APIKeyHeaderKey          string            `json:"apiKeyHeaderKey"`
	DefaultHeaders           map[string]string `json:"defaultHeaders"`

	DefaultModelPresetID ModelPresetID                 `json:"defaultModelPresetID"`
	ModelPresets         map[ModelPresetID]ModelPreset `json:"modelPresets"`
}

type PresetsSchema struct {
	Version         string                          `json:"version"`
	ProviderPresets map[ProviderName]ProviderPreset `json:"providerPresets"`
}
