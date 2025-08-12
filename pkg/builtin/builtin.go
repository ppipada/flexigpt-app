package builtin

import (
	"embed"
)

//go:embed modelpresets
var BuiltInModelPresetsFS embed.FS

//go:embed tools
var BuiltInToolBundlesFS embed.FS

//go:embed prompts
var BuiltInPromptBundlesFS embed.FS

const (
	BuiltInModelPresetsRootDir = "modelpresets"
	BuiltInModelPresetsJSON    = "modelpresets.json"

	BuiltInToolBundlesRootDir = "tools"
	BuiltInToolBundlesJSON    = "tools.bundles.json"

	BuiltInPromptBundlesRootDir = "prompts"
	BuiltInPromptBundlesJSON    = "prompts.bundles.json"
)

const (
	ProviderNameAnthropic    = "anthropic"
	ProviderNameDeepseek     = "deepseek"
	ProviderNameGoogleGemini = "googlegemini"
	ProviderNameHuggingFace  = "huggingface"
	ProviderNameLlamaCPP     = "llamacpp"
	ProviderNameOpenAI       = "openai"
	ProviderNameOpenRouter   = "openrouter"
	ProviderNameXAI          = "xai"
)

var BuiltInProviderNames = []string{
	ProviderNameAnthropic,
	ProviderNameDeepseek,
	ProviderNameGoogleGemini,
	ProviderNameHuggingFace,
	ProviderNameLlamaCPP,
	ProviderNameOpenAI,
	ProviderNameOpenRouter,
	ProviderNameXAI,
}
