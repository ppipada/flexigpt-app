package openaicompat

const (
	APIKeyHeaderKey          = "Authorization"
	ChatCompletionPathPrefix = "/v1/chat/completions"
)

var DefaultHeaders = map[string]string{"content-type": "application/json"}
