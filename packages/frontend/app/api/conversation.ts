import { Conversation } from 'conversationmodel';

export async function saveConversation(conversation: Conversation): Promise<void> {
	await window.ConversationAPI.saveConversation(conversation);
}

export async function startConversation(title: string, oldConversation?: Conversation): Promise<Conversation> {
	return await window.ConversationAPI.startConversation(title, oldConversation);
}

export async function deleteConversation(id: string, title: string): Promise<void> {
	await window.ConversationAPI.deleteConversation(id, title);
}

export async function getConversation(id: string, title: string): Promise<Conversation | null> {
	return await window.ConversationAPI.getConversation(id, title);
}

export async function listConversations(
	token?: string
): Promise<{ conversations: { id: string; title: string }[]; nextToken?: string }> {
	return await window.ConversationAPI.listConversations(token);
}
