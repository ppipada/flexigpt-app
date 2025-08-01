import {
	type ChatCompletionRequestMessage,
	ChatCompletionRoleEnum,
	type CompletionResponse,
	type ModelParams,
} from '@/spec/aiprovider';
import type { ConversationMessage } from '@/spec/conversation';
import { ConversationRoleEnum } from '@/spec/conversation';
import type { ProviderName } from '@/spec/modelpreset';

import { log, providerSetAPI } from '@/apis/baseapi';

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
	prompt: string,
	modelParams: ModelParams,
	prevMessages: Array<ChatCompletionRequestMessage>
): Promise<{ responseMessage: ConversationMessage | undefined; requestDetails: string | undefined }> {
	const providerResp = await providerSetAPI.completion(provider, prompt, modelParams, prevMessages);
	return parseAPIResponse(convoMessage, providerResp);
}

async function handleStreamedCompletion(
	convoMessage: ConversationMessage,
	provider: ProviderName,
	prompt: string,
	modelParams: ModelParams,
	prevMessages: Array<ChatCompletionRequestMessage>,
	onStreamData: (data: string) => void
): Promise<{ responseMessage: ConversationMessage | undefined; requestDetails: string | undefined }> {
	const providerResp = await providerSetAPI.completion(provider, prompt, modelParams, prevMessages, onStreamData);
	return parseAPIResponse(convoMessage, providerResp);
}

export async function GetCompletionMessage(
	provider: ProviderName,
	modelParams: ModelParams,
	convoMessage: ConversationMessage,
	messages?: Array<ConversationMessage>,
	onStreamData?: (data: string) => void
): Promise<{ responseMessage: ConversationMessage | undefined; requestDetails: string | undefined }> {
	try {
		// console.log(JSON.stringify(modelParams, null, 2));
		const allMessages = convertConversationToChatMessages(messages);
		// console.log(JSON.stringify(allMessages, null, 2));
		const promptMsg = allMessages.pop();

		const isStream = modelParams.stream || false;
		// log.info('CompletionRequest', defaultProvider, JSON.stringify(fullCompletionRequest, null, 2));
		if (isStream && onStreamData) {
			return await handleStreamedCompletion(
				convoMessage,
				provider,
				promptMsg?.content || '',
				modelParams,
				allMessages,
				onStreamData
			);
		} else {
			modelParams.stream = false;
			return await handleDirectCompletion(convoMessage, provider, promptMsg?.content || '', modelParams, allMessages);
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
