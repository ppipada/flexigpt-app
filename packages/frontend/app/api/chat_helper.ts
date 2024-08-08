/* eslint-disable @typescript-eslint/no-unused-vars */
import { completion, getCompletionRequest, getDefaultProvider, getProviderInfo } from '@/api/base_aiproviderimpl';
import {
	ChatCompletionRequestMessage,
	ChatCompletionRoleEnum,
	CompletionResponse,
	ProviderName,
} from 'aiprovidermodel';
import { ConversationMessage, ConversationRoleEnum } from 'conversationmodel';
import { log } from 'logger';

const roleMap: Record<ConversationRoleEnum, ChatCompletionRoleEnum> = {
	[ConversationRoleEnum.system]: ChatCompletionRoleEnum.system,
	[ConversationRoleEnum.user]: ChatCompletionRoleEnum.user,
	[ConversationRoleEnum.assistant]: ChatCompletionRoleEnum.assistant,
	[ConversationRoleEnum.function]: ChatCompletionRoleEnum.function,
	[ConversationRoleEnum.feedback]: ChatCompletionRoleEnum.user,
};

function convertConversationToChatMessages(
	conversationMessages?: ConversationMessage[]
): ChatCompletionRequestMessage[] {
	if (!conversationMessages) {
		return [];
	}
	const chatMessages: ChatCompletionRequestMessage[] = [];
	conversationMessages.forEach(convoMsg => {
		chatMessages.push({ role: roleMap[convoMsg.role], content: convoMsg.content });
	});
	return chatMessages;
}

function parseAPIResponse(convoMessage: ConversationMessage, providerResp: CompletionResponse | undefined) {
	let respContent = '';
	let respDetails: string | undefined;
	let requestDetails: string | undefined;
	if (providerResp) {
		if (providerResp.respContent) {
			respContent = providerResp.respContent;
		}
		if (providerResp.responseDetails) {
			respDetails = '```json' + JSON.stringify(providerResp.responseDetails, null, 2) + '```';
		}
		if (providerResp.errorDetails) {
			const errorDetails = '```json' + JSON.stringify(providerResp.errorDetails, null, 2) + '```';
			if (respDetails) {
				respDetails += '\n' + errorDetails;
			} else {
				respDetails = errorDetails;
			}
		}
		if (providerResp.requestDetails) {
			requestDetails = '```json' + JSON.stringify(providerResp.requestDetails, null, 2) + '```';
		}
	}

	convoMessage.content = respContent;
	convoMessage.details = respDetails;
	return { responseMessage: convoMessage, requestDetails: requestDetails };
}

async function handleDirectCompletion(
	convoMessage: ConversationMessage,
	provider: ProviderName,
	fullCompletionRequest: any
): Promise<{ responseMessage: ConversationMessage | undefined; requestDetails: string | undefined }> {
	const providerResp = await completion(provider, fullCompletionRequest);
	return parseAPIResponse(convoMessage, providerResp);
}

async function handleStreamedCompletion(
	convoMessage: ConversationMessage,
	provider: ProviderName,
	fullCompletionRequest: any,
	onStreamData: (data: string) => void
): Promise<{ responseMessage: ConversationMessage | undefined; requestDetails: string | undefined }> {
	const dataFunction = async (data: any): Promise<void> => {
		return new Promise((resolve, reject) => {
			try {
				onStreamData(data);
				resolve();
			} catch (error) {
				reject(error);
			}
		});
	};

	const providerResp = await completion(provider, fullCompletionRequest, dataFunction);
	return parseAPIResponse(convoMessage, providerResp);
}

export async function getCompletionMessage(
	convoMessage: ConversationMessage,
	messages?: Array<ConversationMessage>,
	inputParams?: { [key: string]: any },
	onStreamData?: (data: string) => void
): Promise<{ responseMessage: ConversationMessage | undefined; requestDetails: string | undefined }> {
	const allMessages = convertConversationToChatMessages(messages);
	const promptMsg = allMessages.pop();
	let defaultProvider: ProviderName;
	try {
		defaultProvider = await getDefaultProvider();
	} catch (e) {
		log.error(e);
		throw e;
	}
	const providerInfo = await getProviderInfo(defaultProvider);
	const isStream = providerInfo.streamingSupport || false;
	const fullCompletionRequest = await getCompletionRequest(
		defaultProvider,
		promptMsg?.content || '',
		allMessages,
		inputParams,
		isStream
	);
	// log.info('CompletionRequest', defaultProvider, JSON.stringify(fullCompletionRequest, null, 2));
	if (isStream && onStreamData) {
		return await handleStreamedCompletion(convoMessage, defaultProvider, fullCompletionRequest, onStreamData);
	} else {
		return await handleDirectCompletion(convoMessage, defaultProvider, fullCompletionRequest);
	}
}
