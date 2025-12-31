import type { Attachment } from '@/spec/attachment';
import type {
	InferenceError,
	InferenceUsage,
	InputUnion,
	ModelParam,
	OutputUnion,
	ReasoningContent,
	RoleEnum,
	Status,
	ToolChoice,
} from '@/spec/inference';
import type { ToolStoreChoice, UIToolCallChip, UIToolOutput } from '@/spec/tool';

/** Keep in sync with Go's ConversationSchemaVersion. */
export const CONVERSATION_SCHEMA_VERSION = 'v1.0.0';

export interface StoreConversationMessage {
	id: string;
	// Go type: time
	createdAt: Date;
	role: RoleEnum;
	status: Status;

	modelParam?: ModelParam;
	inputs?: InputUnion[];
	outputs?: OutputUnion[];

	toolChoices?: ToolChoice[];

	toolStoreChoices?: ToolStoreChoice[];
	attachments?: Attachment[];

	usage?: InferenceUsage;
	error?: InferenceError;

	meta?: Record<string, any>;
}

interface UIConversationMessageDetails {
	// UI-only, derived from outputs (we'll derive in helpers)
	content: string;
	details?: string;
	reasoningContents?: ReasoningContent[];
	toolCalls?: UIToolCallChip[];
	toolOutputs?: UIToolOutput[];
}

export type ConversationMessage = StoreConversationMessage & UIConversationMessageDetails;

interface BaseConversation<TMessage> {
	schemaVersion: string;
	id: string;
	title: string;
	createdAt: Date;
	modifiedAt: Date;
	messages: TMessage[];
	meta?: Record<string, any>;
}

export type StoreConversation = BaseConversation<StoreConversationMessage>;
export type Conversation = BaseConversation<ConversationMessage>;

export type ConversationSearchItem = {
	id: string;
	title: string;
	idDate: Date;
	modifiedAt: Date;
};
