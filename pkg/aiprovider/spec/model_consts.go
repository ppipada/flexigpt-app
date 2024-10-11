package spec

var AnthropicModels = map[ModelName]ModelInfo{
	CLAUDE_3_5_SONNET: {
		Name:               CLAUDE_3_5_SONNET,
		DisplayName:        "Claude 3.5 Sonnet",
		Provider:           ProviderNameAnthropic,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
	CLAUDE_3_OPUS: {
		Name:               CLAUDE_3_OPUS,
		DisplayName:        "Claude 3 Opus",
		Provider:           ProviderNameAnthropic,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
	CLAUDE_3_SONNET: {
		Name:               CLAUDE_3_SONNET,
		DisplayName:        "Claude 3 Sonnet",
		Provider:           ProviderNameAnthropic,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
	CLAUDE_3_HAIKU: {
		Name:               CLAUDE_3_HAIKU,
		DisplayName:        "Claude 3 Haiku",
		Provider:           ProviderNameAnthropic,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
}

var GoogleModels = map[ModelName]ModelInfo{
	GEMINI_1_5_FLASH: {
		Name:               GEMINI_1_5_FLASH,
		DisplayName:        "Google Gemini 1.5 Flash",
		Provider:           ProviderNameGoogle,
		MaxPromptLength:    4096,
		MaxOutputLength:    8192,
		DefaultTemperature: 0.1,
	},
	GEMINI_1_5_PRO: {
		Name:               GEMINI_1_5_PRO,
		DisplayName:        "Google Gemini 1.5 Pro",
		Provider:           ProviderNameGoogle,
		MaxPromptLength:    4096,
		MaxOutputLength:    8192,
		DefaultTemperature: 0.1,
	},
}

var HuggingfaceModels = map[ModelName]ModelInfo{
	DEEPSEEK_CODER_1_3B_INSTRUCT: {
		Name:               DEEPSEEK_CODER_1_3B_INSTRUCT,
		DisplayName:        "HF Deepseek Coder 1.3b",
		Provider:           ProviderNameHuggingFace,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
}

var LlamacppModels = map[ModelName]ModelInfo{
	LLAMA_3: {
		Name:               LLAMA_3,
		DisplayName:        "Llama 3",
		Provider:           ProviderNameLlamaCPP,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
	LLAMA_3_1: {
		Name:               LLAMA_3_1,
		DisplayName:        "Llama 3.1",
		Provider:           ProviderNameLlamaCPP,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
}

var OpenaiModels = map[ModelName]ModelInfo{
	GPT_O1_PREVIEW: {
		Name:               GPT_O1_PREVIEW,
		DisplayName:        "OpenAI GPT o1 preview",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 1,
	},
	GPT_O1_MINI: {
		Name:               GPT_O1_MINI,
		DisplayName:        "OpenAI GPT o1 mini",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 1,
	},
	GPT_4O: {
		Name:               GPT_4O,
		DisplayName:        "OpenAI GPT 4o",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
	GPT_4: {
		Name:               GPT_4,
		DisplayName:        "OpenAI GPT 4",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
	GPT_3_5_TURBO: {
		Name:               GPT_3_5_TURBO,
		DisplayName:        "OpenAI GPT 3.5 turbo",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    2400,
		MaxOutputLength:    2400,
		DefaultTemperature: 0.1,
	},
	GPT_4O_MINI: {
		Name:               GPT_4O_MINI,
		DisplayName:        "OpenAI GPT 4o mini",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
}

var AllModelInfo = map[ModelName]ModelInfo{
	CLAUDE_3_5_SONNET:            AnthropicModels[CLAUDE_3_5_SONNET],
	CLAUDE_3_OPUS:                AnthropicModels[CLAUDE_3_OPUS],
	CLAUDE_3_SONNET:              AnthropicModels[CLAUDE_3_SONNET],
	CLAUDE_3_HAIKU:               AnthropicModels[CLAUDE_3_HAIKU],
	GEMINI_1_5_FLASH:             GoogleModels[GEMINI_1_5_FLASH],
	GEMINI_1_5_PRO:               GoogleModels[GEMINI_1_5_PRO],
	DEEPSEEK_CODER_1_3B_INSTRUCT: HuggingfaceModels[DEEPSEEK_CODER_1_3B_INSTRUCT],
	LLAMA_3:                      LlamacppModels[LLAMA_3],
	LLAMA_3_1:                    LlamacppModels[LLAMA_3_1],
	GPT_O1_PREVIEW:               OpenaiModels[GPT_O1_PREVIEW],
	GPT_O1_MINI:                  OpenaiModels[GPT_O1_MINI],
	GPT_4O:                       OpenaiModels[GPT_4O],
	GPT_4:                        OpenaiModels[GPT_4],
	GPT_3_5_TURBO:                OpenaiModels[GPT_3_5_TURBO],
	GPT_4O_MINI:                  OpenaiModels[GPT_4O_MINI],
}
