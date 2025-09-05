package main

import (
	"context"
	"errors"
	"sync"

	"github.com/ppipada/flexigpt-app/pkg/inference"
	inferenceSpec "github.com/ppipada/flexigpt-app/pkg/inference/spec"
	"github.com/ppipada/flexigpt-app/pkg/middleware"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ProviderSetWrapper struct {
	providersetAPI      *inference.ProviderSetAPI
	appContext          context.Context
	completionCancelMux sync.Mutex
	completionCancels   map[string]context.CancelFunc
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
	ps.completionCancels = map[string]context.CancelFunc{}
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

func (w *ProviderSetWrapper) BuildCompletionData(
	req *inferenceSpec.BuildCompletionDataRequest,
) (*inferenceSpec.BuildCompletionDataResponse, error) {
	return middleware.WithRecoveryResp(func() (*inferenceSpec.BuildCompletionDataResponse, error) {
		return w.providersetAPI.BuildCompletionData(context.Background(), req)
	})
}

// FetchCompletion handles the completion request and streams data back to the frontend.
func (w *ProviderSetWrapper) FetchCompletion(
	provider string,
	completionData *inferenceSpec.CompletionData,
	textCallbackID string,
	thinkingCallbackID string,
	requestID string,
) (*inferenceSpec.FetchCompletionResponse, error) {
	return middleware.WithRecoveryResp(func() (*inferenceSpec.FetchCompletionResponse, error) {
		if requestID == "" {
			return nil, errors.New("requestID is empty")
		}

		ctx, cancel := context.WithCancel(w.appContext)

		w.completionCancelMux.Lock()
		w.completionCancels[requestID] = cancel
		w.completionCancelMux.Unlock()
		defer func() {
			w.completionCancelMux.Lock()
			delete(w.completionCancels, requestID)
			w.completionCancelMux.Unlock()
			if textCallbackID != "" {
				runtime.EventsOff(w.appContext, textCallbackID)
			}
			if thinkingCallbackID != "" {
				runtime.EventsOff(w.appContext, thinkingCallbackID)
			}
		}()

		reqBody := &inferenceSpec.FetchCompletionRequestBody{
			CompletionData: completionData,
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
			Provider: modelpresetSpec.ProviderName(provider),
			Body:     reqBody,
		}
		resp, err := w.providersetAPI.FetchCompletion(
			ctx,
			req,
		)
		if err != nil {
			return nil, err
		}

		return resp, nil
	})
}

func (w *ProviderSetWrapper) CancelCompletion(id string) {
	w.completionCancelMux.Lock()
	defer w.completionCancelMux.Unlock()
	if c, ok := w.completionCancels[id]; ok {
		c()
		delete(w.completionCancels, id)
	}
}
