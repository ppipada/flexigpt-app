import type { ModelName, ProviderName, ReasoningParams } from '@/spec/modelpreset';
import type { Tool } from '@/spec/tool';

export enum ChatCompletionRoleEnum {
	system = 'system',
	user = 'user',
	assistant = 'assistant',
	function = 'function',
}

// ChatCompletionAttachmentKind enumerates contextual attachment categories that can be
// associated with messages sent to the inference layer.
export enum ChatCompletionAttachmentKind {
	file = 'file',
	docIndex = 'docIndex',
	pr = 'pr',
	commit = 'commit',
	snapshot = 'snapshot',
}

// ChatCompletionAttachment is a lightweight reference to external context (files, PRs, snapshots, etc.).
export interface ChatCompletionAttachment {
	kind: ChatCompletionAttachmentKind;
	ref: string;
	label: string;
	/** Optional size in bytes for verified file attachments. */
	sizeBytes?: number;
	/** ISO timestamp of last modification for verified file attachments. */
	modTime?: string;
	/** Whether the backend could successfully verify this attachment (e.g. via stat). */
	exists?: boolean;
}

export interface ChatCompletionToolChoice {
	bundleID?: string;
	toolSlug: string;
	toolVersion: string;
	id?: string;
	description: string;
	displayName: string;
}

export interface ChatCompletionDataMessage {
	role: ChatCompletionRoleEnum;
	content?: string | null;
	name?: string | null;
}

/**
 * @public
 */
export interface FetchCompletionToolChoice {
	bundleID?: string;
	toolSlug: string;
	toolVersion: string;
	id?: string;
	tool: Tool | null;
	description: string;
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

export interface FetchCompletionData {
	modelParams: ModelParams;
	messages?: Array<ChatCompletionDataMessage>;
	toolChoices?: Array<FetchCompletionToolChoice>;
	attachments?: Array<ChatCompletionAttachment>;
}

export interface FetchCompletionResponseBody {
	requestDetails?: APIRequestDetails;
	responseDetails?: APIResponseDetails;
	errorDetails?: APIErrorDetails;
	responseContent?: ResponseContent[];
}

export interface IProviderSetAPI {
	buildCompletionData(
		provider: ProviderName,
		modelParams: ModelParams,
		currentMessage: ChatCompletionDataMessage,
		prevMessages?: Array<ChatCompletionDataMessage>,
		toolChoices?: Array<ChatCompletionToolChoice>,
		attachments?: Array<ChatCompletionAttachment>
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
