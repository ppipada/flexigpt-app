package main

import (
	"context"
	"errors"

	"github.com/ppipada/flexigpt-app/pkg/inference"
	inferenceSpec "github.com/ppipada/flexigpt-app/pkg/inference/spec"
	"github.com/ppipada/flexigpt-app/pkg/middleware"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ProviderSetWrapper struct {
	providersetAPI *inference.ProviderSetAPI
	appContext     context.Context
}

// NewProviderSetWrapper creates a new ProviderSet with the specified default provider.
func InitProviderSetWrapper(
	ps *ProviderSetWrapper,
	defaultInbuiltProvider modelpresetSpec.ProviderName,
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
	req *inferenceSpec.SetDefaultProviderRequest,
) (*inferenceSpec.SetDefaultProviderResponse, error) {
	return middleware.WithRecoveryResp(func() (*inferenceSpec.SetDefaultProviderResponse, error) {
		return w.providersetAPI.SetDefaultProvider(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) GetConfigurationInfo(
	req *inferenceSpec.GetConfigurationInfoRequest,
) (*inferenceSpec.GetConfigurationInfoResponse, error) {
	return middleware.WithRecoveryResp(
		func() (*inferenceSpec.GetConfigurationInfoResponse, error) {
			return w.providersetAPI.GetConfigurationInfo(context.Background(), req)
		},
	)
}

func (w *ProviderSetWrapper) AddProvider(
	req *inferenceSpec.AddProviderRequest,
) (*inferenceSpec.AddProviderResponse, error) {
	return middleware.WithRecoveryResp(func() (*inferenceSpec.AddProviderResponse, error) {
		return w.providersetAPI.AddProvider(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) DeleteProvider(
	req *inferenceSpec.DeleteProviderRequest,
) (*inferenceSpec.DeleteProviderResponse, error) {
	return middleware.WithRecoveryResp(func() (*inferenceSpec.DeleteProviderResponse, error) {
		return w.providersetAPI.DeleteProvider(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) SetProviderAPIKey(
	req *inferenceSpec.SetProviderAPIKeyRequest,
) (*inferenceSpec.SetProviderAPIKeyResponse, error) {
	return middleware.WithRecoveryResp(func() (*inferenceSpec.SetProviderAPIKeyResponse, error) {
		return w.providersetAPI.SetProviderAPIKey(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) SetProviderAttribute(
	req *inferenceSpec.SetProviderAttributeRequest,
) (*inferenceSpec.SetProviderAttributeResponse, error) {
	return middleware.WithRecoveryResp(
		func() (*inferenceSpec.SetProviderAttributeResponse, error) {
			return w.providersetAPI.SetProviderAttribute(context.Background(), req)
		},
	)
}

// FetchCompletion handles the completion request and streams data back to the frontend.
func (w *ProviderSetWrapper) FetchCompletion(
	provider string,
	prompt string,
	modelParams inferenceSpec.ModelParams,
	prevMessages []inferenceSpec.ChatCompletionRequestMessage,
	callbackID string,
) (*inferenceSpec.FetchCompletionResponse, error) {
	return middleware.WithRecoveryResp(func() (*inferenceSpec.FetchCompletionResponse, error) {
		onStreamData := func(data string) error {
			runtime.EventsEmit(w.appContext, callbackID, data)
			return nil
		}

		req := &inferenceSpec.FetchCompletionRequest{
			Body: &inferenceSpec.FetchCompletionRequestBody{
				Provider:     modelpresetSpec.ProviderName(provider),
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
