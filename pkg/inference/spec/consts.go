package spec

import (
	modelConsts "github.com/ppipada/flexigpt-app/pkg/model/consts"
	modelSpec "github.com/ppipada/flexigpt-app/pkg/model/spec"
)

var AnthropicProviderInfo = ProviderParams{
	Name:   modelConsts.ProviderNameAnthropic,
	APIKey: "",
	Origin: "https://api.anthropic.com",
	Type:   modelSpec.InbuiltAnthropicCompatible,

	APIKeyHeaderKey: "x-api-key",
	DefaultHeaders: map[string]string{
		"content-type":      "application/json",
		"accept":            "application/json",
		"anthropic-version": "2023-06-01",
	},
	ChatCompletionPathPrefix: "/v1/messages",
}

var DeepseekProviderInfo = ProviderParams{
	Name:   modelConsts.ProviderNameDeepseek,
	APIKey: "",
	Origin: "https://api.deepseek.com",
	Type:   modelSpec.InbuiltOpenAICompatible,

	APIKeyHeaderKey:          modelConsts.OpenAICompatibleAPIKeyHeaderKey,
	DefaultHeaders:           modelConsts.OpenAICompatibleDefaultHeaders,
	ChatCompletionPathPrefix: modelConsts.OpenAICompatibleChatCompletionPathPrefix,
}

var GoogleProviderInfo = ProviderParams{
	Name:   modelConsts.ProviderNameGoogle,
	APIKey: "",
	Origin: "https://generativelanguage.googleapis.com",
	Type:   modelSpec.InbuiltOpenAICompatible,

	APIKeyHeaderKey:          modelConsts.OpenAICompatibleAPIKeyHeaderKey,
	DefaultHeaders:           modelConsts.OpenAICompatibleDefaultHeaders,
	ChatCompletionPathPrefix: "/v1beta/openai/chat/completions",
}

var HuggingfaceProviderInfo = ProviderParams{
	Name:   modelConsts.ProviderNameHuggingFace,
	APIKey: "",
	Origin: "https://api-inference.huggingface.co",
	Type:   modelSpec.InbuiltHuggingFaceCompatible,

	APIKeyHeaderKey:          "Authorization",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/models",
}

var LlamacppProviderInfo = ProviderParams{
	Name:   modelConsts.ProviderNameLlamaCPP,
	APIKey: "",
	Origin: "http://127.0.0.1:8080",
	Type:   modelSpec.InbuiltOpenAICompatible,

	APIKeyHeaderKey:          modelConsts.OpenAICompatibleAPIKeyHeaderKey,
	DefaultHeaders:           modelConsts.OpenAICompatibleDefaultHeaders,
	ChatCompletionPathPrefix: modelConsts.OpenAICompatibleChatCompletionPathPrefix,
}

var OpenAIProviderInfo = ProviderParams{
	Name:   modelConsts.ProviderNameOpenAI,
	APIKey: "",
	Origin: "https://api.openai.com",
	Type:   modelSpec.InbuiltOpenAICompatible,

	APIKeyHeaderKey:          modelConsts.OpenAICompatibleAPIKeyHeaderKey,
	DefaultHeaders:           modelConsts.OpenAICompatibleDefaultHeaders,
	ChatCompletionPathPrefix: modelConsts.OpenAICompatibleChatCompletionPathPrefix,
}

var InbuiltProviders = map[modelSpec.ProviderName]ProviderParams{
	modelConsts.ProviderNameAnthropic:   AnthropicProviderInfo,
	modelConsts.ProviderNameDeepseek:    DeepseekProviderInfo,
	modelConsts.ProviderNameGoogle:      GoogleProviderInfo,
	modelConsts.ProviderNameHuggingFace: HuggingfaceProviderInfo,
	modelConsts.ProviderNameLlamaCPP:    LlamacppProviderInfo,
	modelConsts.ProviderNameOpenAI:      OpenAIProviderInfo,
}
