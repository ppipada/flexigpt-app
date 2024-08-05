import { ChatCompletionRequestMessage, ChatCompletionRoleEnum, providerSet } from 'aiprovider';
import { ConversationMessage, ConversationRoleEnum } from 'conversationmodel';

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

async function handleDirectCompletion(
	convoMessage: ConversationMessage,
	providerAPI: any,
	fullCompletionRequest: any
): Promise<ConversationMessage | undefined> {
	const providerResp = await providerAPI.completion(fullCompletionRequest);
	let respContent = '';
	let respDetails: string | undefined;

	if (providerResp) {
		if (providerResp.respContent) {
			respContent = providerResp.respContent;
		}
		if (providerResp.fullResponse) {
			respDetails = '```json' + JSON.stringify(providerResp.fullResponse, null, 2) + '```';
		}
	}

	convoMessage.content = respContent;
	convoMessage.details = respDetails;
	return convoMessage;
}

async function handleStreamedCompletion(
	convoMessage: ConversationMessage,
	providerAPI: any,
	fullCompletionRequest: any,
	onStreamData: (data: string) => void
): Promise<ConversationMessage | undefined> {
	const providerResp = await providerAPI.completion(fullCompletionRequest, onStreamData);
	let respContent = '';
	let respDetails: string | undefined;

	if (providerResp) {
		if (providerResp.respContent) {
			respContent = providerResp.respContent;
		}
		if (providerResp.fullResponse) {
			respDetails = '```json' + JSON.stringify(providerResp.fullResponse, null, 2) + '```';
		}
	}
	convoMessage.content = respContent;
	convoMessage.details = respDetails;
	return convoMessage;
}

export async function getCompletionMessage(
	convoMessage: ConversationMessage,
	prompt: string,
	messages?: Array<ConversationMessage>,
	inputParams?: { [key: string]: any },
	onStreamData?: (data: string) => void
): Promise<ConversationMessage | undefined> {
	const chatMsgs = convertConversationToChatMessages(messages);
	const providerAPI = providerSet.getProviderAPI(providerSet.getDefaultProvider());
	const isStream = providerAPI.getProviderInfo().streamingSupport || false;
	const fullCompletionRequest = providerAPI.getCompletionRequest(prompt, chatMsgs, inputParams, isStream);

	if (isStream && onStreamData) {
		return handleStreamedCompletion(convoMessage, providerAPI, fullCompletionRequest, onStreamData);
	} else {
		return handleDirectCompletion(convoMessage, providerAPI, fullCompletionRequest);
	}
}
