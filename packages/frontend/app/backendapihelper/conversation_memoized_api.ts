import { conversationStoreAPI } from '@/backendapibase';

import { Conversation, ConversationItem } from '@/models/conversationmodel';
import { ConversationCache } from './conversation_cache';

const conversationCache = new ConversationCache();

export const getConversation = conversationStoreAPI.getConversation;
export const addMessageToConversation = conversationStoreAPI.addMessageToConversation;

export async function saveConversation(conversation: Conversation): Promise<void> {
	await conversationStoreAPI.saveConversation(conversation);
	conversationCache.updateConversationCache(conversation.id, conversation.title, 'update');
}

export async function deleteConversation(id: string, title: string): Promise<void> {
	await conversationStoreAPI.deleteConversation(id, title);
	conversationCache.updateConversationCache(id, title, 'delete');
}

export async function listAllConversations(): Promise<ConversationItem[]> {
	await conversationCache.refreshCache();
	return conversationCache.getConversationsList();
}
