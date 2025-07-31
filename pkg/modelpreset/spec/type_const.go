package spec

import (
	"errors"
	"time"
)

const (
	ModelPresetsPresetsFile            = "modelpresets.json" // Single JSON file.
	ModelPresetsBuiltInOverlayFileName = "modelpresetsbuiltin.overlay.json"

	OpenAICompatibleAPIKeyHeaderKey          = "Authorization"
	OpenAICompatibleChatCompletionPathPrefix = "/v1/chat/completions"

	SchemaVersion = "2025-07-01"

	MaxPageSize           = 256 // Max allowed page size.
	DefaultPageSize       = 256 // Default page size.
	BuiltInSnapshotMaxAge = time.Hour
)

var OpenAICompatibleDefaultHeaders = map[string]string{"content-type": "application/json"}

var (
	ErrInvalidDir = errors.New("invalid directory")

	// Look-ups.
	ErrProviderNotFound      = errors.New("provider not found")
	ErrModelPresetNotFound   = errors.New("model preset not found")
	ErrBuiltInProviderAbsent = errors.New("provider not found in built-in data")

	// Validation.
	ErrNilProvider      = errors.New("provider preset is nil")
	ErrNilModelPreset   = errors.New("model preset is nil")
	ErrNoModelPresets   = errors.New("provider has no model presets")
	ErrInvalidTimestamp = errors.New("zero timestamp")
	ErrBuiltInReadOnly  = errors.New("built-in resource is read-only")
)

type (
	ModelName        string
	ModelDisplayName string
	ModelSlug        string
	ModelPresetID    string
	ReasoningLevel   string
	ReasoningType    string

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
	SchemaVersion string           `json:"schemaVersion" required:"true"`
	ID            ModelPresetID    `json:"id"            required:"true"`
	Name          ModelName        `json:"name"          required:"true"`
	DisplayName   ModelDisplayName `json:"displayName"   required:"true"`
	Slug          ModelSlug        `json:"slug"          required:"true"`
	IsEnabled     bool             `json:"isEnabled"     required:"true"`

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
	APIKeyHeaderKey          string            `json:"apiKeyHeaderKey"          required:"true"`
	DefaultHeaders           map[string]string `json:"defaultHeaders"`

	DefaultModelPresetID ModelPresetID                 `json:"defaultModelPresetID"`
	ModelPresets         map[ModelPresetID]ModelPreset `json:"modelPresets"`
}

type PresetsSchema struct {
	Version         string                          `json:"version"`
	ProviderPresets map[ProviderName]ProviderPreset `json:"providerPresets"`
}
