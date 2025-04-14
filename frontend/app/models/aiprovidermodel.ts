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
	inbuiltProviderModels: Record<ProviderName, Record<ModelName, ModelParams>>;
}

export interface AddProviderRequest {
	provider: ProviderName;
	apiKey: string;
	origin: string;
	chatCompletionPathPrefix: string;
}

export interface IProviderSetAPI {
	setDefaultProvider(provider: ProviderName): Promise<void>;
	getConfigurationInfo(): Promise<ConfigurationResponse>;
	addProvider(providerInfo: AddProviderRequest): Promise<void>;
	deleteProvider(provider: ProviderName): Promise<void>;
	setProviderAPIKey(provider: ProviderName, apiKey: string): Promise<void>;
	setProviderAttribute(provider: ProviderName, origin?: string, chatCompletionPathPrefix?: string): Promise<void>;
	completion(
		provider: ProviderName,
		prompt: string,
		modelParams: ModelParams,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		onStreamData?: (data: string) => void
	): Promise<CompletionResponse | undefined>;
}
