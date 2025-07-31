import type { ModelParams, ProviderInfo, ProviderName } from '@/spec/modelpreset';

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
