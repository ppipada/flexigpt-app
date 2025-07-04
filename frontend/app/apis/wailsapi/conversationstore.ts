import type {
	Conversation,
	ConversationListItem,
	ConversationMessage,
	IConversationStoreAPI,
} from '@/models/conversationmodel';

import {
	DeleteConversation,
	GetConversation,
	ListConversations,
	PutConversation,
	PutMessagesToConversation,
	SearchConversations,
} from '@/apis/wailsjs/go/main/ConversationCollectionWrapper';
import type { spec as wailsSpec } from '@/apis/wailsjs/go/models';

/**
 * @public
 */
export class WailsConversationStoreAPI implements IConversationStoreAPI {
	async putConversation(conversation: Conversation): Promise<void> {
		const req = {
			ID: conversation.id,
			Body: {
				title: conversation.title,
				createdAt: conversation.createdAt,
				modifiedAt: conversation.modifiedAt,
				messages: conversation.messages as wailsSpec.ConversationMessage[],
			} as wailsSpec.PutConversationRequestBody,
		};

		await PutConversation(req as wailsSpec.PutConversationRequest);
	}

	async putMessagesToConversation(id: string, title: string, messages: ConversationMessage[]): Promise<void> {
		const req = {
			ID: id,
			Body: {
				title: title,
				messages: messages as wailsSpec.ConversationMessage[],
			} as wailsSpec.PutMessagesToConversationRequestBody,
		};

		await PutMessagesToConversation(req as wailsSpec.PutMessagesToConversationRequest);
	}

	async deleteConversation(id: string, title: string): Promise<void> {
		const req = { ID: id, Title: title };
		await DeleteConversation(req as wailsSpec.DeleteConversationRequest);
	}

	async getConversation(id: string, title: string): Promise<Conversation | null> {
		const req = { ID: id, Title: title };
		const c = await GetConversation(req as wailsSpec.GetConversationRequest);
		return c.Body as Conversation;
	}

	async listConversations(token?: string): Promise<{ conversations: ConversationListItem[]; nextToken?: string }> {
		const req = { PageToken: token || '' };
		const resp = await ListConversations(req as wailsSpec.ListConversationsRequest);
		return {
			conversations: resp.Body?.conversationListItems as ConversationListItem[],
			nextToken: resp.Body?.nextPageToken,
		};
	}

	async searchConversations(
		query: string,
		token?: string,
		pageSize?: number
	): Promise<{ conversations: ConversationListItem[]; nextToken?: string }> {
		const req = { Query: query, PageToken: token || '', PageSize: pageSize || 10 };
		const resp = await SearchConversations(req as wailsSpec.SearchConversationsRequest);
		return {
			conversations: resp.Body?.conversationListItems as ConversationListItem[],
			nextToken: resp.Body?.nextPageToken,
		};
	}
}
