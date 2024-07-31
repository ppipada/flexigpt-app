import { ChatAPIMessage, ChatCompletionRoleEnum } from 'aiprovider';

export interface ConversationMessage extends ChatAPIMessage {
	details?: string;
}

export interface Conversation {
	id: string;
	title: string;
	createdAt: Date;
	modifiedAt: Date;
	messages: ConversationMessage[];
}

export interface User {
	id: string;
	role: ChatCompletionRoleEnum;
}
