import { type ChatCompletionDataMessage, type FetchCompletionResponseBody, type ModelParam } from '@/spec/aiprovider';
import type { ConversationMessage } from '@/spec/conversation';
import { type CompletionUsage, type ProviderName, type ReasoningContent } from '@/spec/modelpreset';
import type { ToolCall, ToolChoice } from '@/spec/tool';

import { log, providerSetAPI } from '@/apis/baseapi';

/**
 * @public
 */
export function getQuotedJSON(obj: any): string {
	return '```json\n' + JSON.stringify(obj, null, 2) + '\n```';
}

function convertConversationToBuildMessages(conversationMessages?: ConversationMessage[]): ChatCompletionDataMessage[] {
	if (!conversationMessages) {
		return [];
	}
	const chatMessages: ChatCompletionDataMessage[] = [];
	conversationMessages.forEach(convoMsg => {
		const message: ChatCompletionDataMessage = {
			role: convoMsg.role,
			name: convoMsg.name,
			content: convoMsg.content,

			reasoningContents: convoMsg.reasoningContents,
			// Attachments are stored per conversation message; pass them through.
			attachments: convoMsg.attachments,
			// Plumb tool calls and tool outputs through to the provider layer so
			// it can construct the appropriate messages (e.g. tool/tool_result).
			toolCalls: convoMsg.toolCalls,
			toolOutputs: convoMsg.toolOutputs,
		};

		chatMessages.push(message);
	});
	return chatMessages;
}

export function getCompletionDataFromConversation(messages?: Array<ConversationMessage>) {
	const allMessages = convertConversationToBuildMessages(messages);
	const promptMsg = allMessages.pop();
	const prevMessages = allMessages;
	if (!promptMsg) {
		throw Error('Invalid prompt message');
	}
	if (promptMsg.content === '' && !promptMsg.attachments && !promptMsg.toolOutputs) {
		throw Error('Invalid prompt message input. No content or attachments or tool outputs');
	}
	const promptConvoMsg = messages && messages.length > 0 ? messages[messages.length - 1] : undefined;
	const toolChoices = promptConvoMsg?.toolChoices;
	return { promptMsg, prevMessages, toolChoices };
}

function parseAPIResponse(
	convoMessage: ConversationMessage,
	providerResp: FetchCompletionResponseBody | undefined
): {
	responseMessage: ConversationMessage;
	requestDetails: string | undefined;
} {
	let responseContent = '';
	let respDetails: string | undefined;
	let requestDetails: string | undefined;
	let usage: CompletionUsage | undefined;
	let toolCalls: ToolCall[] | undefined;
	let reasoningContents: ReasoningContent[] | undefined;

	if (providerResp) {
		responseContent = providerResp.content ?? '';

		if (providerResp.responseDetails) {
			// Make it clear in the UI what this JSON block represents.
			respDetails = '### Response Details\n' + getQuotedJSON(providerResp.responseDetails);
		}
		if (providerResp.requestDetails) {
			requestDetails = getQuotedJSON(providerResp.requestDetails);
		}
		if (providerResp.errorDetails) {
			respDetails = providerResp.errorDetails.message;
			if (providerResp.errorDetails.responseDetails) {
				respDetails += '\n\n### Response Details\n' + getQuotedJSON(providerResp.errorDetails.responseDetails);
			}
			if (providerResp.errorDetails.requestDetails) {
				const r = providerResp.errorDetails.requestDetails;
				requestDetails = getQuotedJSON(r);
			}
			if (!responseContent) {
				responseContent = convoMessage.content || '';
			}
			responseContent += '\n\n>Got error in API processing. Check details below.';
		}

		if (providerResp.reasoningContents && providerResp.reasoningContents.length > 0) {
			reasoningContents = providerResp.reasoningContents;
		}

		// Capture tool calls both for debug-details and structured usage.
		if (providerResp.toolCalls && providerResp.toolCalls.length > 0) {
			toolCalls = providerResp.toolCalls;
		}

		if (providerResp.usage) {
			usage = providerResp.usage;
		}
	}

	if (!responseContent) {
		responseContent = convoMessage.content || '';
	}

	convoMessage.content = responseContent;
	convoMessage.details = respDetails;
	convoMessage.reasoningContents = reasoningContents;
	convoMessage.usage = usage;
	convoMessage.toolCalls = toolCalls;

	return { responseMessage: convoMessage, requestDetails: requestDetails };
}

async function handleDirectCompletion(
	convoMessage: ConversationMessage,
	provider: ProviderName,
	modelParams: ModelParam,
	promptMsg: ChatCompletionDataMessage,
	prevMessages: ChatCompletionDataMessage[],
	toolChoices: ToolChoice[] | undefined,
	requestId?: string,
	signal?: AbortSignal
): Promise<{
	responseMessage: ConversationMessage | undefined;
	requestDetails: string | undefined;
}> {
	const providerResp = await providerSetAPI.completion(
		provider,
		modelParams,
		promptMsg,
		prevMessages,
		toolChoices,
		requestId,
		signal
	);
	return parseAPIResponse(convoMessage, providerResp);
}

async function handleStreamedCompletion(
	convoMessage: ConversationMessage,
	provider: ProviderName,
	modelParams: ModelParam,
	promptMsg: ChatCompletionDataMessage,
	prevMessages: ChatCompletionDataMessage[],
	toolChoices: ToolChoice[] | undefined,
	requestId?: string,
	signal?: AbortSignal,
	onStreamTextData?: (textData: string) => void,
	onStreamThinkingData?: (thinkingData: string) => void
): Promise<{
	responseMessage: ConversationMessage | undefined;
	requestDetails: string | undefined;
}> {
	const providerResp = await providerSetAPI.completion(
		provider,
		modelParams,
		promptMsg,
		prevMessages,
		toolChoices,
		requestId,
		signal,
		onStreamTextData,
		onStreamThinkingData
	);
	return parseAPIResponse(convoMessage, providerResp);
}

export async function HandleCompletion(
	provider: ProviderName,
	modelParams: ModelParam,
	promptMsg: ChatCompletionDataMessage,
	prevMessages: ChatCompletionDataMessage[],
	toolChoices: ToolChoice[] | undefined,
	convoMessage: ConversationMessage,
	requestId?: string,
	signal?: AbortSignal,
	onStreamTextData?: (textData: string) => void,
	onStreamThinkingData?: (thinkingData: string) => void
): Promise<{
	responseMessage: ConversationMessage | undefined;
	requestDetails: string | undefined;
}> {
	try {
		const isStream = modelParams.stream || false;
		if (isStream && onStreamTextData && onStreamThinkingData) {
			return await handleStreamedCompletion(
				convoMessage,
				provider,
				modelParams,
				promptMsg,
				prevMessages,
				toolChoices,
				requestId,
				signal,
				onStreamTextData,
				onStreamThinkingData
			);
		} else {
			modelParams.stream = false;
			return await handleDirectCompletion(
				convoMessage,
				provider,
				modelParams,
				promptMsg,
				prevMessages,
				toolChoices,
				requestId,
				signal
			);
		}
	} catch (error) {
		if ((error as DOMException).name === 'AbortError') {
			throw error; // let the caller decide
		}

		// Log full error for debugging
		const details = JSON.stringify(error, null, 2);
		log.error('provider completion failed', details);

		// Preserve any content that is already in the message (e.g. partial stream)
		const existing = (convoMessage.content || '').trimEnd();
		const suffix = '\n\n>Got error in API processing. Check details below.';

		convoMessage.content = existing ? existing + suffix : suffix;
		convoMessage.details = details;

		return { responseMessage: convoMessage, requestDetails: undefined };
	}
}
