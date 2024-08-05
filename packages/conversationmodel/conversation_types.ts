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

export interface ConversationItem {
	id: string;
	title: string;
	createdAt: Date;
}

export type Conversation = ConversationItem & {
	modifiedAt: Date;
	messages: ConversationMessage[];
};

export interface IConversationAPI {
	saveConversation: (conversation: Conversation) => Promise<void>;
	deleteConversation: (id: string, title: string) => Promise<void>;
	getConversation: (id: string, title: string) => Promise<Conversation | null>;
	listConversations: (token?: string) => Promise<{ conversations: ConversationItem[]; nextToken?: string }>;
	addMessageToConversation(id: string, title: string, newMessage: ConversationMessage): Promise<void>;
}
