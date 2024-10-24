import { Conversation, ConversationItem, ConversationMessage, IConversationStoreAPI } from '@/models/conversationmodel';

export class ElectronConversationStoreAPI implements IConversationStoreAPI {
	// Implement the saveConversation method
	async saveConversation(conversation: Conversation): Promise<void> {
		await window.ConversationStoreAPI.saveConversation(conversation);
	}

	// Implement the deleteConversation method
	async deleteConversation(id: string, title: string): Promise<void> {
		await window.ConversationStoreAPI.deleteConversation(id, title);
	}

	// Implement the getConversation method
	async getConversation(id: string, title: string): Promise<Conversation | null> {
		return await window.ConversationStoreAPI.getConversation(id, title);
	}

	// Implement the listConversations method
	async listConversations(token?: string): Promise<{ conversations: ConversationItem[]; nextToken?: string }> {
		return await window.ConversationStoreAPI.listConversations(token);
	}

	// Implement the addMessageToConversation method
	async addMessageToConversation(id: string, title: string, newMessage: ConversationMessage): Promise<void> {
		await window.ConversationStoreAPI.addMessageToConversation(id, title, newMessage);
	}
}
