import { v7 as uuidv7 } from 'uuid';
import { Conversation, ConversationMessage, ConversationRoleEnum } from './conversation_types';

export function initConversation(title: string = 'New Conversation'): Conversation {
	return {
		id: uuidv7(),
		title: title.substring(0, 64),
		createdAt: new Date(),
		modifiedAt: new Date(),
		messages: [],
	};
}

export function initConversationMessage(role: ConversationRoleEnum, content: string): ConversationMessage {
	const d = new Date();
	return {
		id: d.toISOString(),
		createdAt: new Date(),
		role: role,
		content: content,
	};
}
