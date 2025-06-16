package main

import (
	"context"
	"errors"

	"github.com/ppipada/flexigpt-app/pkg/inference"
	"github.com/ppipada/flexigpt-app/pkg/middleware"
	modelSpec "github.com/ppipada/flexigpt-app/pkg/model/spec"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ProviderSetWrapper struct {
	providersetAPI *inference.ProviderSetAPI
	appContext     context.Context
}

// NewProviderSetWrapper creates a new ProviderSet with the specified default provider.
func InitProviderSetWrapper(
	ps *ProviderSetWrapper,
	defaultInbuiltProvider modelSpec.ProviderName,
) error {
	p, err := inference.NewProviderSetAPI(defaultInbuiltProvider, false)
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
	req *inference.SetDefaultProviderRequest,
) (*inference.SetDefaultProviderResponse, error) {
	return middleware.WithRecoveryResp(func() (*inference.SetDefaultProviderResponse, error) {
		return w.providersetAPI.SetDefaultProvider(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) GetConfigurationInfo(
	req *inference.GetConfigurationInfoRequest,
) (*inference.GetConfigurationInfoResponse, error) {
	return middleware.WithRecoveryResp(
		func() (*inference.GetConfigurationInfoResponse, error) {
			return w.providersetAPI.GetConfigurationInfo(context.Background(), req)
		},
	)
}

func (w *ProviderSetWrapper) AddProvider(
	req *inference.AddProviderRequest,
) (*inference.AddProviderResponse, error) {
	return middleware.WithRecoveryResp(func() (*inference.AddProviderResponse, error) {
		return w.providersetAPI.AddProvider(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) DeleteProvider(
	req *inference.DeleteProviderRequest,
) (*inference.DeleteProviderResponse, error) {
	return middleware.WithRecoveryResp(func() (*inference.DeleteProviderResponse, error) {
		return w.providersetAPI.DeleteProvider(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) SetProviderAPIKey(
	req *inference.SetProviderAPIKeyRequest,
) (*inference.SetProviderAPIKeyResponse, error) {
	return middleware.WithRecoveryResp(func() (*inference.SetProviderAPIKeyResponse, error) {
		return w.providersetAPI.SetProviderAPIKey(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) SetProviderAttribute(
	req *inference.SetProviderAttributeRequest,
) (*inference.SetProviderAttributeResponse, error) {
	return middleware.WithRecoveryResp(
		func() (*inference.SetProviderAttributeResponse, error) {
			return w.providersetAPI.SetProviderAttribute(context.Background(), req)
		},
	)
}

// FetchCompletion handles the completion request and streams data back to the frontend.
func (w *ProviderSetWrapper) FetchCompletion(
	provider string,
	prompt string,
	modelParams modelSpec.ModelParams,
	prevMessages []inference.ChatCompletionRequestMessage,
	callbackID string,
) (*inference.FetchCompletionResponse, error) {
	return middleware.WithRecoveryResp(func() (*inference.FetchCompletionResponse, error) {
		onStreamData := func(data string) error {
			runtime.EventsEmit(w.appContext, callbackID, data)
			return nil
		}

		req := &inference.FetchCompletionRequest{
			Body: &inference.FetchCompletionRequestBody{
				Provider:     modelSpec.ProviderName(provider),
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
