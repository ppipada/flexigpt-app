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
) error {
	p, err := inference.NewProviderSetAPI(false)
	if err != nil {
		return errors.Join(err, errors.New("invalid default provider"))
	}
	ps.providersetAPI = p
	return nil
}

func SetWrappedProviderAppContext(w *ProviderSetWrapper, ctx context.Context) {
	w.appContext = ctx
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

// FetchCompletion handles the completion request and streams data back to the frontend.
func (w *ProviderSetWrapper) FetchCompletion(
	provider string,
	prompt string,
	modelParams inferenceSpec.ModelParams,
	prevMessages []inferenceSpec.ChatCompletionRequestMessage,
	textCallbackID string,
	thinkingCallbackID string,
) (*inferenceSpec.FetchCompletionResponse, error) {
	return middleware.WithRecoveryResp(func() (*inferenceSpec.FetchCompletionResponse, error) {
		reqBody := &inferenceSpec.FetchCompletionRequestBody{
			Provider:     modelpresetSpec.ProviderName(provider),
			Prompt:       prompt,
			ModelParams:  modelParams,
			PrevMessages: prevMessages,
		}
		if textCallbackID != "" && thinkingCallbackID != "" {
			reqBody.OnStreamTextData = func(textData string) error {
				runtime.EventsEmit(w.appContext, textCallbackID, textData)
				return nil
			}
			reqBody.OnStreamThinkingData = func(thinkingData string) error {
				runtime.EventsEmit(w.appContext, thinkingCallbackID, thinkingData)
				return nil
			}
		}

		req := &inferenceSpec.FetchCompletionRequest{
			Body: reqBody,
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
