package settingstore

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// We keep a few constants for organization.
const (
	tag        = "Settings"
	pathPrefix = "/settings"
)

// InitSettingStoreHandlers registers all endpoints related to settings.
func InitSettingStoreHandlers(api huma.API, settingsStoreAPI *SettingStore) {
	// GET /settings
	huma.Register(api, huma.Operation{
		OperationID: "get-all-settings",
		Method:      http.MethodGet,
		Path:        pathPrefix,
		Summary:     "Get all settings",
		Description: "Get the entire settings object from the store",
		Tags:        []string{tag},
	}, settingsStoreAPI.GetAllSettings)

	huma.Register(api, huma.Operation{
		OperationID: "set-app-settings",
		Method:      http.MethodPut,
		Path:        pathPrefix + "/app",
		Summary:     "Set app-level settings",
		Description: "Fully replace or update the 'app' portion of the settings",
		Tags:        []string{tag},
	}, settingsStoreAPI.SetAppSettings)

	huma.Register(api, huma.Operation{
		OperationID: "add-ai-setting",
		Method:      http.MethodPost,
		Path:        pathPrefix + "/aisettings/{providerName}",
		Summary:     "Create/add a new AI Setting",
		Tags:        []string{tag},
	}, settingsStoreAPI.AddAISetting)

	huma.Register(api, huma.Operation{
		OperationID: "delete-ai-setting",
		Method:      http.MethodDelete,
		Path:        pathPrefix + "/aisettings/{providerName}",
		Summary:     "Delete an AI provider",
		Tags:        []string{tag},
	}, settingsStoreAPI.DeleteAISetting)

	huma.Register(api, huma.Operation{
		OperationID: "set-ai-setting-apikey",
		Method:      http.MethodPut,
		Path:        pathPrefix + "/aisettings/{providerName}/apiKey",
		Summary:     "Set API key for a given AI provider",
		Tags:        []string{tag},
	}, settingsStoreAPI.SetAISettingAPIKey)

	huma.Register(api, huma.Operation{
		OperationID: "set-ai-setting-attrs",
		Method:      http.MethodPatch,
		Path:        pathPrefix + "/aisettings/{providerName}",
		Summary:     "Partially update certain AI provider attributes",
		Tags:        []string{tag},
	}, settingsStoreAPI.SetAISettingAttrs)

	huma.Register(api, huma.Operation{
		OperationID: "add-model-setting",
		Method:      http.MethodPut,
		Path:        pathPrefix + "/aisettings/{providerName}/modelsettings/{modelName}",
		Summary:     "Add or replace a single model setting for a given AI provider",
		Tags:        []string{tag},
	}, settingsStoreAPI.AddModelSetting)

	huma.Register(api, huma.Operation{
		OperationID: "delete-model-setting",
		Method:      http.MethodDelete,
		Path:        pathPrefix + "/aisettings/{providerName}/modelsettings/{modelName}",
		Summary:     "Delete a single model setting for a given AI provider",
		Tags:        []string{tag},
	}, settingsStoreAPI.DeleteModelSetting)
}
