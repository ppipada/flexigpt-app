export type ModelName = string;
export const DefaultModelName: ModelName = 'gpt-4o';
export const DefaultModelTitle = 'OpenAI GPT 4o';

export type ProviderName = string;
export const DefaultProviderName: ProviderName = 'openai';

// Define the ReasoningType enum
export enum ReasoningType {
	HybridWithTokens = 'hybridWithTokens',
	SingleWithLevels = 'singleWithLevels',
}

// Define the ReasoningLevel enum
export enum ReasoningLevel {
	Low = 'low',
	Medium = 'medium',
	High = 'high',
}

// Define the ReasoningParams interface
export interface ReasoningParams {
	type: ReasoningType;
	level: ReasoningLevel;
	tokens: number;
}

export interface ModelParams {
	name: ModelName;
	stream: boolean;
	maxPromptLength: number;
	maxOutputLength: number;
	temperature?: number;
	reasoning?: ReasoningParams;
	systemPrompt: string;
	timeout: number;
	additionalParameters: Record<string, any>;
}

export const DefaultModelParams: ModelParams = {
	name: '',
	stream: false,
	maxPromptLength: 2048,
	maxOutputLength: 1024,
	temperature: 0.1,
	reasoning: {
		type: ReasoningType.SingleWithLevels,
		level: ReasoningLevel.Medium,
		tokens: 1024,
	},
	systemPrompt: '',
	timeout: 60,
	additionalParameters: {},
};

export interface ProviderInfo {
	name: ProviderName;
	apiKey: string;
	origin: string;
	chatCompletionPathPrefix: string;
	apiKeyHeaderKey: string;
	defaultHeaders: Record<string, string>;
}

export const ProviderInfoDescription = {
	apiKey: 'Your API key for the provider.',
	origin: 'Origin/URL to use for requests. This can be used to talk to any server that serves a compatible API',
	defaultModel: 'Default model to use for chat requests',
	timeout: 'The timeout duration in milliseconds.',
	apiKeyHeaderKey: 'The header key for the API key.',
	defaultHeaders: 'The default headers to be included in requests.',
	chatCompletionPathPrefix: 'The path prefix for chat completions.',
	modelPrefixes: 'Optional prefixes for models.',
};

export interface ModelPreset {
	name: string;
	displayName: string;
	shortCommand: string;
	isEnabled: boolean;
	stream?: boolean;
	maxPromptLength?: number;
	maxOutputLength?: number;
	temperature?: number;
	reasoning?: ReasoningParams;
	systemPrompt?: string;
	timeout?: number;
	additionalParameters?: Record<string, any>;
}

export const DefaultModelPreset: ModelPreset = {
	name: '',
	displayName: '',
	shortCommand: '',
	isEnabled: true,
	stream: DefaultModelParams.stream,
	maxPromptLength: DefaultModelParams.maxPromptLength,
	maxOutputLength: DefaultModelParams.maxOutputLength,
	temperature: DefaultModelParams.temperature ?? 0.1,
	reasoning: DefaultModelParams.reasoning,
	systemPrompt: DefaultModelParams.systemPrompt,
	timeout: DefaultModelParams.timeout,
	additionalParameters: DefaultModelParams.additionalParameters,
};

export interface ChatOptions extends ModelParams {
	title: string;
	provider: ProviderName;
	disablePreviousMessages: boolean;
}

export const DefaultChatOptions: ChatOptions = {
	...DefaultModelParams,
	provider: 'No Provider',
	name: 'No Model',
	title: 'No Model configured',
	disablePreviousMessages: false,
};

export type ProviderModelPresets = Record<ModelName, ModelPreset>;

export type ModelPresetsSchema = {
	version: string;
	modelPresets: Record<ProviderName, ProviderModelPresets>;
};

export interface IModelPresetStoreAPI {
	getAllModelPresets: () => Promise<ModelPresetsSchema>;
	createModelPresets: (providerName: ProviderName, providerModelPresets: ProviderModelPresets) => Promise<void>;
	deleteModelPresets: (providerName: ProviderName) => Promise<void>;
	addModelPreset: (providerName: ProviderName, modelName: ModelName, modelPreset: ModelPreset) => Promise<void>;
	deleteModelPreset: (providerName: ProviderName, modelName: ModelName) => Promise<void>;
}
