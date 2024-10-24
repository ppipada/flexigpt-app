import { basename } from 'node:path';
import { CollectionMonthPartitioned, PATHORDER_DESC } from 'securejsondb';
import { ConversationItem, ConversationMessage, IConversationStoreAPI } from './conversation_types';
// import { getSampleConversations } from './message_samples';
import { Conversation } from './store_types';

import { log } from 'logger';
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

export class ConversationCollection implements IConversationStoreAPI {
	private partitionedCollection: CollectionMonthPartitioned<Conversation>;

	constructor(baseDir: string) {
		this.partitionedCollection = new CollectionMonthPartitioned<Conversation>(baseDir, {
			id: '',
			title: '',
			createdAt: new Date(),
			modifiedAt: new Date(),
			messages: [],
		});
		// const sampleConvos = getSampleConversations();
		// sampleConvos.map(convo => this.saveConversation(convo));
	}

	getConversationFilename(conversation: Conversation): string {
		return `${conversation.id}_${conversation.title.replace(/[^a-zA-Z0-9]/g, '_')}.json`;
	}

	async saveConversation(conversation: Conversation): Promise<void> {
		const filename = this.getConversationFilename(conversation);
		await this.partitionedCollection.addFile(filename, conversation);
	}

	async deleteConversation(id: string, title: string): Promise<void> {
		const filename = this.getConversationFilename({
			id,
			title,
			createdAt: new Date(),
			modifiedAt: new Date(),
			messages: [],
		});
		await this.partitionedCollection.deleteFile(filename);
	}

	async getConversation(id: string, title: string): Promise<Conversation | null> {
		const filename = this.getConversationFilename({
			id,
			title,
			createdAt: new Date(),
			modifiedAt: new Date(),
			messages: [],
		});
		return this.partitionedCollection.getFile(filename);
	}

	async listConversations(token?: string): Promise<{ conversations: ConversationItem[]; nextToken?: string }> {
		const { files, nextToken } = await this.partitionedCollection.listFiles(PATHORDER_DESC, token);
		const conversations = files.map(file => {
			const filename = basename(file);
			const [id, ...titleParts] = filename.replace('.json', '').split('_');
			const title = titleParts.join('_').replace(/_/g, ' ');
			const convo: ConversationItem = {
				id: id,
				title: title,
				createdAt: getDateFromUUIDv7(id),
			};
			return convo;
		});
		return { conversations, nextToken };
	}

	async addMessageToConversation(id: string, title: string, newMessage: ConversationMessage): Promise<void> {
		const conversation = await this.getConversation(id, title);
		if (!conversation) {
			throw new Error('Conversation not found');
		}
		conversation.messages.push(newMessage);
		conversation.modifiedAt = new Date();
		await this.saveConversation(conversation);
	}
}
