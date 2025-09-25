import {
	type ChatCompletionDataMessage,
	ChatCompletionRoleEnum,
	type CompletionData,
	type CompletionResponse,
	type ModelParams,
	ResponseContentType,
} from '@/spec/aiprovider';
import type { ConversationMessage } from '@/spec/conversation';
import { ConversationRoleEnum } from '@/spec/conversation';
import type { ProviderName } from '@/spec/modelpreset';

import { CustomMDLanguage } from '@/lib/text_utils';

import { log, providerSetAPI } from '@/apis/baseapi';

const roleMap: Record<ConversationRoleEnum, ChatCompletionRoleEnum> = {
	[ConversationRoleEnum.system]: ChatCompletionRoleEnum.system,
	[ConversationRoleEnum.user]: ChatCompletionRoleEnum.user,
	[ConversationRoleEnum.assistant]: ChatCompletionRoleEnum.assistant,
	[ConversationRoleEnum.function]: ChatCompletionRoleEnum.function,
	[ConversationRoleEnum.feedback]: ChatCompletionRoleEnum.user,
};

function convertConversationToChatMessages(conversationMessages?: ConversationMessage[]): ChatCompletionDataMessage[] {
	if (!conversationMessages) {
		return [];
	}
	const chatMessages: ChatCompletionDataMessage[] = [];
	conversationMessages.forEach(convoMsg => {
		chatMessages.push({ role: roleMap[convoMsg.role], content: convoMsg.content });
	});
	return chatMessages;
}

export function getQuotedJSON(obj: any): string {
	return '```json\n' + JSON.stringify(obj, null, 2) + '\n```';
}

// export const normalizeThinkingChunk = (s: string) => s.replace(/~~~+/g, '~~\u200b~'); // break accidental fences

function parseAPIResponse(convoMessage: ConversationMessage, providerResp: CompletionResponse | undefined) {
	let respContent = '';
	let respDetails: string | undefined;
	let requestDetails: string | undefined;

	if (providerResp) {
		if (providerResp.responseContent && providerResp.responseContent.length) {
			type ContentChunk = { type: ResponseContentType; content: string };

			// 1) Collapse adjacent same-type blocks
			const collapsed: ContentChunk[] = [];
			let last: ContentChunk | undefined;
			for (const chunk of providerResp.responseContent as ContentChunk[]) {
				if (last && last.type === chunk.type) {
					// Merge with a newline separator if there's already content
					last.content += (last.content ? '\n' : '') + chunk.content;
				} else {
					// Copy to avoid mutating original
					collapsed.push({ type: chunk.type, content: chunk.content });
				}
				last = chunk;
			}

			// 2) Fence the Thinking blocks and 3) concat result
			let respFullText = '';
			for (const chunk of collapsed) {
				if (chunk.type === ResponseContentType.ThinkingSummary) {
					respFullText += `\n~~~${CustomMDLanguage.ThinkingSummary}\n${chunk.content}\n~~~\n`;
				} else if (chunk.type === ResponseContentType.Thinking) {
					respFullText += `\n~~~${CustomMDLanguage.Thinking}\n${chunk.content}\n~~~\n`;
				} else {
					respFullText += `\n${chunk.content}\n`;
				}
			}

			respContent += respFullText;
		}

		if (providerResp.responseDetails) {
			respDetails = getQuotedJSON(providerResp.responseDetails);
		}
		if (providerResp.requestDetails) {
			requestDetails = getQuotedJSON(providerResp.requestDetails);
		}
		if (providerResp.errorDetails) {
			respDetails = providerResp.errorDetails.message;
			if (providerResp.errorDetails.responseDetails) {
				respDetails += '\n### Response Details\n' + getQuotedJSON(providerResp.errorDetails.responseDetails);
			}
			if (providerResp.errorDetails.requestDetails) {
				const r = providerResp.errorDetails.requestDetails;
				requestDetails = getQuotedJSON(r);
			}
			if (!respContent) {
				respContent = convoMessage.content || '';
			}
			respContent += '\n\n>Got error in api processing. Check details...';
		}
	}

	if (!respContent) {
		respContent = convoMessage.content || '';
	}

	convoMessage.content = respContent;
	convoMessage.details = respDetails;
	return { responseMessage: convoMessage, requestDetails };
}

async function handleDirectCompletion(
	convoMessage: ConversationMessage,
	provider: ProviderName,
	completionData: CompletionData,
	requestId?: string,
	signal?: AbortSignal
): Promise<{ responseMessage: ConversationMessage | undefined; requestDetails: string | undefined }> {
	const providerResp = await providerSetAPI.completion(provider, completionData, requestId, signal);
	return parseAPIResponse(convoMessage, providerResp);
}

async function handleStreamedCompletion(
	convoMessage: ConversationMessage,
	provider: ProviderName,
	completionData: CompletionData,
	requestId?: string,
	signal?: AbortSignal,
	onStreamTextData?: (textData: string) => void,
	onStreamThinkingData?: (thinkingData: string) => void
): Promise<{ responseMessage: ConversationMessage | undefined; requestDetails: string | undefined }> {
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

export async function BuildCompletionData(
	provider: ProviderName,
	modelParams: ModelParams,
	messages?: Array<ConversationMessage>
): Promise<CompletionData> {
	const allMessages = convertConversationToChatMessages(messages);
	// console.log(JSON.stringify(allMessages, null, 2));
	const promptMsg = allMessages.pop();
	const completionData = providerSetAPI.buildCompletionData(
		provider,
		promptMsg?.content || '',
		modelParams,
		allMessages
	);
	return completionData;
}

export async function GetCompletionMessage(
	provider: ProviderName,
	completionData: CompletionData,
	convoMessage: ConversationMessage,
	requestId?: string,
	signal?: AbortSignal,
	onStreamTextData?: (textData: string) => void,
	onStreamThinkingData?: (thinkingData: string) => void
): Promise<{ responseMessage: ConversationMessage | undefined; requestDetails: string | undefined }> {
	try {
		const isStream = completionData.modelParams.stream || false;
		// log.info('CompletionData', defaultProvider, JSON.stringify(fullCompletionData, null, 2));
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
		const msg = '\n\n>Got error in api processing. Check details...';
		const details = JSON.stringify(error, null, 2);
		log.error(msg, details);
		convoMessage.content = msg;
		convoMessage.details = details;
		return { responseMessage: convoMessage, requestDetails: undefined };
	}
}
