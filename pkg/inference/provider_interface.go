package inference

import (
	"context"

	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
)

type CompletionProvider interface {
	InitLLM(ctx context.Context) error
	DeInitLLM(ctx context.Context) error
	GetProviderInfo(
		ctx context.Context,
	) *spec.ProviderParams
	IsConfigured(ctx context.Context) bool
	SetProviderAPIKey(
		ctx context.Context,
		apiKey string,
	) error
	BuildCompletionData(
		ctx context.Context,
		modelParams spec.ModelParams,
		currentMessage spec.ChatCompletionDataMessage,
		prevMessages []spec.ChatCompletionDataMessage,
	) (*spec.CompletionData, error)
	FetchCompletion(
		ctx context.Context,
		completionData *spec.CompletionData,
		OnStreamTextData func(textData string) error,
		OnStreamThinkingData func(thinkingData string) error,
	) (*spec.CompletionResponse, error)
}
