import { Conversation } from 'conversationmodel';
import {
	addMessageToConversation as apiAddMessageToConversation,
	createNewConversation as apiCreateNewConversation,
	deleteConversation as apiDeleteConversation,
	getConversation as apiGetConversation,
	saveConversation as apiSaveConversation,
} from './conversation_base_api';
import { ConversationCache } from './conversation_cache';

const conversationCache = new ConversationCache();

export const createNewConversation = apiCreateNewConversation;
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

export async function listAllConversations(): Promise<{ id: string; title: string }[]> {
	await conversationCache.refreshCache();
	return conversationCache.getConversationsList();
}
