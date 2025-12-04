package inference

import (
	"context"
	"encoding/json"
	"log/slog"
	"runtime/debug"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/attachment"
	"github.com/ppipada/flexigpt-app/pkg/inference/debugclient"
	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	"github.com/ppipada/mapstore-go/jsonencdec"
)

// attachDebugResp adds HTTP-debug information and error context—without panics.
//
// - ctx may or may not contain debug information.
// - respErr is the transport/SDK error (may be nil).
// - isNilResp tells whether the model returned an empty/invalid response.
// - rawModelJSON is an optional, provider-level JSON representation of the *final* model response (e.g. OpenAI
// responses `resp.RawJSON()` or `json.Marshal(fullResponse)` for other SDKs). If provided and the HTTP debug layer
// did not already set ResponseDetails.Data, we will sanitize and store this JSON there.
func attachDebugResp(
	ctx context.Context,
	completionResp *spec.FetchCompletionResponseBody,
	respErr error,
	isNilResp bool,
	rawModelJSON string,
	fullObj any,
) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("attach debug resp panic",
				"recover", r,
				"stack", string(debug.Stack()))
		}
	}()

	if completionResp == nil {
		return
	}

	debugResp, _ := debugclient.GetDebugHTTPResponse(ctx)

	// Always attach request/response debug info from the HTTP layer if available.
	if debugResp != nil {
		if debugResp.RequestDetails != nil && completionResp.RequestDetails == nil {
			completionResp.RequestDetails = debugResp.RequestDetails
		}
		if debugResp.ResponseDetails != nil && completionResp.ResponseDetails == nil {
			completionResp.ResponseDetails = debugResp.ResponseDetails
		}
	}

	// If the HTTP layer didn't populate ResponseDetails.Data (most common in
	// streaming/SSE cases), and we have a provider-level raw JSON for the final
	// model response, sanitize that and use it as the debug body.

	if rawModelJSON != "" {
		if completionResp.ResponseDetails == nil {
			completionResp.ResponseDetails = &spec.APIResponseDetails{}
		}

		replace := false
		switch v := completionResp.ResponseDetails.Data.(type) {
		case nil:
			replace = true
		case string:
			// Likely raw stream (SSE / JSONL / etc), not structured JSON.
			replace = true
			_ = v // We could also inspect v for "data:" prefixes, etc.
		}

		if replace {
			completionResp.ResponseDetails.Data = debugclient.SanitizeJSONForDebug([]byte(rawModelJSON), true)
		}
	} else if fullObj != nil {
		if completionResp.ResponseDetails == nil {
			completionResp.ResponseDetails = &spec.APIResponseDetails{}
		}
		// We got a object. Lets replace always.
		m, err := jsonencdec.StructWithJSONTagsToMap(fullObj)
		if err == nil {
			completionResp.ResponseDetails.Data = debugclient.ScrubAnyForDebug(m, true)
		}

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
) *spec.FetchCompletionData {
	completionData := spec.FetchCompletionData{
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

	msgs := make([]spec.ChatCompletionDataMessage, 0, len(prevMessages))
	for _, m := range prevMessages {
		msgs = append(msgs, convertBuildMessageToChatMessage(m))
	}
	msgs = append(msgs, convertBuildMessageToChatMessage(currentMessage))

	completionData.Messages = msgs

	// Assuming filterMessagesByTokenCount is implemented elsewhere.
	completionData.Messages = FilterMessagesByTokenCount(
		completionData.Messages,
		completionData.ModelParams.MaxPromptLength,
	)

	return &completionData
}

func convertBuildMessageToChatMessage(
	msg spec.ChatCompletionDataMessage,
) spec.ChatCompletionDataMessage {
	// Keep the role, erase the name.
	out := spec.ChatCompletionDataMessage{
		Role: msg.Role,
		Name: nil,
	}
	if msg.Content != nil {
		c := *msg.Content
		out.Content = &c
	}

	// Copy attachments.
	if len(msg.Attachments) > 0 {
		out.Attachments = make([]attachment.Attachment, len(msg.Attachments))
		copy(out.Attachments, msg.Attachments)
	}
	return out
}

func toolFunctionName(ct spec.FetchCompletionToolChoice) string {
	slug := sanitizeToolNameComponent(string(ct.ToolSlug))
	version := sanitizeToolNameComponent(strings.ReplaceAll(ct.ToolVersion, ".", "_"))

	idPart := sanitizeToolNameComponent(strings.ReplaceAll(string(ct.ToolID), "-", ""))
	if len(idPart) > 8 {
		idPart = idPart[:8]
	}

	parts := make([]string, 0, 3)
	if slug != "" {
		parts = append(parts, slug)
	}
	if version != "" {
		parts = append(parts, version)
	}
	if idPart != "" {
		parts = append(parts, idPart)
	}
	if len(parts) == 0 {
		parts = append(parts, "tool")
	}

	name := strings.Join(parts, "_")
	if len(name) > 64 {
		name = name[:64]
	}
	name = strings.Trim(name, "_-")
	if name == "" {
		return "tool"
	}
	return name
}

func sanitizeToolNameComponent(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "_-")
	return out
}

func decodeToolArgSchema(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{"type": "object"}, nil
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func toolDescription(ct spec.FetchCompletionToolChoice) string {
	if desc := strings.TrimSpace(ct.Description); desc != "" {
		return desc
	}
	return ""
}
