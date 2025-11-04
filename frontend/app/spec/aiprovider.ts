import type { ModelName, ProviderName, ReasoningParams } from '@/spec/modelpreset';
import type { Tool } from '@/spec/tool';

export enum ChatCompletionRoleEnum {
	system = 'system',
	user = 'user',
	assistant = 'assistant',
	function = 'function',
}

interface ChatCompletionDataMessage {
	role: ChatCompletionRoleEnum;
	content?: string;
	name?: string;
	toolAttachments?: ChatCompletionToolAttachment[];
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

interface ChatCompletionToolAttachment {
	bundleID: string;
	toolSlug: string;
	toolVersion: string;
	id?: string;
}

export interface BuildCompletionToolAttachment {
	bundleID?: string;
	toolSlug: string;
	toolVersion: string;
	id?: string;
	displayName?: string;
}

export interface BuildCompletionDataMessage {
	role: ChatCompletionRoleEnum;
	content?: string;
	name?: string;
	toolAttachments?: BuildCompletionToolAttachment[];
}

interface CompletionTool {
	bundleID: string;
	tool: Tool;
	attachmentIDs?: string[];
}

export interface CompletionData {
	modelParams: ModelParams;
	messages?: ChatCompletionDataMessage[];
	tools?: CompletionTool[];
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
}

interface ResponseContent {
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
		modelParams: ModelParams,
		currentMessage: BuildCompletionDataMessage,
		prevMessages?: Array<BuildCompletionDataMessage>
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
