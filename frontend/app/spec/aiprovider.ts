import type { Attachment } from '@/spec/attachment';
import type {
	CompletionUsage,
	ModelName,
	ProviderName,
	ReasoningContent,
	ReasoningParams,
	RoleEnum,
} from '@/spec/modelpreset';
import type { Tool, ToolCall, ToolChoice, ToolOutput } from '@/spec/tool';

export interface ChatCompletionDataMessage {
	role: RoleEnum;
	content?: string | null;
	name?: string | null;

	reasoningContents?: ReasoningContent[];
	attachments?: Attachment[];
	toolCalls?: ToolCall[];
	toolOutputs?: ToolOutput[];
}

/**
 * @public
 */
export type FetchCompletionToolChoice = ToolChoice & {
	tool: Tool | null;
};

export interface ModelParam {
	name: ModelName;
	stream: boolean;
	maxPromptLength: number;
	maxOutputLength: number;
	temperature?: number;
	reasoning?: ReasoningParams; // This is from modelpresetspec and not inferencego.
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

export interface FetchCompletionData {
	modelParams: ModelParam;
	messages?: Array<ChatCompletionDataMessage>;
	toolChoices?: Array<FetchCompletionToolChoice>;
}

export interface FetchCompletionResponseBody {
	requestDetails?: APIRequestDetails;
	responseDetails?: APIResponseDetails;
	errorDetails?: APIErrorDetails;

	content?: string;

	reasoningContents?: ReasoningContent[];
	toolCalls?: ToolCall[];
	usage?: CompletionUsage;
}

export interface IProviderSetAPI {
	buildCompletionData(
		provider: ProviderName,
		modelParams: ModelParam,
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
