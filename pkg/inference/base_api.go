package inference

import (
	"context"
	"slices"
	"strings"

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
		prompt string,
		modelParams spec.ModelParams,
		prevMessages []spec.ChatCompletionDataMessage,
	) (*spec.CompletionData, error)
	FetchCompletion(
		ctx context.Context,
		prompt string,
		modelParams spec.ModelParams,
		prevMessages []spec.ChatCompletionDataMessage,
		OnStreamTextData func(textData string) error,
		OnStreamThinkingData func(thinkingData string) error,
	) (*spec.CompletionResponse, error)
}

// attachDebugResp adds HTTP-debug information and error context—without panics.
//
// – ctx may or may not contain debug information.
// – respErr is the transport/SDK error (may be nil).
// – isNilResp tells whether the model returned an empty/invalid response.
func attachDebugResp(
	ctx context.Context,
	completionResp *spec.CompletionResponse,
	respErr error,
	isNilResp bool,
) {
	if completionResp == nil {
		return
	}

	debugResp, _ := GetDebugHTTPResponse(ctx)

	// Always attach request/response debug info if available.
	if debugResp != nil {
		completionResp.RequestDetails = debugResp.RequestDetails
		completionResp.ResponseDetails = debugResp.ResponseDetails
	}

	// Gather error-message fragments.
	var msgParts []string
	if debugResp != nil && debugResp.ErrorDetails != nil {
		if m := strings.TrimSpace(debugResp.ErrorDetails.Message); m != "" {
			msgParts = append(msgParts, m)
		}
	}
	if respErr != nil {
		msgParts = append(msgParts, respErr.Error())
	}
	if isNilResp {
		msgParts = append(msgParts, "got nil response from LLM api")
	}

	if len(msgParts) == 0 {
		// Nothing to write; leave ErrorDetails as-is (nil or previously set).
		return
	}

	// Prepare ErrorDetails without aliasing the debug struct pointer.
	if debugResp != nil && debugResp.ErrorDetails != nil {
		ed := *debugResp.ErrorDetails
		ed.Message = strings.Join(msgParts, "; ")
		completionResp.ErrorDetails = &ed
	} else {
		completionResp.ErrorDetails = &spec.APIErrorDetails{
			Message: strings.Join(msgParts, "; "),
		}
	}
}

func getCompletionData(
	prompt string,
	modelParams spec.ModelParams,
	prevMessages []spec.ChatCompletionDataMessage,
) *spec.CompletionData {
	completionData := spec.CompletionData{
		ModelParams: spec.ModelParams{
			Name:                        modelParams.Name,
			Stream:                      modelParams.Stream,
			MaxPromptLength:             modelParams.MaxPromptLength,
			MaxOutputLength:             modelParams.MaxOutputLength,
			Temperature:                 modelParams.Temperature,
			Reasoning:                   modelParams.Reasoning,
			SystemPrompt:                modelParams.SystemPrompt,
			Timeout:                     modelParams.Timeout,
			AdditionalParametersRawJSON: modelParams.AdditionalParametersRawJSON,
		},
	}

	// Handle messages.
	messages := slices.Clone(prevMessages)
	if prompt != "" {
		message := spec.ChatCompletionDataMessage{
			Role:    "user",
			Content: &prompt,
		}
		messages = append(messages, message)
	}
	completionData.Messages = messages

	// Assuming filterMessagesByTokenCount is implemented elsewhere.
	completionData.Messages = FilterMessagesByTokenCount(
		completionData.Messages,
		completionData.ModelParams.MaxPromptLength,
	)

	return &completionData
}

func TrimInbuiltPrompts(systemPrompt, inbuiltPrompt string) string {
	// Split both prompts into lines.
	inbuiltLines := strings.Split(inbuiltPrompt, "\n")
	promptLines := strings.Split(systemPrompt, "\n")

	// Remove matching lines from the start.
	for len(inbuiltLines) > 0 && len(promptLines) > 0 && promptLines[0] == inbuiltLines[0] {
		promptLines = promptLines[1:]
		inbuiltLines = inbuiltLines[1:]
	}
	// Re-join the remaining lines.
	return inbuiltPrompt + "\n" + strings.TrimLeft(strings.Join(promptLines, "\n"), "\n")
}
