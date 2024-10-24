package main

import (
	"context"

	"github.com/flexigpt/flexiui/pkg/aiprovider"
	aiproviderSpec "github.com/flexigpt/flexiui/pkg/aiprovider/spec"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type WrappedProviderSetAPI struct {
	*aiprovider.ProviderSetAPI
	appContext context.Context
}

// NewWrappedProviderSetAPI creates a new ProviderSet with the specified default provider
func NewWrappedProviderSetAPI(
	defaultProvider aiproviderSpec.ProviderName,
) *WrappedProviderSetAPI {
	return &WrappedProviderSetAPI{
		ProviderSetAPI: aiprovider.NewProviderSetAPI(defaultProvider),
	}
}

func SetWrappedProviderAppContext(w *WrappedProviderSetAPI, ctx context.Context) {
	w.appContext = ctx
}

// FetchCompletion handles the completion request and streams data back to the frontend
func (w *WrappedProviderSetAPI) FetchCompletion(
	provider string,
	input aiproviderSpec.CompletionRequest,
	callbackId string,
) (*aiproviderSpec.CompletionResponse, error) {
	onStreamData := func(data string) error {
		runtime.EventsEmit(w.appContext, callbackId, data)
		return nil
	}

	resp, err := w.ProviderSetAPI.FetchCompletion(
		aiproviderSpec.ProviderName(provider),
		input,
		onStreamData,
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
