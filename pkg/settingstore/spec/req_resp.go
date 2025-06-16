package spec

import (
	modelSpec "github.com/ppipada/flexigpt-app/pkg/model/spec"
)

type GetAllSettingsRequest struct {
	ForceFetch bool `query:"forceFetch" doc:"Force refresh the settings and get" required:"false"`
}

type GetAllSettingsResponse struct {
	Body *SettingsSchema
}

type SetSettingRequestBody struct {
	Value any `json:"value" required:"true" doc:"Value to be set"`
}

type SetSettingRequest struct {
	Key  string `path:"key" doc:"a dot separated setting key"`
	Body *SetSettingRequestBody
}

type SetSettingResponse struct{}

type SetAppSettingsRequest struct {
	Body *AppSettings
}

type SetAppSettingsResponse struct{}

type AddAISettingRequest struct {
	ProviderName modelSpec.ProviderName `path:"providerName"`
	Body         *AISetting
}
type AddAISettingResponse struct{}

type DeleteAISettingRequest struct {
	ProviderName modelSpec.ProviderName `path:"providerName"`
}
type DeleteAISettingResponse struct{}

type SetAISettingAPIKeyRequest struct {
	ProviderName modelSpec.ProviderName `path:"providerName"`
	Body         *SetAISettingAPIKeyRequestBody
}

type SetAISettingAPIKeyRequestBody struct {
	APIKey string `json:"apiKey"`
}

type SetAISettingAPIKeyResponse struct{}

type SetAISettingAttrsRequest struct {
	ProviderName modelSpec.ProviderName `path:"providerName"`
	Body         *SetAISettingAttrsRequestBody
}

type SetAISettingAttrsRequestBody struct {
	IsEnabled                *bool                `json:"isEnabled,omitempty"`
	Origin                   *string              `json:"origin,omitempty"`
	ChatCompletionPathPrefix *string              `json:"chatCompletionPathPrefix,omitempty"`
	DefaultModel             *modelSpec.ModelName `json:"defaultModel,omitempty"`
}

type SetAISettingAttrsResponse struct{}

type AddModelPresetRequest struct {
	ProviderName modelSpec.ProviderName `path:"providerName"`
	ModelName    modelSpec.ModelName    `path:"modelName"`
	Body         *modelSpec.ModelPreset
}

type AddModelPresetResponse struct{}

type DeleteModelPresetRequest struct {
	ProviderName modelSpec.ProviderName `path:"providerName"`
	ModelName    modelSpec.ModelName    `path:"modelName"`
}

type DeleteModelPresetResponse struct{}
