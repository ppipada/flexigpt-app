package aiprovider

import "github.com/flexigpt/flexiui/pkg/aiprovider/spec"

// SetAttribute updates the ProviderInfo and APICaller based on the provided parameters.
func SetAttribute(
	providerInfo *spec.ProviderInfo,
	apiKey *string,
	defaultModel *string,
	defaultTemperature *float64,
	defaultOrigin *string,
) {
	if apiKey != nil {
		providerInfo.ApiKey = *apiKey
	}
	if defaultOrigin != nil {
		providerInfo.DefaultOrigin = *defaultOrigin
	}
	if defaultModel != nil {
		providerInfo.DefaultModel = spec.ModelName(*defaultModel)
	}
	if defaultTemperature != nil {
		providerInfo.DefaultTemperature = *defaultTemperature
	}
}
