package main

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/flexigpt/inference-go/debugclient"
	inferencegoSpec "github.com/flexigpt/inference-go/spec"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/flexigpt/flexigpt-app/internal/inferencewrapper"
	inferencewrapperSpec "github.com/flexigpt/flexigpt-app/internal/inferencewrapper/spec"
	"github.com/flexigpt/flexigpt-app/internal/middleware"
	toolStore "github.com/flexigpt/flexigpt-app/internal/tool/store"
)

type ProviderSetWrapper struct {
	providersetAPI      *inferencewrapper.ProviderSetAPI
	appContext          context.Context
	completionCancelMux sync.Mutex
	completionCancels   map[string]context.CancelFunc
	preCanceled         map[string]time.Time
}

// InitProviderSetWrapper creates a new ProviderSet with the specified default provider.
func InitProviderSetWrapper(
	ps *ProviderSetWrapper,
	ts *toolStore.ToolStore,
) error {
	p, err := inferencewrapper.NewProviderSetAPI(
		ts,
		inferencewrapper.WithLogger(slog.Default()),
		inferencewrapper.WithDebugConfig(&debugclient.DebugConfig{
			Disable:                 false,
			DisableRequestBody:      false,
			DisableResponseBody:     false,
			DisableContentStripping: false,
			LogToSlog:               false,
		}),
	)
	if err != nil {
		return errors.Join(err, errors.New("invalid default provider"))
	}
	ps.providersetAPI = p
	ps.completionCancels = map[string]context.CancelFunc{}
	ps.preCanceled = map[string]time.Time{}

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
		if w.appContext == nil {
			return nil, errors.New("appContext is not set (call SetWrappedProviderAppContext during startup)")
		}
		if w.completionCancels == nil {
			w.completionCancels = map[string]context.CancelFunc{}
		}
		if w.preCanceled == nil {
			w.preCanceled = map[string]time.Time{}
		}

		ctx, cancel := context.WithCancel(w.appContext)
		defer cancel()

		w.completionCancelMux.Lock()
		// If a cancel arrived before the fetch registered, honor it.
		if _, ok := w.preCanceled[requestID]; ok {
			delete(w.preCanceled, requestID)
			w.completionCancelMux.Unlock()
			return nil, context.Canceled
		}
		// Protect against requestID reuse while in-flight.
		if _, exists := w.completionCancels[requestID]; exists {
			w.completionCancelMux.Unlock()
			return nil, errors.New("duplicate requestID: a completion with this id is already in flight")
		}

		w.completionCancels[requestID] = cancel
		w.completionCancelMux.Unlock()

		defer func() {
			w.completionCancelMux.Lock()
			delete(w.completionCancels, requestID)
			w.completionCancelMux.Unlock()
		}()

		req := &inferencewrapperSpec.CompletionRequest{
			Provider: inferencegoSpec.ProviderName(provider),
			Body:     completionData,
		}

		if textCallbackID != "" {
			req.OnStreamText = func(textData string) error {
				if err := ctx.Err(); err != nil {
					return err
				}
				//nolint:contextcheck // Need to pass app context here and not new context.
				runtime.EventsEmit(w.appContext, textCallbackID, textData)
				return nil
			}
		}
		if thinkingCallbackID != "" {
			req.OnStreamThinking = func(thinkingData string) error {
				if err := ctx.Err(); err != nil {
					return err
				}
				//nolint:contextcheck // Need to pass app context here and not new context.
				runtime.EventsEmit(w.appContext, thinkingCallbackID, thinkingData)
				return nil
			}
		}
		resp, err := w.providersetAPI.FetchCompletion(
			ctx,
			req,
		)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				// Expected lifecycle event; return partial resp if present without noisy error logging.
				if resp != nil {
					return resp, nil
				}
				return nil, err
			}
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
	if id == "" {
		return
	}
	w.completionCancelMux.Lock()
	defer w.completionCancelMux.Unlock()
	if c, ok := w.completionCancels[id]; ok {
		c()
		delete(w.completionCancels, id)
		return
	}

	// Cancel arrived before FetchCompletion registered the cancel func.
	w.preCanceled[id] = time.Now()

	// Best-effort pruning to avoid unbounded growth.
	cutoff := time.Now().Add(-2 * time.Minute)
	for k, t := range w.preCanceled {
		if t.Before(cutoff) {
			delete(w.preCanceled, k)
		}
	}
}
