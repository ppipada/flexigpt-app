package api

import (
	"github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"
	"github.com/tmc/langchaingo/llms"
)

var LangchainRoleMap = map[spec.ChatCompletionRoleEnum]llms.ChatMessageType{
	// No developer prompt support in langchain as of now.
	spec.Developer: llms.ChatMessageTypeSystem,
	spec.System:    llms.ChatMessageTypeSystem,
	spec.User:      llms.ChatMessageTypeHuman,
	spec.Assistant: llms.ChatMessageTypeAI,
	spec.Function:  llms.ChatMessageTypeTool,
}
