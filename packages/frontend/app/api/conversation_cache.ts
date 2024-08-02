import { listConversations as apiListConversations } from './conversation_base_api';

interface CacheData {
	conversationsDict: { [id: string]: { id: string; title: string } };
	conversationsList: { id: string; title: string }[];
	timestamp: number;
}

export class ConversationCache {
	private cache: CacheData | null = null;
	private static readonly CACHE_INVALIDATION_TIME = 60 * 60 * 1000; // 1 hour in milliseconds

	private isCacheValid(): boolean {
		if (!this.cache) return false;
		return Date.now() - this.cache.timestamp < ConversationCache.CACHE_INVALIDATION_TIME;
	}

	async refreshCache(force: boolean = false): Promise<void> {
		if (!force && this.isCacheValid()) {
			return;
		}

		const conversations = await this.listAllConversationsFromAPI();
		const conversationsDict: { [id: string]: { id: string; title: string } } = {};
		conversations.forEach(conv => {
			conversationsDict[conv.id] = conv;
		});

		this.cache = {
			conversationsDict,
			conversationsList: conversations,
			timestamp: Date.now(),
		};
	}

	private async listAllConversationsFromAPI(): Promise<{ id: string; title: string }[]> {
		let allConversations: { id: string; title: string }[] = [];
		let token: string | undefined = undefined;

		do {
			const response = await apiListConversations(token);
			allConversations = allConversations.concat(response.conversations);
			token = response.nextToken;
		} while (token);

		return allConversations;
	}

	updateConversationCache(id: string, title: string, type: 'add' | 'delete' | 'update'): void {
		if (!this.isCacheValid()) {
			return;
		}

		if (type === 'delete') {
			delete this.cache!.conversationsDict[id];
			this.cache!.conversationsList = this.cache!.conversationsList.filter(c => c.id !== id);
		} else if (type === 'add' || type === 'update') {
			const cachedConversation = this.cache!.conversationsDict[id];
			if (cachedConversation) {
				if (cachedConversation.title !== title) {
					cachedConversation.title = title;
					const index = this.cache!.conversationsList.findIndex(c => c.id === id);
					if (index >= 0) {
						this.cache!.conversationsList[index].title = title;
					}
				}
			} else {
				this.cache!.conversationsDict[id] = { id, title };
				this.cache!.conversationsList.push({ id, title });
			}
		}
	}

	getConversationsList(): { id: string; title: string }[] {
		if (this.isCacheValid() && this.cache) {
			return this.cache.conversationsList;
		}
		return [];
	}
}
