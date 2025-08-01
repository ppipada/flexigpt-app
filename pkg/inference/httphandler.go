package inference

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

const (
	tag        = "ProviderSet"
	pathPrefix = "/providerset"
)

func InitProviderSetHandlers(api huma.API, providerSetAPI *ProviderSetAPI) {
	huma.Register(api, huma.Operation{
		OperationID: "add-provider",
		Method:      http.MethodPut,
		Path:        pathPrefix + "/providers/{provider}",
		Summary:     "Add provider",
		Description: "Add provider",
		Tags:        []string{tag},
	}, providerSetAPI.AddProvider)

	huma.Register(api, huma.Operation{
		OperationID: "delete-provider",
		Method:      http.MethodDelete,
		Path:        pathPrefix + "/providers/{provider}",
		Summary:     "Delete provider",
		Description: "Delete provider",
		Tags:        []string{tag},
	}, providerSetAPI.DeleteProvider)

	huma.Register(api, huma.Operation{
		OperationID: "set-provider-apikey",
		Method:      http.MethodPatch,
		Path:        pathPrefix + "/providers/{provider}/apikey",
		Summary:     "Set provider apikey",
		Description: "Set provider apikey",
		Tags:        []string{tag},
	}, providerSetAPI.SetProviderAPIKey)

	huma.Register(api, huma.Operation{
		OperationID: "fetch-provider-completion",
		Method:      http.MethodPost,
		Path:        pathPrefix + "/providers/{provider}/completion",
		Summary:     "Fetch completion for a provider",
		Description: "Fetch completion for a provider",
		Tags:        []string{tag},
	}, providerSetAPI.FetchCompletion)
}
