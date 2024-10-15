import { ProviderName } from './provider_types';

export const anthropicProviderInfo = {
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
};

export const googleProviderInfo = {
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
};

export const huggingfaceProviderInfo = {
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
};

export const llamacppProviderInfo = {
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
};

export const openaiProviderInfo = {
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
};

export const ALL_AI_PROVIDERS_DESCRIPTION: Record<ProviderName, Record<string, string>> = {
	[ProviderName.ANTHROPIC]: anthropicProviderInfo,
	[ProviderName.GOOGLE]: googleProviderInfo,
	[ProviderName.HUGGINGFACE]: huggingfaceProviderInfo,
	[ProviderName.LLAMACPP]: llamacppProviderInfo,
	[ProviderName.OPENAI]: openaiProviderInfo,
};
