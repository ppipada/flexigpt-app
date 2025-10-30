import type { ModelName, ProviderName, ReasoningParams } from '@/spec/modelpreset';

export enum ChatCompletionRoleEnum {
	system = 'system',
	user = 'user',
	assistant = 'assistant',
	function = 'function',
}

export interface ChatCompletionDataMessage {
	role: ChatCompletionRoleEnum;
	content?: string;
	name?: string;
	toolAttachments?: ChatCompletionToolAttachment[];
}

export interface ChatCompletionResponseMessage {
	role: ChatCompletionRoleEnum;
	content?: string;
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

export interface ChatCompletionToolAttachment {
	bundleID: string;
	toolSlug: string;
	toolVersion: string;
	id?: string;
}
export interface CompletionData {
	modelParams: ModelParams;
	messages?: ChatCompletionDataMessage[];
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
