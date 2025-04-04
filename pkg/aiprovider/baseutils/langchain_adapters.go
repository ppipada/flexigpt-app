package baseutils

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
	"github.com/tmc/langchaingo/llms"
)

var LangchainRoleMap = map[spec.ChatCompletionRoleEnum]llms.ChatMessageType{
	// No developer prompt support in langchain as of now
	spec.Developer: llms.ChatMessageTypeSystem,
	spec.System:    llms.ChatMessageTypeSystem,
	spec.User:      llms.ChatMessageTypeHuman,
	spec.Assistant: llms.ChatMessageTypeAI,
	spec.Function:  llms.ChatMessageTypeTool,
}
