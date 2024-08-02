import { ChatAPIMessage, ChatCompletionRoleEnum } from 'aiprovider';

export interface ConversationMessage extends ChatAPIMessage {
	details?: string;
	userId: string;
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
	name: string;
	role: ChatCompletionRoleEnum;
}

export interface IConversationAPI {
	createNewConversation: (title: string) => Promise<Conversation>;
	saveConversation: (conversation: Conversation) => Promise<void>;
	deleteConversation: (id: string, title: string) => Promise<void>;
	getConversation: (id: string, title: string) => Promise<Conversation | null>;
	listConversations: (
		token?: string
	) => Promise<{ conversations: { id: string; title: string }[]; nextToken?: string }>;
	addMessageToConversation(id: string, title: string, newMessage: ConversationMessage): Promise<void>;
}
