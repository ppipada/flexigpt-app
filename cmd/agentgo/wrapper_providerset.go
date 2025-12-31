package main

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	inferencegoSpec "github.com/ppipada/inference-go/spec"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/ppipada/flexigpt-app/pkg/inferencewrapper"
	inferencewrapperSpec "github.com/ppipada/flexigpt-app/pkg/inferencewrapper/spec"
	"github.com/ppipada/flexigpt-app/pkg/middleware"
	toolStore "github.com/ppipada/flexigpt-app/pkg/tool/store"
)

type ProviderSetWrapper struct {
	providersetAPI      *inferencewrapper.ProviderSetAPI
	appContext          context.Context
	completionCancelMux sync.Mutex
	completionCancels   map[string]context.CancelFunc
}

// InitProviderSetWrapper creates a new ProviderSet with the specified default provider.
func InitProviderSetWrapper(
	ps *ProviderSetWrapper,
	ts *toolStore.ToolStore,
) error {
	p, err := inferencewrapper.NewProviderSetAPI(slog.Default(), ts)
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
	req *inferencewrapperSpec.AddProviderRequest,
) (*inferencewrapperSpec.AddProviderResponse, error) {
	return middleware.WithRecoveryResp(func() (*inferencewrapperSpec.AddProviderResponse, error) {
		return w.providersetAPI.AddProvider(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) DeleteProvider(
	req *inferencewrapperSpec.DeleteProviderRequest,
) (*inferencewrapperSpec.DeleteProviderResponse, error) {
	return middleware.WithRecoveryResp(func() (*inferencewrapperSpec.DeleteProviderResponse, error) {
		return w.providersetAPI.DeleteProvider(context.Background(), req)
	})
}

func (w *ProviderSetWrapper) SetProviderAPIKey(
	req *inferencewrapperSpec.SetProviderAPIKeyRequest,
) (*inferencewrapperSpec.SetProviderAPIKeyResponse, error) {
	return middleware.WithRecoveryResp(func() (*inferencewrapperSpec.SetProviderAPIKeyResponse, error) {
		return w.providersetAPI.SetProviderAPIKey(context.Background(), req)
	})
}

// FetchCompletion handles the completion request and streams data back to the frontend.
func (w *ProviderSetWrapper) FetchCompletion(
	provider string,
	completionData *inferencewrapperSpec.CompletionRequestBody,
	textCallbackID string,
	thinkingCallbackID string,
	requestID string,
) (*inferencewrapperSpec.CompletionResponse, error) {
	return middleware.WithRecoveryResp(func() (*inferencewrapperSpec.CompletionResponse, error) {
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

		req := &inferencewrapperSpec.CompletionRequest{
			Provider: inferencegoSpec.ProviderName(provider),
			Body:     completionData,
		}

		if textCallbackID != "" && thinkingCallbackID != "" {
			req.OnStreamText = func(textData string) error {
				runtime.EventsEmit(w.appContext, textCallbackID, textData)
				return nil
			}
			req.OnStreamThinking = func(thinkingData string) error {
				runtime.EventsEmit(w.appContext, thinkingCallbackID, thinkingData)
				return nil
			}
		}
		resp, err := w.providersetAPI.FetchCompletion(
			ctx,
			req,
		)
		if err != nil {
			// If we have a partial response, attach error info there and return it.
			if resp != nil && resp.Body != nil && resp.Body.InferenceResponse != nil {
				if resp.Body.InferenceResponse.Error == nil {
					resp.Body.InferenceResponse.Error = &inferencegoSpec.Error{
						Message: err.Error(),
					}
				}
				// Log, but do not propagate Go error so Wails resolves the Promise.
				slog.Error("fetchCompletion failed", "provider", provider, "err", err)
				return resp, nil
			}
			// No response at all => infrastructure error.
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
