export type ModelName = string;
export const DefaultModelName: ModelName = 'gpt-4o';
export const DefaultModelTitle = 'OpenAI GPT 4o';

export type ProviderName = string;
export const DefaultProviderName: ProviderName = 'openai';

export interface ModelInfo {
	name: ModelName;
	displayName: string;
	provider: string;
	maxPromptLength: number;
	maxOutputLength: number;
	defaultTemperature: number;
	streamingSupport: boolean;
	reasoningSupport: boolean;
	defaultSystemPrompt: string;
	timeout: number;
}

export interface ProviderInfo {
	name: ProviderName;
	apiKey: string;
	defaultModel: ModelName;
	engine: string;
	origin: string;

	apiKeyHeaderKey: string;
	defaultHeaders: Record<string, string>;
	chatCompletionPathPrefix: string;
	modelPrefixes?: string[];
	models: Record<ModelName, ModelInfo>;
}

export const ProviderInfoDescription = {
	apiKey: 'Your API key for the provider.',
	engine: 'The engine to be used for processing. Is present with Azure etc.',
	origin: 'Origin/URL to use for requests. This can be used to talk to any server that serves a compatible API',
	defaultModel: 'Default model to use for chat requests',
	timeout: 'The timeout duration in milliseconds.',
	apiKeyHeaderKey: 'The header key for the API key.',
	defaultHeaders: 'The default headers to be included in requests.',
	chatCompletionPathPrefix: 'The path prefix for chat completions.',
	modelPrefixes: 'Optional prefixes for models.',
};

export enum ChatCompletionRoleEnum {
	system = 'system',
	user = 'user',
	assistant = 'assistant',
	function = 'function',
}

export interface ChatCompletionFunctions {
	name: string;
	description?: string;
	parameters?: { [key: string]: any };
}

export interface ChatCompletionRequestMessageFunctionCall {
	name?: string;
	arguments?: string;
}

export interface ChatCompletionRequestMessage {
	role: ChatCompletionRoleEnum;
	content?: string;
	name?: string;
	functionCall?: ChatCompletionRequestMessageFunctionCall;
}

export interface ChatCompletionResponseMessage {
	role: ChatCompletionRoleEnum;
	content?: string;
	functionCall?: ChatCompletionRequestMessageFunctionCall;
}

export type CreateChatCompletionRequestFunctionCall = CreateChatCompletionRequestFunctionCallOneOf | string;

export interface CreateChatCompletionRequestFunctionCallOneOf {
	name: string;
}

export interface ModelParams {
	name: string;
	stream?: boolean;
	promptLength?: number;
	outputLength?: number;
	temperature?: number;
	reasoningSupport?: boolean;
	systemPrompt?: string;
	timeout?: number;
	additionalParameters?: Record<string, any>;
}

export interface CompletionRequest {
	modelParams: ModelParams;
	messages?: ChatCompletionRequestMessage[];
	functions?: ChatCompletionFunctions[];
	functionCall?: CreateChatCompletionRequestFunctionCall;
}

export interface APIRequestDetails {
	url?: string;
	method?: string;
	headers?: any;
	params?: any;
	data?: any;
	timeout?: number;
	curlCommand?: string;
}

export interface APIResponseDetails {
	data: any;
	status: number;
	headers: any;
	requestDetails?: APIRequestDetails;
}

export interface APIErrorDetails {
	message: string;
	requestDetails?: APIRequestDetails;
	responseDetails?: APIResponseDetails;
}

export interface APIFetchResponse<T> {
	data?: T;
	responseDetails?: APIResponseDetails;
	requestDetails?: APIRequestDetails;
	errorDetails?: APIErrorDetails;
}

export interface CompletionResponse {
	requestDetails?: APIRequestDetails;
	responseDetails?: APIResponseDetails;
	errorDetails?: APIErrorDetails;
	respContent?: string;
	functionName?: string;
	functionArgs?: any;
}

export interface ConfigurationResponse {
	defaultProvider: ProviderName;
	configuredProviders: Record<ProviderName, ProviderInfo>;
}

export interface IProviderSetAPI {
	setDefaultProvider(provider: ProviderName): Promise<void>;
	getConfigurationInfo(): Promise<ConfigurationResponse>;
	setAttribute(provider: ProviderName, apiKey?: string, defaultModel?: ModelName, origin?: string): Promise<void>;

	getCompletionRequest(
		provider: ProviderName,
		prompt: string,
		modelParams: ModelParams,
		prevMessages?: Array<ChatCompletionRequestMessage>
	): Promise<CompletionRequest>;
	completion(
		provider: ProviderName,
		input: CompletionRequest,
		onStreamData?: (data: string) => void
	): Promise<CompletionResponse | undefined>;
}
