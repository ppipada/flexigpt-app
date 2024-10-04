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

function getQuotedJSON(obj: any): string {
	return '```json\n' + JSON.stringify(obj, null, 2) + '\n```';
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
				requestDetails = getQuotedJSON(providerResp.errorDetails.requestDetails);
			}
			respContent += '\nGot error in api processing. Check details...';
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
	try {
		const allMessages = convertConversationToChatMessages(messages);
		const promptMsg = allMessages.pop();

		let completionProvider = await getDefaultProvider();
		if (inputParams && 'provider' in inputParams) {
			const providerStr = inputParams['provider'] as string;
			if (Object.values(ProviderName).includes(providerStr as ProviderName)) {
				completionProvider = providerStr as ProviderName;
			}
			delete inputParams['provider'];
		}

		const providerInfo = await getProviderInfo(completionProvider);
		let isStream: boolean = providerInfo.streamingSupport || false;

		if (isStream && completionProvider === ProviderName.OPENAI) {
			// HACK for o models
			if (
				inputParams &&
				typeof inputParams['model'] === 'string' &&
				inputParams['model'].toLowerCase().startsWith('o')
			) {
				isStream = false;
				inputParams['temperature'] = 1;
				log.info('Detected openai o models');
			}
		}
		const fullCompletionRequest = await getCompletionRequest(
			completionProvider,
			promptMsg?.content || '',
			allMessages,
			inputParams,
			isStream
		);
		// log.info('CompletionRequest', defaultProvider, JSON.stringify(fullCompletionRequest, null, 2));
		if (isStream && onStreamData) {
			return await handleStreamedCompletion(convoMessage, completionProvider, fullCompletionRequest, onStreamData);
		} else {
			return await handleDirectCompletion(convoMessage, completionProvider, fullCompletionRequest);
		}
	} catch (error) {
		const msg = 'Got error in api processing. Check details...';
		const details = JSON.stringify(error, null, 2);
		log.error(msg, details);
		convoMessage.content = msg;
		convoMessage.details = details;
		return { responseMessage: convoMessage, requestDetails: undefined };
	}
}
