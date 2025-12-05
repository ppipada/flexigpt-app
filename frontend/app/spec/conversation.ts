import type { Attachment } from '@/spec/attachment';
import type { CompletionUsage, RoleEnum } from '@/spec/modelpreset';
import type { ToolChoice } from '@/spec/tool';

export interface ConversationMessage {
	id: string;
	createdAt?: Date;
	role: RoleEnum;
	content: string;
	name?: string;
	details?: string;
	toolChoices?: ToolChoice[];
	attachments?: Attachment[];
	usage?: CompletionUsage;
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
