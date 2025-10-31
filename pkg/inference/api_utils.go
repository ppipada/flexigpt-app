package inference

import (
	"context"
	"slices"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
)

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
	modelParams spec.ModelParams,
	currentMessage spec.ChatCompletionDataMessage,
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
	for idx := range messages {
		// Omit any tools or name in previous messages.
		messages[idx].ToolAttachments = []spec.ChatCompletionToolAttachment{}
		messages[idx].Name = nil
	}

	messages = append(messages, currentMessage)
	completionData.Messages = messages

	// Assuming filterMessagesByTokenCount is implemented elsewhere.
	completionData.Messages = FilterMessagesByTokenCount(
		completionData.Messages,
		completionData.ModelParams.MaxPromptLength,
	)

	return &completionData
}
