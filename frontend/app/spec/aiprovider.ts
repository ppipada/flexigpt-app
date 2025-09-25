import type { ModelName, ProviderName, ReasoningParams } from '@/spec/modelpreset';

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

export interface ChatCompletionDataMessageFunctionCall {
	name?: string;
	arguments?: string;
}

export interface ChatCompletionDataMessage {
	role: ChatCompletionRoleEnum;
	content?: string;
	name?: string;
	functionCall?: ChatCompletionDataMessageFunctionCall;
}

export interface ChatCompletionResponseMessage {
	role: ChatCompletionRoleEnum;
	content?: string;
	functionCall?: ChatCompletionDataMessageFunctionCall;
}

export type CreateChatCompletionDataFunctionCall = CreateChatCompletionDataFunctionCallOneOf | string;

export interface CreateChatCompletionDataFunctionCallOneOf {
	name: string;
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
	additionalParametersRawJSON?: string;
}

export interface CompletionData {
	modelParams: ModelParams;
	messages?: ChatCompletionDataMessage[];
	functions?: ChatCompletionFunctions[];
	functionCall?: CreateChatCompletionDataFunctionCall;
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

export enum ResponseContentType {
	Text = 'text',
	Thinking = 'thinking',
	ThinkingSummary = 'thinkingSummary',
}

export interface ResponseContent {
	type: ResponseContentType;
	content: string;
}

export interface CompletionResponse {
	requestDetails?: APIRequestDetails;
	responseDetails?: APIResponseDetails;
	errorDetails?: APIErrorDetails;
	responseContent?: ResponseContent[];
}

export interface IProviderSetAPI {
	buildCompletionData(
		provider: ProviderName,
		prompt: string,
		modelParams: ModelParams,
		prevMessages?: Array<ChatCompletionDataMessage>
	): Promise<CompletionData>;
	completion(
		provider: ProviderName,
		completionData: CompletionData,
		requestId?: string,
		signal?: AbortSignal,
		onStreamTextData?: (textData: string) => void,
		onStreamThinkingData?: (thinkingData: string) => void
	): Promise<CompletionResponse | undefined>;
	cancelCompletion(requestId: string): Promise<void>;
}
