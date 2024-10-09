import {
	addMessageToConversation as apiAddMessageToConversation,
	deleteConversation as apiDeleteConversation,
	getConversation as apiGetConversation,
	saveConversation as apiSaveConversation,
} from '@/backendapibase/conversation';
import { Conversation, ConversationItem } from '@/models/conversationmodel';
import { ConversationCache } from './conversation_cache';

const conversationCache = new ConversationCache();

export const getConversation = apiGetConversation;
export const addMessageToConversation = apiAddMessageToConversation;

export async function saveConversation(conversation: Conversation): Promise<void> {
	await apiSaveConversation(conversation);
	conversationCache.updateConversationCache(conversation.id, conversation.title, 'update');
}

export async function deleteConversation(id: string, title: string): Promise<void> {
	await apiDeleteConversation(id, title);
	conversationCache.updateConversationCache(id, title, 'delete');
}

export async function listAllConversations(): Promise<ConversationItem[]> {
	await conversationCache.refreshCache();
	return conversationCache.getConversationsList();
}
