package spec

type ModelPresetID string

// ModelPreset is the entire “model + default knobs” bundle the user can save.
// Anything not present in the preset is considered to be taken as default from any global defaults or inbuilt model defaults.
type ModelPreset struct {
	ID           ModelPresetID     `json:"id"           required:"true"`
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

type ProviderPreset struct {
	DefaultModelPresetID ModelPresetID                 `json:"defaultModelPresetID"`
	ModelPresets         map[ModelPresetID]ModelPreset `json:"modelPresets"`
}

type PresetsSchema struct {
	Version         string                          `json:"version"`
	ProviderPresets map[ProviderName]ProviderPreset `json:"providerPresets"`
}

type GetAllModelPresetsRequest struct {
	ForceFetch bool `query:"forceFetch" doc:"Force refresh the model presets and get" required:"false"`
}

type GetAllModelPresetsResponse struct {
	Body *PresetsSchema
}

type CreateProviderPresetRequest struct {
	ProviderName ProviderName `path:"providerName" required:"true"`
	Body         *ProviderPreset
}
type CreateProviderPresetResponse struct{}

type DeleteProviderPresetRequest struct {
	ProviderName ProviderName `path:"providerName" required:"true"`
}
type DeleteProviderPresetResponse struct{}

type AddModelPresetRequest struct {
	ProviderName  ProviderName  `path:"providerName"  required:"true"`
	ModelPresetID ModelPresetID `path:"modelPresetID" required:"true"`
	Body          *ModelPreset
}
type AddModelPresetResponse struct{}

type DeleteModelPresetRequest struct {
	ProviderName  ProviderName  `path:"providerName"  required:"true"`
	ModelPresetID ModelPresetID `path:"modelPresetID" required:"true"`
}
type DeleteModelPresetResponse struct{}

type SetDefaultModelPresetRequest struct {
	ProviderName ProviderName `path:"providerName"`
	Body         *SetDefaultModelPresetRequestBody
}

type SetDefaultModelPresetRequestBody struct {
	ModelPresetID ModelPresetID `path:"modelPresetID" required:"true"`
}

type SetDefaultModelPresetResponse struct{}
