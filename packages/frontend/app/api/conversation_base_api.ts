import { Conversation, ConversationMessage } from 'conversationmodel';

export async function saveConversation(conversation: Conversation): Promise<void> {
	await window.ConversationAPI.saveConversation(conversation);
}

export async function createNewConversation(title: string): Promise<Conversation> {
	return await window.ConversationAPI.createNewConversation(title);
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

export async function addMessageToConversation(
	id: string,
	title: string,
	newMessage: ConversationMessage
): Promise<void> {
	await window.ConversationAPI.addMessageToConversation(id, title, newMessage);
}

export async function listAllConversations(): Promise<{ id: string; title: string }[]> {
	let allConversations: { id: string; title: string }[] = [];
	let token: string | undefined = undefined;

	do {
		const response = await listConversations(token);
		allConversations = allConversations.concat(response.conversations);
		token = response.nextToken;
	} while (token);

	return allConversations;
}
