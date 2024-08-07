import { ModelInfo, ModelName, ProviderInfo, ProviderInfoImpl, ProviderName } from './provider_types';

export const anthropicProviderInfo: ProviderInfo = new ProviderInfoImpl({
	apiKey: '',
	engine: '',
	defaultOrigin: 'https://api.anthropic.com',
	defaultModel: ModelName.CLAUDE_3_HAIKU,
	additionalSettings: {},
	timeout: 120,
	apiKeyHeaderKey: 'x-api-key',
	defaultHeaders: {
		'content-type': 'application/json',
		accept: 'application/json',
		'anthropic-version': '2023-06-01',
	},
	chatCompletionPathPrefix: '/v1/messages',
	defaultTemperature: 0.1,
	streamingSupport: true,
	descriptions: {
		apiKey: 'Your anthropic API key.',
		engine: 'The engine to be used for processing.',
		defaultOrigin:
			'Default origin to use for requests. This can be used to talk to any server that serves a compatible API',
		defaultModel: 'Default model to use for chat requests',
		additionalSettings: 'Any additional settings to pass to the model. Input as a JSON object',
		timeout: 'The timeout duration in milliseconds.',
		apiKeyHeaderKey: 'The header key for the API key.',
		defaultHeaders: 'The default headers to be included in requests.',
		chatCompletionPathPrefix: 'The path prefix for chat completions.',
		defaultTemperature: 'Default temperature setting for chat requests',
		modelPrefixes: 'Optional prefixes for models.',
	},
});

export const googleProviderInfo: ProviderInfo = new ProviderInfoImpl({
	apiKey: '',
	engine: '',
	defaultOrigin: 'https://generativelanguage.googleapis.com',
	defaultModel: ModelName.GEMINI_1_5_FLASH,
	additionalSettings: {},
	timeout: 120,
	apiKeyHeaderKey: 'Authorization',
	defaultHeaders: {
		'content-type': 'application/json',
	},
	chatCompletionPathPrefix: '/v1beta',
	defaultTemperature: 0.1,
	streamingSupport: false,
	descriptions: {
		apiKey: 'Your google generative AI API key.',
		engine: 'The engine to be used for processing.',
		defaultOrigin:
			'Default origin to use for requests. This can be used to talk to any server that serves a compatible API',
		defaultModel: 'Default model to use for chat requests',
		additionalSettings: 'Any additional settings to pass to the model. Input as a JSON object',
		timeout: 'The timeout duration in milliseconds.',
		apiKeyHeaderKey: 'The header key for the API key.',
		defaultHeaders: 'The default headers to be included in requests.',
		chatCompletionPathPrefix: 'The path prefix for chat completions.',
		defaultTemperature: 'Default temperature setting for chat requests',
		modelPrefixes: 'Optional prefixes for models.',
	},
});

export const huggingfaceProviderInfo: ProviderInfo = new ProviderInfoImpl({
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
	defaultTemperature: 0.1,
	modelPrefixes: ['microsoft/', 'replit/', 'Salesforce/', 'bigcode/', 'deepseek-ai/'],
	streamingSupport: false,
	descriptions: {
		apiKey: 'Your huggingface API key.',
		engine: 'The engine to be used for processing.',
		defaultOrigin:
			'Default origin to use for requests. This can be used to talk to any server that serves a compatible API',
		defaultModel: 'Default model to use for chat requests',
		additionalSettings: 'Any additional settings to pass to the model. Input as a JSON object',
		timeout: 'The timeout duration in milliseconds.',
		apiKeyHeaderKey: 'The header key for the API key.',
		defaultHeaders: 'The default headers to be included in requests.',
		chatCompletionPathPrefix: 'The path prefix for chat completions.',
		defaultTemperature: 'Default temperature setting for chat requests',
		modelPrefixes: 'Optional prefixes for models.',
	},
});

export const llamacppProviderInfo: ProviderInfo = new ProviderInfoImpl({
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
	defaultTemperature: 0.1,
	streamingSupport: false,
	descriptions: {
		apiKey: 'Your llamacpp API key.',
		engine: 'The engine to be used for processing.',
		defaultOrigin:
			'Default origin to use for requests. This can be used to talk to any server that serves a compatible API',
		defaultModel: 'Default model to use for chat requests',
		additionalSettings: 'Any additional settings to pass to the model. Input as a JSON object',
		timeout: 'The timeout duration in milliseconds.',
		apiKeyHeaderKey: 'The header key for the API key.',
		defaultHeaders: 'The default headers to be included in requests.',
		chatCompletionPathPrefix: 'The path prefix for chat completions.',
		defaultTemperature: 'Default temperature setting for chat requests',
		modelPrefixes: 'Optional prefixes for models.',
	},
});

export const openaiProviderInfo: ProviderInfo = new ProviderInfoImpl({
	apiKey: '',
	engine: '',
	defaultOrigin: 'https://api.openai.com',
	defaultModel: ModelName.GPT_4O_MINI,
	additionalSettings: {},
	timeout: 120,
	apiKeyHeaderKey: 'Authorization',
	defaultHeaders: {
		'content-type': 'application/json',
	},
	chatCompletionPathPrefix: '/v1/chat/completions',
	defaultTemperature: 0.1,
	streamingSupport: true,
	descriptions: {
		apiKey: 'Your openAI API key. Can be seen at https://beta.openai.com/account/api-keys',
		engine: 'The engine to be used for processing.',
		defaultOrigin:
			'Default origin to use for requests. This can be used to talk to any server that serves a compatible API',
		defaultModel: 'Default model to use for chat requests',
		additionalSettings: 'Any additional settings to pass to the model. Input as a JSON object',
		timeout: 'The timeout duration in milliseconds.',
		apiKeyHeaderKey: 'The header key for the API key.',
		defaultHeaders: 'The default headers to be included in requests.',
		chatCompletionPathPrefix: 'The path prefix for chat completions.',
		defaultTemperature: 'Default temperature setting for chat requests',
		modelPrefixes: 'Optional prefixes for models.',
	},
});

export const ANTHROPIC_MODELS: { [key in ModelName]?: ModelInfo } = {
	[ModelName.CLAUDE_3_5_SONNET]: {
		name: ModelName.CLAUDE_3_5_SONNET,
		provider: ProviderName.ANTHROPIC,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
		defaultTemperature: 0.1,
	},
	[ModelName.CLAUDE_3_OPUS]: {
		name: ModelName.CLAUDE_3_OPUS,
		provider: ProviderName.ANTHROPIC,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
		defaultTemperature: 0.1,
	},
	[ModelName.CLAUDE_3_SONNET]: {
		name: ModelName.CLAUDE_3_SONNET,
		provider: ProviderName.ANTHROPIC,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
		defaultTemperature: 0.1,
	},
	[ModelName.CLAUDE_3_HAIKU]: {
		name: ModelName.CLAUDE_3_HAIKU,
		provider: ProviderName.ANTHROPIC,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
		defaultTemperature: 0.1,
	},
};

export const GOOGLE_MODELS: { [key in ModelName]?: ModelInfo } = {
	[ModelName.GEMINI_1_5_FLASH]: {
		name: ModelName.GEMINI_1_5_FLASH,
		provider: ProviderName.GOOGLE,
		maxPromptLength: 4096,
		maxOutputLength: 8192,
		defaultTemperature: 0.1,
	},
	[ModelName.GEMINI_1_5_PRO]: {
		name: ModelName.GEMINI_1_5_PRO,
		provider: ProviderName.GOOGLE,
		maxPromptLength: 4096,
		maxOutputLength: 8192,
		defaultTemperature: 0.1,
	},
};

export const HUGGINGFACE_MODELS: { [key in ModelName]?: ModelInfo } = {
	[ModelName.DEEPSEEK_CODER_1_3B_INSTRUCT]: {
		name: ModelName.DEEPSEEK_CODER_1_3B_INSTRUCT,
		provider: ProviderName.HUGGINGFACE,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
		defaultTemperature: 0.1,
	},
};

export const LLAMACPP_MODELS: { [key in ModelName]?: ModelInfo } = {
	[ModelName.LLAMA_3]: {
		name: ModelName.LLAMA_3,
		provider: ProviderName.LLAMACPP,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
		defaultTemperature: 0.1,
	},
	[ModelName.LLAMA_3_1]: {
		name: ModelName.LLAMA_3_1,
		provider: ProviderName.LLAMACPP,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
		defaultTemperature: 0.1,
	},
};

export const OPENAI_MODELS: { [key in ModelName]?: ModelInfo } = {
	[ModelName.GPT_4O]: {
		name: ModelName.GPT_4O,
		provider: ProviderName.OPENAI,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
		defaultTemperature: 0.1,
	},
	[ModelName.GPT_4]: {
		name: ModelName.GPT_4,
		provider: ProviderName.OPENAI,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
		defaultTemperature: 0.1,
	},
	[ModelName.GPT_3_5_TURBO]: {
		name: ModelName.GPT_3_5_TURBO,
		provider: ProviderName.OPENAI,
		maxPromptLength: 2400,
		maxOutputLength: 2400,
		defaultTemperature: 0.1,
	},
	[ModelName.GPT_4O_MINI]: {
		name: ModelName.GPT_4O_MINI,
		provider: ProviderName.OPENAI,
		maxPromptLength: 4096,
		maxOutputLength: 4096,
		defaultTemperature: 0.1,
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
