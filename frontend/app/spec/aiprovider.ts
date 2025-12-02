import type { Attachment } from '@/spec/attachment';
import type { ModelName, ProviderName, ReasoningParams } from '@/spec/modelpreset';
import type { Tool, ToolChoice } from '@/spec/tool';

export enum ChatCompletionRoleEnum {
	system = 'system',
	user = 'user',
	assistant = 'assistant',
	function = 'function',
}

export interface ChatCompletionDataMessage {
	role: ChatCompletionRoleEnum;
	content?: string | null;
	name?: string | null;
	attachments?: Array<Attachment>;
}

/**
 * @public
 */
export type FetchCompletionToolChoice = ToolChoice & {
	tool: Tool | null;
};

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

interface APIRequestDetails {
	url?: string;
	method?: string;
	headers?: any;
	params?: any;
	data?: any;
	timeout?: number;
	curlCommand?: string;
}

interface APIResponseDetails {
	data: any;
	status: number;
	headers: any;
	requestDetails?: APIRequestDetails;
}

interface APIErrorDetails {
	message: string;
	requestDetails?: APIRequestDetails;
	responseDetails?: APIResponseDetails;
}

export enum ResponseContentType {
	Text = 'text',
	Thinking = 'thinking',
	ThinkingSummary = 'thinkingSummary',
	Image = 'image',
}

interface ResponseContent {
	type: ResponseContentType;
	content: string;
}

/**
 * @public
 */
export interface ResponseToolCall {
	id: string;
	callID: string;
	name: string;
	arguments: string;
	type: string;
	status?: string;
}

export interface FetchCompletionData {
	modelParams: ModelParams;
	messages?: Array<ChatCompletionDataMessage>;
	toolChoices?: Array<FetchCompletionToolChoice>;
}

export interface FetchCompletionResponseBody {
	requestDetails?: APIRequestDetails;
	responseDetails?: APIResponseDetails;
	errorDetails?: APIErrorDetails;
	responseContent?: ResponseContent[];
	toolCalls?: ResponseToolCall[];
}

export interface IProviderSetAPI {
	buildCompletionData(
		provider: ProviderName,
		modelParams: ModelParams,
		currentMessage: ChatCompletionDataMessage,
		prevMessages?: Array<ChatCompletionDataMessage>,
		toolChoices?: Array<ToolChoice>
	): Promise<FetchCompletionData>;
	completion(
		provider: ProviderName,
		completionData: FetchCompletionData,
		requestId?: string,
		signal?: AbortSignal,
		onStreamTextData?: (textData: string) => void,
		onStreamThinkingData?: (thinkingData: string) => void
	): Promise<FetchCompletionResponseBody | undefined>;
	cancelCompletion(requestId: string): Promise<void>;
}
