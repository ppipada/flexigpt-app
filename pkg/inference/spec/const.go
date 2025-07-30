package spec

import (
	modelpresetConsts "github.com/ppipada/flexigpt-app/pkg/modelpreset/consts"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

var AnthropicProviderInfo = ProviderParams{
	Name:   modelpresetConsts.ProviderNameAnthropic,
	APIKey: "",
	Origin: "https://api.anthropic.com",
	Type:   modelpresetSpec.InbuiltAnthropicCompatible,

	APIKeyHeaderKey: "x-api-key",
	DefaultHeaders: map[string]string{
		"content-type":      "application/json",
		"accept":            "application/json",
		"anthropic-version": "2023-06-01",
	},
	ChatCompletionPathPrefix: "/v1/messages",
}

var DeepseekProviderInfo = ProviderParams{
	Name:   modelpresetConsts.ProviderNameDeepseek,
	APIKey: "",
	Origin: "https://api.deepseek.com",
	Type:   modelpresetSpec.InbuiltOpenAICompatible,

	APIKeyHeaderKey:          modelpresetSpec.OpenAICompatibleAPIKeyHeaderKey,
	DefaultHeaders:           modelpresetSpec.OpenAICompatibleDefaultHeaders,
	ChatCompletionPathPrefix: modelpresetSpec.OpenAICompatibleChatCompletionPathPrefix,
}

var GoogleProviderInfo = ProviderParams{
	Name:   modelpresetConsts.ProviderNameGoogle,
	APIKey: "",
	Origin: "https://generativelanguage.googleapis.com",
	Type:   modelpresetSpec.InbuiltOpenAICompatible,

	APIKeyHeaderKey:          modelpresetSpec.OpenAICompatibleAPIKeyHeaderKey,
	DefaultHeaders:           modelpresetSpec.OpenAICompatibleDefaultHeaders,
	ChatCompletionPathPrefix: "/v1beta/openai/chat/completions",
}

var HuggingfaceProviderInfo = ProviderParams{
	Name:   modelpresetConsts.ProviderNameHuggingFace,
	APIKey: "",
	Origin: "https://api-inference.huggingface.co",
	Type:   modelpresetSpec.InbuiltHuggingFaceCompatible,

	APIKeyHeaderKey:          "Authorization",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/models",
}

var LlamacppProviderInfo = ProviderParams{
	Name:   modelpresetConsts.ProviderNameLlamaCPP,
	APIKey: "",
	Origin: "http://127.0.0.1:8080",
	Type:   modelpresetSpec.InbuiltOpenAICompatible,

	APIKeyHeaderKey:          modelpresetSpec.OpenAICompatibleAPIKeyHeaderKey,
	DefaultHeaders:           modelpresetSpec.OpenAICompatibleDefaultHeaders,
	ChatCompletionPathPrefix: modelpresetSpec.OpenAICompatibleChatCompletionPathPrefix,
}

var OpenAIProviderInfo = ProviderParams{
	Name:   modelpresetConsts.ProviderNameOpenAI,
	APIKey: "",
	Origin: "https://api.openai.com",
	Type:   modelpresetSpec.InbuiltOpenAICompatible,

	APIKeyHeaderKey:          modelpresetSpec.OpenAICompatibleAPIKeyHeaderKey,
	DefaultHeaders:           modelpresetSpec.OpenAICompatibleDefaultHeaders,
	ChatCompletionPathPrefix: modelpresetSpec.OpenAICompatibleChatCompletionPathPrefix,
}

var InbuiltProviders = map[modelpresetSpec.ProviderName]ProviderParams{
	modelpresetConsts.ProviderNameAnthropic:   AnthropicProviderInfo,
	modelpresetConsts.ProviderNameDeepseek:    DeepseekProviderInfo,
	modelpresetConsts.ProviderNameGoogle:      GoogleProviderInfo,
	modelpresetConsts.ProviderNameHuggingFace: HuggingfaceProviderInfo,
	modelpresetConsts.ProviderNameLlamaCPP:    LlamacppProviderInfo,
	modelpresetConsts.ProviderNameOpenAI:      OpenAIProviderInfo,
}

var InbuiltProviderModels = map[modelpresetSpec.ProviderName]modelpresetSpec.ProviderPreset{
	modelpresetConsts.ProviderNameAnthropic:   modelpresetConsts.AnthropicProviderPreset,
	modelpresetConsts.ProviderNameDeepseek:    modelpresetConsts.DeepseekProviderPreset,
	modelpresetConsts.ProviderNameGoogle:      modelpresetConsts.GoogleProviderPreset,
	modelpresetConsts.ProviderNameHuggingFace: modelpresetConsts.HuggingfaceProviderPreset,
	modelpresetConsts.ProviderNameLlamaCPP:    modelpresetConsts.LlamacppProviderPreset,
	modelpresetConsts.ProviderNameOpenAI:      modelpresetConsts.OpenAIProviderPreset,
}
