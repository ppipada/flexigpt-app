import { ChatCompletionRoleEnum } from 'aiprovider';
import { v7 as uuidv7 } from 'uuid';
import { Conversation, ConversationMessage } from './conversation_types';

export function initConversation(title: string = 'New Conversation'): Conversation {
	return {
		id: uuidv7(),
		title: title.substring(0, 64),
		createdAt: new Date(),
		modifiedAt: new Date(),
		messages: [],
	};
}

export function initConversationMessage(
	role: ChatCompletionRoleEnum,
	content: string,
	userId: string
): ConversationMessage {
	return {
		id: new Date().toISOString(),
		role: role,
		content: content,
		userId: userId,
	};
}
