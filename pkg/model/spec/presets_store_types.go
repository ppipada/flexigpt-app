package spec

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

type ProviderModelPresets map[ModelName]ModelPreset

type ModelPresetsSchema struct {
	Version      string                                `json:"version"`
	ModelPresets map[ProviderName]ProviderModelPresets `json:"modelPresets"`
}

type GetAllModelPresetsRequest struct {
	ForceFetch bool `query:"forceFetch" doc:"Force refresh the model presets and get" required:"false"`
}

type GetAllModelPresetsResponse struct {
	Body *ModelPresetsSchema
}

type CreateModelPresetsRequest struct {
	ProviderName ProviderName `path:"providerName"`
	Body         *ProviderModelPresets
}
type CreateModelPresetsResponse struct{}

type DeleteModelPresetsRequest struct {
	ProviderName ProviderName `path:"providerName"`
}
type DeleteModelPresetsResponse struct{}

type AddModelPresetRequest struct {
	ProviderName ProviderName `path:"providerName"`
	ModelName    ModelName    `path:"modelName"`
	Body         *ModelPreset
}
type AddModelPresetResponse struct{}

type DeleteModelPresetRequest struct {
	ProviderName ProviderName `path:"providerName"`
	ModelName    ModelName    `path:"modelName"`
}
type DeleteModelPresetResponse struct{}
