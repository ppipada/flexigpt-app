package main

import (
	"context"
	"errors"

	"github.com/ppipada/flexigpt-app/pkg/aiprovider"
	aiproviderAPI "github.com/ppipada/flexigpt-app/pkg/aiprovider/api"
	aiproviderSpec "github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"
	"github.com/ppipada/flexigpt-app/pkg/middleware"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ProviderSetWrapper struct {
	providersetAPI *aiprovider.ProviderSetAPI
	appContext     context.Context
}

// NewProviderSetWrapper creates a new ProviderSet with the specified default provider.
func InitProviderSetWrapper(
	ps *ProviderSetWrapper,
	defaultInbuiltProvider aiproviderSpec.ProviderName,
) error {
	p, err := aiprovider.NewProviderSetAPI(defaultInbuiltProvider, false)
	if err != nil {
		return errors.Join(err, errors.New("invalid default provider"))
	}
	ps.providersetAPI = p
	return nil
}

func SetWrappedProviderAppContext(w *ProviderSetWrapper, ctx context.Context) {
	w.appContext = ctx
}

func (w *ProviderSetWrapper) SetDefaultProvider(
	req *aiproviderAPI.SetDefaultProviderRequest,
) (*aiproviderAPI.SetDefaultProviderResponse, error) {
	return middleware.WithRecoveryResp(func() (*aiproviderAPI.SetDefaultProviderResponse, error) {
		return w.providersetAPI.SetDefaultProvider(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) GetConfigurationInfo(
	req *aiproviderAPI.GetConfigurationInfoRequest,
) (*aiproviderAPI.GetConfigurationInfoResponse, error) {
	return middleware.WithRecoveryResp(
		func() (*aiproviderAPI.GetConfigurationInfoResponse, error) {
			return w.providersetAPI.GetConfigurationInfo(context.Background(), req)
		},
	)
}

func (w *ProviderSetWrapper) AddProvider(
	req *aiproviderAPI.AddProviderRequest,
) (*aiproviderAPI.AddProviderResponse, error) {
	return middleware.WithRecoveryResp(func() (*aiproviderAPI.AddProviderResponse, error) {
		return w.providersetAPI.AddProvider(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) DeleteProvider(
	req *aiproviderAPI.DeleteProviderRequest,
) (*aiproviderAPI.DeleteProviderResponse, error) {
	return middleware.WithRecoveryResp(func() (*aiproviderAPI.DeleteProviderResponse, error) {
		return w.providersetAPI.DeleteProvider(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) SetProviderAPIKey(
	req *aiproviderAPI.SetProviderAPIKeyRequest,
) (*aiproviderAPI.SetProviderAPIKeyResponse, error) {
	return middleware.WithRecoveryResp(func() (*aiproviderAPI.SetProviderAPIKeyResponse, error) {
		return w.providersetAPI.SetProviderAPIKey(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) SetProviderAttribute(
	req *aiproviderAPI.SetProviderAttributeRequest,
) (*aiproviderAPI.SetProviderAttributeResponse, error) {
	return middleware.WithRecoveryResp(
		func() (*aiproviderAPI.SetProviderAttributeResponse, error) {
			return w.providersetAPI.SetProviderAttribute(context.Background(), req)
		},
	)
}

// FetchCompletion handles the completion request and streams data back to the frontend.
func (w *ProviderSetWrapper) FetchCompletion(
	provider string,
	prompt string,
	modelParams aiproviderSpec.ModelParams,
	prevMessages []aiproviderSpec.ChatCompletionRequestMessage,
	callbackID string,
) (*aiproviderAPI.FetchCompletionResponse, error) {
	return middleware.WithRecoveryResp(func() (*aiproviderAPI.FetchCompletionResponse, error) {
		onStreamData := func(data string) error {
			runtime.EventsEmit(w.appContext, callbackID, data)
			return nil
		}

		req := &aiproviderAPI.FetchCompletionRequest{
			Body: &aiproviderAPI.FetchCompletionRequestBody{
				Provider:     aiproviderSpec.ProviderName(provider),
				Prompt:       prompt,
				ModelParams:  modelParams,
				PrevMessages: prevMessages,
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
	})
}
