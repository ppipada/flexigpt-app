package store

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// We keep a few constants for organization.
const (
	tag        = "ModelPresetStore"
	pathPrefix = "/modelpresetstore"
)

// InitModelPresetStoreHandlers registers all endpoints related to settings.
func InitModelPresetStoreHandlers(api huma.API, modelPresetStoreAPI *ModelPresetStore) {
	huma.Register(api, huma.Operation{
		OperationID: "get-all-model-presets",
		Method:      http.MethodGet,
		Path:        pathPrefix,
		Summary:     "Get all model presets",
		Description: "Get the entire model presets object from the store",
		Tags:        []string{tag},
	}, modelPresetStoreAPI.GetAllModelPresets)

	huma.Register(api, huma.Operation{
		OperationID: "create-provider-preset",
		Method:      http.MethodPost,
		Path:        pathPrefix + "/{providerName}",
		Summary:     "Create new model presets for a provider",
		Tags:        []string{tag},
	}, modelPresetStoreAPI.CreateProviderPreset)

	huma.Register(api, huma.Operation{
		OperationID: "delete-provider-preset",
		Method:      http.MethodDelete,
		Path:        pathPrefix + "/{providerName}",
		Summary:     "Delete all model presets for a provider",
		Tags:        []string{tag},
	}, modelPresetStoreAPI.DeleteProviderPreset)

	huma.Register(api, huma.Operation{
		OperationID: "set-default-model-preset",
		Method:      http.MethodPut,
		Path:        pathPrefix + "/{providerName}/default",
		Summary:     "Set the default model preset for a provider",
		Tags:        []string{tag},
	}, modelPresetStoreAPI.SetDefaultModelPreset)

	huma.Register(api, huma.Operation{
		OperationID: "add-model-preset",
		Method:      http.MethodPut,
		Path:        pathPrefix + "/{providerName}/modelpresets/{modelName}",
		Summary:     "Add or replace a single model preset for a given provider",
		Tags:        []string{tag},
	}, modelPresetStoreAPI.AddModelPreset)

	huma.Register(api, huma.Operation{
		OperationID: "delete-model-preset",
		Method:      http.MethodDelete,
		Path:        pathPrefix + "/{providerName}/modelpresets/{modelName}",
		Summary:     "Delete a single model preset for a given provider",
		Tags:        []string{tag},
	}, modelPresetStoreAPI.DeleteModelPreset)
}
