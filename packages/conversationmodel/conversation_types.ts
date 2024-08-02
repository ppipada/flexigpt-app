export enum ConversationRoleEnum {
	system = 'system',
	user = 'user',
	assistant = 'assistant',
	function = 'function',
	feedback = 'feedback',
}
export interface ConversationMessage {
	id: string;
	createdAt?: Date;
	role: ConversationRoleEnum;
	content: string;
	timestamp?: string;
	name?: string;
	details?: string;
}

export interface Conversation {
	id: string;
	title: string;
	createdAt: Date;
	modifiedAt: Date;
	messages: ConversationMessage[];
}

export interface IConversationAPI {
	saveConversation: (conversation: Conversation) => Promise<void>;
	deleteConversation: (id: string, title: string) => Promise<void>;
	getConversation: (id: string, title: string) => Promise<Conversation | null>;
	listConversations: (
		token?: string
	) => Promise<{ conversations: { id: string; title: string }[]; nextToken?: string }>;
	addMessageToConversation(id: string, title: string, newMessage: ConversationMessage): Promise<void>;
}
