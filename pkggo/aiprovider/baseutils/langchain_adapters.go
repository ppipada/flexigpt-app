package baseutils

import (
	"github.com/flexigpt/flexiui/pkggo/aiprovider/spec"
	"github.com/tmc/langchaingo/llms"
)

var LangchainRoleMap = map[spec.ChatCompletionRoleEnum]llms.ChatMessageType{
	spec.System:    llms.ChatMessageTypeSystem,
	spec.User:      llms.ChatMessageTypeHuman,
	spec.Assistant: llms.ChatMessageTypeAI,
	spec.Function:  llms.ChatMessageTypeTool,
}
