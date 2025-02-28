import { ChatCompletionRequestMessage, CompletionRequest, CompletionResponse } from './chat_types';

export type ModelName = string;
export const DefaultModelName: ModelName = 'gpt-4o';
export const DefaultModelTitle = 'OpenAI GPT 4o';

export type ProviderName = string;
export const DefaultProviderName: ProviderName = 'openai';

export interface ModelInfo {
	name: ModelName;
	displayName: string;
	provider: ProviderName;
	maxPromptLength: number;
	maxOutputLength: number;
	defaultTemperature: number;
}

export interface ProviderInfo {
	name: ProviderName;
	apiKey: string;
	engine: string;
	defaultOrigin: string;
	defaultModel: ModelName;
	additionalSettings: Record<string, any>;
	timeout: number;
	apiKeyHeaderKey: string;
	defaultHeaders: Record<string, string>;
	chatCompletionPathPrefix: string;
	defaultTemperature: number;
	streamingSupport: boolean;
	models: Record<string, ModelInfo>;
	modelPrefixes?: string[];
}

export interface IProviderSetAPI {
	getDefaultProvider(): Promise<ProviderName>;
	setDefaultProvider(provider: ProviderName): Promise<void>;
	getConfigurationInfo(): Promise<Record<string, any>>;
	setAttribute(
		provider: ProviderName,
		apiKey?: string,
		defaultModel?: ModelName,
		defaultTemperature?: number,
		defaultOrigin?: string
	): Promise<void>;

	getCompletionRequest(
		provider: ProviderName,
		prompt: string,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		inputParams?: { [key: string]: any }
	): Promise<CompletionRequest>;
	completion(
		provider: ProviderName,
		input: CompletionRequest,
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined>;
}

export const ProviderInfoDescription = {
	apiKey: 'Your API key for the provider.',
	engine: 'The engine to be used for processing. Is present with Azure etc.',
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
