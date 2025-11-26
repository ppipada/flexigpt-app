package inference

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	completionResp *spec.FetchCompletionResponseBody,
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

	// Handle messages.
	messages := slices.Clone(prevMessages)
	for idx := range messages {
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

func toolFunctionName(ct spec.FetchCompletionToolChoice) string {
	slug := sanitizeToolNameComponent(ct.ToolSlug)
	version := sanitizeToolNameComponent(strings.ReplaceAll(ct.ToolVersion, ".", "_"))

	idPart := sanitizeToolNameComponent(strings.ReplaceAll(ct.ID, "-", ""))
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

// imageEncodingFromRef returns the base64-encoded contents of the image along
// with a best-effort MIME type derived from the ref's Format field.
func imageEncodingFromRef(
	ref *spec.ChatCompletionImageRef,
) (encoded, mimeType string, err error) {
	if ref == nil {
		return "", "", errors.New("nil image ref")
	}
	encoded, err = encodeFileToBase64(ref.Path)
	if err != nil {
		return "", "", err
	}

	format := strings.ToLower(strings.TrimSpace(ref.Format))
	switch format {
	case "jpg", "jpeg":
		mimeType = "image/jpeg"
	case "png":
		mimeType = "image/png"
	case "gif":
		mimeType = "image/gif"
	case "webp":
		mimeType = "image/webp"
	case "bmp":
		mimeType = "image/bmp"
	case "":
		// Default to PNG if we don't know the format.
		mimeType = "image/png"
	default:
		// Fallback for uncommon formats; most renderers can still display these.
		mimeType = "application/octet-stream"
	}
	return encoded, mimeType, nil
}

// fileEncodingFromRef returns the base64-encoded contents of the file along
// with a best-effort filename derived from the path.
func fileEncodingFromRef(
	ref *spec.ChatCompletionFileRef,
) (encoded, filename string, err error) {
	if ref == nil {
		return "", "", errors.New("nil file ref")
	}
	encoded, err = encodeFileToBase64(ref.Path)
	if err != nil {
		return "", "", err
	}
	filename = strings.TrimSpace(filepath.Base(ref.Path))
	if filename == "" {
		filename = "attachment"
	}
	return encoded, filename, nil
}

// encodeFileToBase64 reads the file at the given path and returns its
// base64-encoded contents. The caller is responsible for ensuring the path
// points to a regular file.
func encodeFileToBase64(path string) (string, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		return "", errors.New("empty attachment path")
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// formatAttachmentAsText normalizes an attachment into a short, human-readable
// string that can be injected into text prompts when a richer modality is not
// available (for example, for generic refs or providers without native file
// support).
func formatAttachmentAsText(att spec.ChatCompletionAttachment) string {
	label := strings.TrimSpace(att.Label)
	var detail string

	switch att.Kind {
	case spec.AttachmentFile:
		if att.FileRef != nil {
			detail = strings.TrimSpace(att.FileRef.Path)
		}
	case spec.AttachmentImage:
		if att.ImageRef != nil {
			detail = strings.TrimSpace(att.ImageRef.Path)
		}
	case spec.AttachmentDocIndex, spec.AttachmentPR, spec.AttachmentCommit, spec.AttachmentSnapshot:
		if att.GenericRef != nil {
			detail = strings.TrimSpace(att.GenericRef.Handle)
		}
	default:
		detail = ""
	}

	if label == "" {
		label = detail
	}

	if label == "" && detail == "" {
		return ""
	}
	if detail == "" || label == detail {
		return fmt.Sprintf("[Attachment: %s]", label)
	}
	return fmt.Sprintf("[Attachment: %s — %s]", label, detail)
}
