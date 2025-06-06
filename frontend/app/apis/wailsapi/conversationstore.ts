import type {
	Conversation,
	ConversationItem,
	ConversationMessage,
	IConversationStoreAPI,
} from '@/models/conversationmodel';

import {
	AddMessageToConversation,
	DeleteConversation,
	GetConversation,
	ListConversations,
	SaveConversation,
	SearchConversations,
} from '@/apis/wailsjs/go/main/ConversationCollectionWrapper';
import type { spec as wailsSpec } from '@/apis/wailsjs/go/models';

/**
 * @public
 */
export class WailsConversationStoreAPI implements IConversationStoreAPI {
	async saveConversation(conversation: Conversation): Promise<void> {
		const req = { Body: conversation };
		await SaveConversation(req as wailsSpec.SaveConversationRequest);
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

	async listConversations(token?: string): Promise<{ conversations: ConversationItem[]; nextToken?: string }> {
		const req = { Token: token || '' };
		const resp = await ListConversations(req as wailsSpec.ListConversationsRequest);
		return { conversations: resp.Body?.conversationItems as ConversationItem[], nextToken: resp.Body?.nextPageToken };
	}

	async searchConversations(
		query: string,
		token?: string,
		pageSize?: number
	): Promise<{ conversations: ConversationItem[]; nextToken?: string }> {
		const req = { Query: query, Token: token || '', PageSize: pageSize || 10 };
		const resp = await SearchConversations(req as wailsSpec.SearchConversationsRequest);
		return { conversations: resp.Body?.conversationItems as ConversationItem[], nextToken: resp.Body?.nextPageToken };
	}

	async addMessageToConversation(id: string, title: string, newMessage: ConversationMessage): Promise<void> {
		const req = {
			ID: id,
			Body: {
				title: title,
				newMessage: newMessage,
			},
		};
		await AddMessageToConversation(req as wailsSpec.AddMessageToConversationRequest);
	}
}
