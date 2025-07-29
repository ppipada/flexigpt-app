package spec

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

type SetDefaultModelPresetRequestBody struct {
	ModelPresetID ModelPresetID `path:"modelPresetID" required:"true"`
}
type SetDefaultModelPresetRequest struct {
	ProviderName ProviderName `path:"providerName"`
	Body         *SetDefaultModelPresetRequestBody
}

type SetDefaultModelPresetResponse struct{}
