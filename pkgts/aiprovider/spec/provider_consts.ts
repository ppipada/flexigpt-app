/* eslint-disable @typescript-eslint/no-unused-vars */
import { ModelInfo, ModelName, ProviderInfo, ProviderName } from './provider_types';

export const anthropicProviderInfo: ProviderInfo = {
	name: ProviderName.ANTHROPIC,
	apiKey: '',
	engine: '',
	defaultOrigin: 'https://api.anthropic.com/v1',
	defaultModel: ModelName.CLAUDE_3_HAIKU,
	additionalSettings: {},
	timeout: 120,
	apiKeyHeaderKey: 'x-api-key',
	defaultHeaders: {
		'content-type': 'application/json',
		accept: 'application/json',
		'anthropic-version': '2023-06-01',
	},
	chatCompletionPathPrefix: '/messages',
	defaultTemperature: 0.0,
	streamingSupport: true,
};

export const googleProviderInfo: ProviderInfo = {
	name: ProviderName.GOOGLE,
	apiKey: '',
	engine: '',
	defaultOrigin: 'https://generativelanguage.googleapis.com',
	defaultModel: ModelName.GEMINI_1_5_FLASH,
	additionalSettings: {},
	timeout: 120,
	apiKeyHeaderKey: 'x-goog-api-key',
	defaultHeaders: {
		'content-type': 'application/json',
	},
	chatCompletionPathPrefix: '/v1beta/models',
	defaultTemperature: 0.0,
	streamingSupport: true,
};

export const huggingfaceProviderInfo: ProviderInfo = {
	name: ProviderName.HUGGINGFACE,
	apiKey: '',
	engine: '',
	defaultOrigin: 'https://api-inference.huggingface.co',
	defaultModel: ModelName.DEEPSEEK_CODER_1_3B_INSTRUCT,
	additionalSettings: {},
	timeout: 120,
	apiKeyHeaderKey: 'Authorization',
	defaultHeaders: {
		'content-type': 'application/json',
	},
	chatCompletionPathPrefix: '/models',
	defaultTemperature: 0.0,
	modelPrefixes: ['microsoft/', 'replit/', 'Salesforce/', 'bigcode/', 'deepseek-ai/'],
	streamingSupport: false,
};

export const llamacppProviderInfo: ProviderInfo = {
	name: ProviderName.LLAMACPP,
	apiKey: '',
	engine: '',
	defaultOrigin: 'http://127.0.0.1:8080',
	defaultModel: ModelName.LLAMA_3,
	additionalSettings: {},
	timeout: 120,
	apiKeyHeaderKey: 'Authorization',
	defaultHeaders: {
		'content-type': 'application/json',
	},
	chatCompletionPathPrefix: '/completion',
	defaultTemperature: 0.0,
	streamingSupport: false,
};

export const openaiProviderInfo: ProviderInfo = {
	name: ProviderName.OPENAI,
	apiKey: '',
	engine: '',
	defaultOrigin: 'https://api.openai.com/v1',
	defaultModel: ModelName.GPT_4O_MINI,
	additionalSettings: {},
	timeout: 120,
	apiKeyHeaderKey: 'Authorization',
	defaultHeaders: {
		'content-type': 'application/json',
	},
	chatCompletionPathPrefix: '/chat/completions',
	defaultTemperature: 0.0,
	streamingSupport: true,
};

export const ANTHROPIC_MODELS: { [key in ModelName]?: ModelInfo } = {
	[ModelName.CLAUDE_3_5_SONNET]: {
		name: ModelName.CLAUDE_3_5_SONNET,
		displayName: 'Claude 3.5 Sonnet',
		provider: ProviderName.ANTHROPIC,
		maxPromptLength: 8192,
		maxOutputLength: 8192,
	},
	[ModelName.CLAUDE_3_OPUS]: {
		name: ModelName.CLAUDE_3_OPUS,
		displayName: 'Claude 3 Opus',
		provider: ProviderName.ANTHROPIC,
		maxPromptLength: 8192,
		maxOutputLength: 8192,
	},
	[ModelName.CLAUDE_3_SONNET]: {
		name: ModelName.CLAUDE_3_SONNET,
		displayName: 'Claude 3 Sonnet',
		provider: ProviderName.ANTHROPIC,
		maxPromptLength: 8192,
		maxOutputLength: 8192,
	},
	[ModelName.CLAUDE_3_HAIKU]: {
		name: ModelName.CLAUDE_3_HAIKU,
		displayName: 'Claude 3 Haiku',
		provider: ProviderName.ANTHROPIC,
		maxPromptLength: 8192,
		maxOutputLength: 8192,
	},
};

export const GOOGLE_MODELS: { [key in ModelName]?: ModelInfo } = {
	[ModelName.GEMINI_1_5_FLASH]: {
		name: ModelName.GEMINI_1_5_FLASH,
		displayName: 'Google Gemini 1.5 Flash',
		provider: ProviderName.GOOGLE,
		maxPromptLength: 4096,
		maxOutputLength: 8192,
	},
	[ModelName.GEMINI_1_5_PRO]: {
		name: ModelName.GEMINI_1_5_PRO,
		displayName: 'Google Gemini 1.5 Pro',
		provider: ProviderName.GOOGLE,
		maxPromptLength: 4096,
		maxOutputLength: 8192,
	},
};

export const HUGGINGFACE_MODELS: { [key in ModelName]?: ModelInfo } = {
	[ModelName.DEEPSEEK_CODER_1_3B_INSTRUCT]: {
		name: ModelName.DEEPSEEK_CODER_1_3B_INSTRUCT,
		displayName: 'HF Deepseek Coder 1.3b',
		provider: ProviderName.HUGGINGFACE,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
	},
};

export const LLAMACPP_MODELS: { [key in ModelName]?: ModelInfo } = {
	[ModelName.LLAMA_3]: {
		name: ModelName.LLAMA_3,
		displayName: 'Llama 3',
		provider: ProviderName.LLAMACPP,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
	},
	[ModelName.LLAMA_3_1]: {
		name: ModelName.LLAMA_3_1,
		displayName: 'Llama 3.1',
		provider: ProviderName.LLAMACPP,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
	},
};

export const OPENAI_MODELS: { [key in ModelName]?: ModelInfo } = {
	[ModelName.GPT_O1]: {
		name: ModelName.GPT_O1,
		displayName: 'OpenAI GPT o1',
		provider: ProviderName.OPENAI,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
		defaultTemperature: 1,
		streamingSupport: false,
	},
	[ModelName.GPT_O1_PREVIEW]: {
		name: ModelName.GPT_O1_PREVIEW,
		displayName: 'OpenAI GPT o1 preview',
		provider: ProviderName.OPENAI,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
		defaultTemperature: 1,
		streamingSupport: false,
	},
	[ModelName.GPT_O1_MINI]: {
		name: ModelName.GPT_O1_MINI,
		displayName: 'OpenAI GPT o1 mini',
		provider: ProviderName.OPENAI,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
		defaultTemperature: 1,
		streamingSupport: false,
	},
	[ModelName.GPT_4O]: {
		name: ModelName.GPT_4O,
		displayName: 'OpenAI GPT 4o',
		provider: ProviderName.OPENAI,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
	},
	[ModelName.GPT_4]: {
		name: ModelName.GPT_4,
		displayName: 'OpenAI GPT 4',
		provider: ProviderName.OPENAI,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
	},
	[ModelName.GPT_3_5_TURBO]: {
		name: ModelName.GPT_3_5_TURBO,
		displayName: 'OpenAI GPT 3.5 turbo',
		provider: ProviderName.OPENAI,
		maxPromptLength: 2400,
		maxOutputLength: 2400,
	},
	[ModelName.GPT_4O_MINI]: {
		name: ModelName.GPT_4O_MINI,
		displayName: 'OpenAI GPT 4o mini',
		provider: ProviderName.OPENAI,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
	},
};

export const ALL_MODEL_INFO: { [key in ModelName]: ModelInfo } = {
	...ANTHROPIC_MODELS,
	...GOOGLE_MODELS,
	...HUGGINGFACE_MODELS,
	...LLAMACPP_MODELS,
	...OPENAI_MODELS,
} as { [key in ModelName]: ModelInfo };

export const ALL_AI_PROVIDERS: Record<ProviderName, ProviderInfo> = {
	[ProviderName.ANTHROPIC]: anthropicProviderInfo,
	[ProviderName.GOOGLE]: googleProviderInfo,
	[ProviderName.HUGGINGFACE]: huggingfaceProviderInfo,
	[ProviderName.LLAMACPP]: llamacppProviderInfo,
	[ProviderName.OPENAI]: openaiProviderInfo,
};
