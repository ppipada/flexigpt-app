import type { ConversationItem } from '@/models/conversationmodel';

import { conversationStoreAPI } from '@/apis/baseapi';

export async function listAllConversations(): Promise<ConversationItem[]> {
	let allConversations: ConversationItem[] = [];
	let token: string | undefined = undefined;

	do {
		const response = await conversationStoreAPI.listConversations(token);
		allConversations = allConversations.concat(response.conversations);
		token = response.nextToken;
	} while (token);

	return allConversations;
}
