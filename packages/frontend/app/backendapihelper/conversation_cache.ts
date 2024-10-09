import { listConversations as apiListConversations } from '@/backendapibase/conversation';
import { ConversationItem } from '@/models/conversationmodel';

import { log } from '@/logger';
import { parse } from 'uuid';

export function getDateFromUUIDv7(uuid: string): Date {
	// Check if the UUID is valid and has the correct length
	if (!uuid || uuid.length !== 36) {
		log.error('Invalid UUIDv7 string');
		return new Date();
	}

	// Parse the UUID into bytes
	const uuidBytes = parse(uuid);

	// Prepare an array to hold the 8-byte timestamp
	const timestampBytes = new Uint8Array(8);

	// Set the first 6 bytes from the UUID bytes (first 48 bits)
	timestampBytes.set(uuidBytes.slice(0, 6), 2);

	// Convert the 8-byte array into a 64-bit integer
	const timestampMs = new DataView(timestampBytes.buffer).getBigUint64(0, false);

	// Convert the timestamp from milliseconds to a Date object
	return new Date(Number(timestampMs));
}

interface CacheData {
	conversationsDict: { [id: string]: ConversationItem };
	conversationsList: ConversationItem[];
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
		const conversationsDict: { [id: string]: ConversationItem } = {};
		conversations.forEach(conv => {
			conversationsDict[conv.id] = conv;
		});

		this.cache = {
			conversationsDict,
			conversationsList: conversations,
			timestamp: Date.now(),
		};
	}

	private async listAllConversationsFromAPI(): Promise<ConversationItem[]> {
		let allConversations: ConversationItem[] = [];
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
				const convo: ConversationItem = { id: id, title: title, createdAt: getDateFromUUIDv7(id) };
				this.cache!.conversationsDict[id] = convo;
				this.cache!.conversationsList.unshift(convo);
			}
		}
	}

	getConversationsList(): ConversationItem[] {
		if (this.isCacheValid() && this.cache) {
			return this.cache.conversationsList;
		}
		return [];
	}
}
