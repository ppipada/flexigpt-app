import { ChatCompletionRequestMessage, ChatCompletionRoleEnum, providerSet } from 'aiprovider';
import { ConversationMessage, ConversationRoleEnum, initConversationMessage } from 'conversationmodel';

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

	const newMsg = initConversationMessage(ConversationRoleEnum.assistant, respContent);
	newMsg.details = respDetails;
	return newMsg;
}

async function handleStreamedCompletion(
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

	const newMsg = initConversationMessage(ConversationRoleEnum.assistant, respContent);
	newMsg.details = respDetails;
	return newMsg;
}

export async function getCompletionMessage(
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
		return handleStreamedCompletion(providerAPI, fullCompletionRequest, onStreamData);
	} else {
		return handleDirectCompletion(providerAPI, fullCompletionRequest);
	}
}
