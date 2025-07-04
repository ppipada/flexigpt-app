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
	name?: string;
	details?: string;
}

export interface ConversationListItem {
	id: string;
	sanatizedTitle: string;
}

export type ConversationSearchItem = {
	id: string;
	title: string;
	idDate: Date;
};

export type Conversation = {
	id: string;
	title: string;
	createdAt: Date;
	modifiedAt: Date;
	messages: ConversationMessage[];
};

export interface IConversationStoreAPI {
	putConversation: (conversation: Conversation) => Promise<void>;
	putMessagesToConversation(id: string, title: string, messages: ConversationMessage[]): Promise<void>;
	deleteConversation: (id: string, title: string) => Promise<void>;
	getConversation: (id: string, title: string) => Promise<Conversation | null>;
	listConversations: (token?: string) => Promise<{ conversations: ConversationListItem[]; nextToken?: string }>;
	searchConversations: (
		query: string,
		token?: string,
		pageSize?: number
	) => Promise<{ conversations: ConversationListItem[]; nextToken?: string }>;
}
