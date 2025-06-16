package store

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// We keep a few constants for organization.
const (
	tag        = "ModelPresets"
	pathPrefix = "/modelpresets"
)

// InitModelPresetsStoreHandlers registers all endpoints related to settings.
func InitModelPresetsStoreHandlers(api huma.API, modelPresetsStoreAPI *ModelPresetsStore) {
	huma.Register(api, huma.Operation{
		OperationID: "get-all-model-presets",
		Method:      http.MethodGet,
		Path:        pathPrefix,
		Summary:     "Get all model presets",
		Description: "Get the entire model presets object from the store",
		Tags:        []string{tag},
	}, modelPresetsStoreAPI.GetAllModelPresets)

	huma.Register(api, huma.Operation{
		OperationID: "create-model-presets",
		Method:      http.MethodPost,
		Path:        pathPrefix + "/{providerName}",
		Summary:     "Create new model presets for a provider",
		Tags:        []string{tag},
	}, modelPresetsStoreAPI.CreateModelPresets)

	huma.Register(api, huma.Operation{
		OperationID: "delete-model-presets",
		Method:      http.MethodDelete,
		Path:        pathPrefix + "/{providerName}",
		Summary:     "Delete all model presets for a provider",
		Tags:        []string{tag},
	}, modelPresetsStoreAPI.DeleteModelPresets)

	huma.Register(api, huma.Operation{
		OperationID: "add-model-preset",
		Method:      http.MethodPut,
		Path:        pathPrefix + "/{providerName}/modelpresets/{modelName}",
		Summary:     "Add or replace a single model preset for a given provider",
		Tags:        []string{tag},
	}, modelPresetsStoreAPI.AddModelPreset)

	huma.Register(api, huma.Operation{
		OperationID: "delete-model-preset",
		Method:      http.MethodDelete,
		Path:        pathPrefix + "/{providerName}/modelpresets/{modelName}",
		Summary:     "Delete a single model preset for a given provider",
		Tags:        []string{tag},
	}, modelPresetsStoreAPI.DeleteModelPreset)
}
