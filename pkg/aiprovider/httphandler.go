package aiprovider

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

const (
	tag        = "LLMProviders"
	pathPrefix = "/llmproviders"
)

func InitProviderSetHandlers(api huma.API, providerSetAPI *ProviderSetAPI) {
	huma.Register(api, huma.Operation{
		OperationID: "set-default-provider",
		Method:      http.MethodPut,
		Path:        pathPrefix + "/default",
		Summary:     "Set default provider",
		Description: "Set default provider",
		Tags:        []string{tag},
	}, providerSetAPI.SetDefaultProvider)

	huma.Register(api, huma.Operation{
		OperationID: "get-allprovider-configuration",
		Method:      http.MethodGet,
		Path:        pathPrefix + "/allconfig",
		Summary:     "Get all config",
		Description: "Get configuration info for all providers",
		Tags:        []string{tag},
	}, providerSetAPI.GetConfigurationInfo)

	huma.Register(api, huma.Operation{
		OperationID: "add-provider",
		Method:      http.MethodPut,
		Path:        pathPrefix + "/{provider}",
		Summary:     "Add provider",
		Description: "Add provider",
		Tags:        []string{tag},
	}, providerSetAPI.AddProvider)

	huma.Register(api, huma.Operation{
		OperationID: "delete-provider",
		Method:      http.MethodDelete,
		Path:        pathPrefix + "/{provider}",
		Summary:     "Delete provider",
		Description: "Delete provider",
		Tags:        []string{tag},
	}, providerSetAPI.DeleteProvider)

	huma.Register(api, huma.Operation{
		OperationID: "set-provider-attributes",
		Method:      http.MethodPatch,
		Path:        pathPrefix + "/{provider}",
		Summary:     "Set provider attributes",
		Description: "Set provider attributes",
		Tags:        []string{tag},
	}, providerSetAPI.SetProviderAttribute)

	huma.Register(api, huma.Operation{
		OperationID: "fetch-provider-completion",
		Method:      http.MethodPost,
		Path:        pathPrefix + "/{provider}/completion",
		Summary:     "Fetch completion for a provider",
		Description: "Fetch completion for a provider",
		Tags:        []string{tag},
	}, providerSetAPI.FetchCompletion)
}
