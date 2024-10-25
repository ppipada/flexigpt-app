package settingstore

import (
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

const tag = "Settings"
const pathPrefix = "/settings"

func InitSettingStoreHandlers(api huma.API, settingsStoreAPI *SettingStore) {
	// Register GET /settings
	huma.Register(api, huma.Operation{
		OperationID: "get-all-settings",
		Method:      http.MethodGet,
		Path:        pathPrefix,
		Summary:     "Get all settings",
		Description: "Get full settings object",
		Tags:        []string{tag},
	}, settingsStoreAPI.GetAllSettings)

	// Register PUT /settings/{key}
	huma.Register(api, huma.Operation{
		OperationID: "set-settings",
		Method:      http.MethodPut,
		Path:        fmt.Sprintf("%s/{key}", pathPrefix),
		Summary:     "Set a setting",
		Description: "Set a setting. Key can be dot separated key",
		Tags:        []string{tag},
	}, settingsStoreAPI.SetSetting)
}
