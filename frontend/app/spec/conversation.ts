export enum ConversationRoleEnum {
	system = 'system',
	user = 'user',
	assistant = 'assistant',
	function = 'function',
	feedback = 'feedback',
}

export enum ConversationAttachmentKind {
	File = 'file',
	DocIndex = 'docIndex',
	PR = 'pr',
	Commit = 'commit',
	Snapshot = 'snapshot',
}

export interface ConversationAttachment {
	kind: ConversationAttachmentKind;
	ref: string;
	label: string;
}

export interface ConversationToolChoice {
	bundleID: string;
	toolSlug: string;
	toolVersion: string;
	displayName?: string;
	description?: string;
	id?: string;
}

export interface ConversationMessage {
	id: string;
	createdAt?: Date;
	role: ConversationRoleEnum;
	content: string;
	name?: string;
	details?: string;
	toolChoices?: ConversationToolChoice[];
	attachments?: ConversationAttachment[];
	enabledTools?: string[];
}

export type ConversationSearchItem = {
	id: string;
	title: string;
	idDate: Date;
	modifiedAt: Date;
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
	getConversation: (id: string, title: string, forceFetch?: boolean) => Promise<Conversation | null>;
	listConversations: (
		token?: string,
		pageSize?: number
	) => Promise<{ conversations: ConversationSearchItem[]; nextToken?: string }>;
	searchConversations: (
		query: string,
		token?: string,
		pageSize?: number
	) => Promise<{ conversations: ConversationSearchItem[]; nextToken?: string }>;
}
