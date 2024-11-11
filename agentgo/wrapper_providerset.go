package main

import (
	"context"

	"github.com/flexigpt/flexiui/pkggo/aiprovider"
	aiproviderSpec "github.com/flexigpt/flexiui/pkggo/aiprovider/spec"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ProviderSetWrapper struct {
	providersetAPI *aiprovider.ProviderSetAPI
	appContext     context.Context
}

// NewProviderSetWrapper creates a new ProviderSet with the specified default provider
func InitProviderSetWrapper(
	ps *ProviderSetWrapper,
	defaultProvider aiproviderSpec.ProviderName,
) error {
	p, err := aiprovider.NewProviderSetAPI(defaultProvider)
	if err != nil {
		panic("Invalid default provider")
	}
	ps.providersetAPI = p
	return nil
}

func SetWrappedProviderAppContext(w *ProviderSetWrapper, ctx context.Context) {
	w.appContext = ctx
}

func (w *ProviderSetWrapper) GetDefaultProvider(
	req *aiproviderSpec.GetDefaultProviderRequest,
) (*aiproviderSpec.GetDefaultProviderResponse, error) {
	return w.providersetAPI.GetDefaultProvider(context.Background(), req)
}

func (w *ProviderSetWrapper) SetDefaultProvider(
	req *aiproviderSpec.SetDefaultProviderRequest,
) (*aiproviderSpec.SetDefaultProviderResponse, error) {
	return w.providersetAPI.SetDefaultProvider(context.Background(), req)
}

func (w *ProviderSetWrapper) GetConfigurationInfo(
	req *aiproviderSpec.GetConfigurationInfoRequest,
) (*aiproviderSpec.GetConfigurationInfoResponse, error) {
	return w.providersetAPI.GetConfigurationInfo(context.Background(), req)
}

func (w *ProviderSetWrapper) SetProviderAttribute(
	req *aiproviderSpec.SetProviderAttributeRequest,
) (*aiproviderSpec.SetProviderAttributeResponse, error) {
	return w.providersetAPI.SetProviderAttribute(context.Background(), req)
}

func (w *ProviderSetWrapper) MakeCompletion(
	req *aiproviderSpec.MakeCompletionRequest,
) (*aiproviderSpec.MakeCompletionResponse, error) {
	return w.providersetAPI.MakeCompletion(context.Background(), req)
}

// FetchCompletion handles the completion request and streams data back to the frontend
func (w *ProviderSetWrapper) FetchCompletion(
	provider string,
	input aiproviderSpec.CompletionRequest,
	callbackId string,
) (*aiproviderSpec.FetchCompletionResponse, error) {
	onStreamData := func(data string) error {
		runtime.EventsEmit(w.appContext, callbackId, data)
		return nil
	}

	req := &aiproviderSpec.FetchCompletionRequest{
		Body: &aiproviderSpec.FetchCompletionRequestBody{
			Provider:     aiproviderSpec.ProviderName(provider),
			Input:        &input,
			OnStreamData: onStreamData,
		},
	}
	resp, err := w.providersetAPI.FetchCompletion(
		context.Background(),
		req,
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
