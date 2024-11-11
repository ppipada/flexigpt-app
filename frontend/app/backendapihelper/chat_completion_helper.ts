/* eslint-disable @typescript-eslint/no-unused-vars */
import { log, providerSetAPI } from '@/backendapibase';
import {
	ChatCompletionRequestMessage,
	ChatCompletionRoleEnum,
	CompletionResponse,
	ProviderName,
} from '@/models/aiprovidermodel';
import { ConversationMessage, ConversationRoleEnum } from '@/models/conversationmodel';

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
	const providerResp = await providerSetAPI.completion(provider, fullCompletionRequest);
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

	const providerResp = await providerSetAPI.completion(provider, fullCompletionRequest, dataFunction);
	return parseAPIResponse(convoMessage, providerResp);
}

export async function GetCompletionMessage(
	convoMessage: ConversationMessage,
	messages?: Array<ConversationMessage>,
	inputParams?: { [key: string]: any },
	onStreamData?: (data: string) => void
): Promise<{ responseMessage: ConversationMessage | undefined; requestDetails: string | undefined }> {
	try {
		const allMessages = convertConversationToChatMessages(messages);
		const promptMsg = allMessages.pop();

		let completionProvider = await providerSetAPI.getDefaultProvider();
		if (inputParams && 'provider' in inputParams) {
			const providerStr = inputParams['provider'] as string;
			if (Object.values(ProviderName).includes(providerStr as ProviderName)) {
				completionProvider = providerStr as ProviderName;
			}
			delete inputParams['provider'];
		}

		const fullCompletionRequest = await providerSetAPI.getCompletionRequest(
			completionProvider,
			promptMsg?.content || '',
			allMessages,
			inputParams
		);
		const isStream = fullCompletionRequest.stream;
		// log.info('CompletionRequest', defaultProvider, JSON.stringify(fullCompletionRequest, null, 2));
		if (isStream && onStreamData) {
			return await handleStreamedCompletion(convoMessage, completionProvider, fullCompletionRequest, onStreamData);
		} else {
			fullCompletionRequest.stream = false;
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
