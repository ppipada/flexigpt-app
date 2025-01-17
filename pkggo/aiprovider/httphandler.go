package aiprovider

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

const tag = "LLMProviders"
const pathPrefix = "/llmproviders"

func InitProviderSetHandlers(api huma.API, providerSetAPI *ProviderSetAPI) {

	huma.Register(api, huma.Operation{
		OperationID: "get-default-provider",
		Method:      http.MethodGet,
		Path:        pathPrefix + "/default",
		Summary:     "Get default provider",
		Description: "Get default provider",
		Tags:        []string{tag},
	}, providerSetAPI.GetDefaultProvider)

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
		OperationID: "set-provider-attributes",
		Method:      http.MethodPatch,
		Path:        pathPrefix + "/{provider}",
		Summary:     "Set provider attributes",
		Description: "Set provider attributes",
		Tags:        []string{tag},
	}, providerSetAPI.SetProviderAttribute)

	huma.Register(api, huma.Operation{
		OperationID: "make-provider-completion-req",
		Method:      http.MethodPost,
		Path:        pathPrefix + "/{provider}/completion/makereq",
		Summary:     "Make completion request",
		Description: "Make completion request for a provider",
		Tags:        []string{tag},
	}, providerSetAPI.MakeCompletion)

	huma.Register(api, huma.Operation{
		OperationID: "fetch-provider-completion",
		Method:      http.MethodPost,
		Path:        pathPrefix + "/{provider}/completion",
		Summary:     "Fetch completion for a provider",
		Description: "Fetch completion for a provider",
		Tags:        []string{tag},
	}, providerSetAPI.FetchCompletion)

}
