import type { InferenceError, InputOutputContent, ModelParam, Usage } from '@/spec/inference';
import type { ReasoningContent } from '@/spec/modelpreset';
import type { ToolCall, ToolChoice, ToolOutput } from '@/spec/tool';

export interface ConversationMessage {
	id: string;
	createdAt: Date;
	role: string;
	status: string;
	modelParam?: ModelParam;
	messages?: InputOutputContent[];
	reasoning?: ReasoningContent[];
	toolCalls?: ToolCall[];
	toolOutputs?: ToolOutput[];
	toolChoices?: ToolChoice[];
	usage?: Usage;
	error?: InferenceError;
	meta?: Record<string, any>;
}

export type ConversationSearchItem = {
	id: string;
	title: string;
	idDate: Date;
	modifiedAt: Date;
};

export type Conversation = {
	schemaVersion: string;
	id: string;
	title?: string;
	// Go type: time
	createdAt: Date;
	// Go type: time
	modifiedAt: Date;
	messages: ConversationMessage[];
	meta?: Record<string, any>;
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
