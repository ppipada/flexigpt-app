import {
	type ChatCompletionDataMessage,
	type FetchCompletionData,
	type FetchCompletionResponseBody,
	type ModelParams,
	ResponseContentType,
} from '@/spec/aiprovider';
import type { ConversationMessage } from '@/spec/conversation';
import type { CompletionUsage, ProviderName } from '@/spec/modelpreset';

import { CustomMDLanguage } from '@/lib/text_utils';

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
			content: convoMsg.content,
			name: convoMsg.name,
			// Attachments are stored per conversation message; pass them through.
			attachments: convoMsg.attachments,
		};

		chatMessages.push(message);
	});
	return chatMessages;
}

export async function BuildCompletionDataFromConversation(
	provider: ProviderName,
	modelParams: ModelParams,
	messages?: Array<ConversationMessage>
): Promise<FetchCompletionData> {
	const allMessages = convertConversationToBuildMessages(messages);
	// console.log(JSON.stringify(allMessages, null, 2));
	const promptMsg = allMessages.pop();
	if (!promptMsg || promptMsg.content === '') {
		throw Error('Invalid prompt message input');
	}
	const promptConvoMsg = messages && messages.length > 0 ? messages[messages.length - 1] : undefined;
	const toolChoices = promptConvoMsg?.toolChoices;

	const completionData = providerSetAPI.buildCompletionData(provider, modelParams, promptMsg, allMessages, toolChoices);
	return completionData;
}

// export const normalizeThinkingChunk = (s: string) => s.replace(/~~~+/g, '~~\u200b~'); // break accidental fences

function parseAPIResponse(
	convoMessage: ConversationMessage,
	providerResp: FetchCompletionResponseBody | undefined
): {
	responseMessage: ConversationMessage;
	requestDetails: string | undefined;
	usage?: CompletionUsage;
} {
	let respContent = '';
	let respDetails: string | undefined;
	let requestDetails: string | undefined;
	let usage: CompletionUsage | undefined;

	if (providerResp) {
		if (providerResp.responseContent && providerResp.responseContent.length) {
			type ContentChunk = { rtype: ResponseContentType; content: string };

			// 1) Collapse adjacent same-type blocks
			const collapsed: ContentChunk[] = [];
			let last: ContentChunk | undefined;
			for (const c of providerResp.responseContent) {
				const chunk: ContentChunk = { rtype: c.type, content: c.content };
				if (last && last.rtype === chunk.rtype) {
					// Merge with a newline separator if there's already content
					last.content += (last.content ? '\n' : '') + chunk.content;
				} else {
					// Copy to avoid mutating original
					collapsed.push({ rtype: chunk.rtype, content: chunk.content });
				}
				last = chunk;
			}

			// 2) Fence the Thinking blocks and 3) concat result
			let respFullText = '';
			for (const chunk of collapsed) {
				if (chunk.rtype === ResponseContentType.ThinkingSummary) {
					respFullText += `\n~~~${CustomMDLanguage.ThinkingSummary}\n${chunk.content}\n~~~\n`;
				} else if (chunk.rtype === ResponseContentType.Thinking) {
					respFullText += `\n~~~${CustomMDLanguage.Thinking}\n${chunk.content}\n~~~\n`;
				} else {
					respFullText += `\n${chunk.content}\n`;
				}
			}

			respContent += respFullText;
		}

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
			if (!respContent) {
				respContent = convoMessage.content || '';
			}
			respContent += '\n\n>Got error in API processing. Check details below.';
		}

		// Tool calls are not directly rendered in the main message bubble, so
		// surface them in the debug/details panel.
		if (providerResp.toolCalls && providerResp.toolCalls.length > 0) {
			const toolCallsBlock = '### Tool calls\n' + getQuotedJSON(providerResp.toolCalls);
			respDetails = respDetails ? `${respDetails}\n\n${toolCallsBlock}` : toolCallsBlock;
		}

		if (providerResp.usage) {
			usage = providerResp.usage;
		}
	}

	if (!respContent) {
		respContent = convoMessage.content || '';
	}

	convoMessage.content = respContent;
	convoMessage.details = respDetails;
	return { responseMessage: convoMessage, requestDetails: requestDetails, usage: usage };
}

async function handleDirectCompletion(
	convoMessage: ConversationMessage,
	provider: ProviderName,
	completionData: FetchCompletionData,
	requestId?: string,
	signal?: AbortSignal
): Promise<{
	responseMessage: ConversationMessage | undefined;
	requestDetails: string | undefined;
	usage?: CompletionUsage;
}> {
	const providerResp = await providerSetAPI.completion(provider, completionData, requestId, signal);
	return parseAPIResponse(convoMessage, providerResp);
}

async function handleStreamedCompletion(
	convoMessage: ConversationMessage,
	provider: ProviderName,
	completionData: FetchCompletionData,
	requestId?: string,
	signal?: AbortSignal,
	onStreamTextData?: (textData: string) => void,
	onStreamThinkingData?: (thinkingData: string) => void
): Promise<{
	responseMessage: ConversationMessage | undefined;
	requestDetails: string | undefined;
	usage?: CompletionUsage;
}> {
	const providerResp = await providerSetAPI.completion(
		provider,
		completionData,
		requestId,
		signal,
		onStreamTextData,
		onStreamThinkingData
	);
	return parseAPIResponse(convoMessage, providerResp);
}

export async function HandleCompletion(
	provider: ProviderName,
	completionData: FetchCompletionData,
	convoMessage: ConversationMessage,
	requestId?: string,
	signal?: AbortSignal,
	onStreamTextData?: (textData: string) => void,
	onStreamThinkingData?: (thinkingData: string) => void
): Promise<{
	responseMessage: ConversationMessage | undefined;
	requestDetails: string | undefined;
	usage?: CompletionUsage;
}> {
	try {
		const isStream = completionData.modelParams.stream || false;
		// log.info('FetchCompletionData', defaultProvider, JSON.stringify(fullCompletionData, null, 2));
		if (isStream && onStreamTextData && onStreamThinkingData) {
			return await handleStreamedCompletion(
				convoMessage,
				provider,
				completionData,
				requestId,
				signal,
				onStreamTextData,
				onStreamThinkingData
			);
		} else {
			completionData.modelParams.stream = false;
			return await handleDirectCompletion(convoMessage, provider, completionData, requestId, signal);
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

		return { responseMessage: convoMessage, requestDetails: undefined, usage: undefined };
	}
}
