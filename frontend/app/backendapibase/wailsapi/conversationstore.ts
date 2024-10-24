import {
	AddMessageToConversation,
	DeleteConversation,
	GetConversation,
	ListConversations,
	SaveConversation,
} from '@/backendapibase/wailsjs/go/conversationstore/ConversationCollection';
import { spec as wailsSpec } from '@/backendapibase/wailsjs/go/models';
import { Conversation, ConversationItem, ConversationMessage, IConversationStoreAPI } from '@/models/conversationmodel';

export class WailsConversationStoreAPI implements IConversationStoreAPI {
	async saveConversation(conversation: Conversation): Promise<void> {
		await SaveConversation(conversation as wailsSpec.Conversation);
	}

	async deleteConversation(id: string, title: string): Promise<void> {
		await DeleteConversation(id, title);
	}

	async getConversation(id: string, title: string): Promise<Conversation | null> {
		const c = await GetConversation(id, title);
		return c as Conversation;
	}

	async listConversations(token?: string): Promise<{ conversations: ConversationItem[]; nextToken?: string }> {
		const resp = await ListConversations(token || '');
		return { conversations: resp.ConversationItems as ConversationItem[], nextToken: resp.NextPageToken };
	}

	async addMessageToConversation(id: string, title: string, newMessage: ConversationMessage): Promise<void> {
		await AddMessageToConversation(id, title, newMessage as wailsSpec.ConversationMessage);
	}
}
