package inference

import (
	"context"

	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	inferencegoSpec "github.com/ppipada/inference-go/spec"
)

type CompletionProvider interface {
	InitLLM(ctx context.Context) error
	DeInitLLM(ctx context.Context) error
	GetProviderInfo(
		ctx context.Context,
	) *inferencegoSpec.ProviderParam
	IsConfigured(ctx context.Context) bool
	SetProviderAPIKey(
		ctx context.Context,
		apiKey string,
	) error
	FetchCompletion(
		ctx context.Context,
		completionData *spec.FetchCompletionData,
		OnStreamTextData func(textData string) error,
		OnStreamThinkingData func(thinkingData string) error,
	) (*spec.FetchCompletionResponse, error)
}
