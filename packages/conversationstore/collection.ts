import { Conversation as ConversationBase, IConversationAPI } from 'conversationmodel';
import { basename } from 'node:path';
import { CollectionMonthPartitioned, SecureSchema } from 'securejsondb';
import { v7 as uuidv7 } from 'uuid';
import { messageSamplesList } from './message_samples';

export type Conversation = ConversationBase & SecureSchema;

export function getSampleConversation(): Conversation {
	return {
		id: uuidv7(),
		title: 'Sample Conversation',
		createdAt: new Date(),
		modifiedAt: new Date(),
		messages: messageSamplesList,
	};
}

export class ConversationCollection extends CollectionMonthPartitioned<Conversation> implements IConversationAPI {
	constructor(baseDir: string) {
		super(baseDir, { id: '', title: '', createdAt: new Date(), modifiedAt: new Date(), messages: [] });
		// const sampleConvo = getSampleConversation();
		// this.saveConversation(sampleConvo);
	}

	createNewConversation(title: string): Conversation {
		return {
			id: uuidv7(),
			title: title.substring(0, 64),
			createdAt: new Date(),
			modifiedAt: new Date(),
			messages: [],
		};
	}

	getConversationFilename(conversation: Conversation): string {
		return `${conversation.id}_${conversation.title.replace(/[^a-zA-Z0-9]/g, '_')}.json`;
	}

	async saveConversation(conversation: Conversation): Promise<void> {
		const filename = this.getConversationFilename(conversation);
		await this.addFile(filename, conversation);
	}

	async startConversation(title: string, oldConversation?: Conversation): Promise<Conversation> {
		if (oldConversation) {
			await this.saveConversation(oldConversation);
		}
		const conversation = this.createNewConversation(title);
		return conversation;
	}

	async deleteConversation(id: string, title: string): Promise<void> {
		const filename = this.getConversationFilename({
			id,
			title,
			createdAt: new Date(),
			modifiedAt: new Date(),
			messages: [],
		});
		await this.deleteFile(filename);
	}

	async getConversation(id: string, title: string): Promise<Conversation | null> {
		const filename = this.getConversationFilename({
			id,
			title,
			createdAt: new Date(),
			modifiedAt: new Date(),
			messages: [],
		});
		return this.getFile(filename);
	}

	async listConversations(
		token?: string
	): Promise<{ conversations: { id: string; title: string }[]; nextToken?: string }> {
		const { files, nextToken } = await this.listFiles(token);
		const conversations = files.map(file => {
			const filename = basename(file);
			const [id, ...titleParts] = filename.replace('.json', '').split('_');
			const title = titleParts.join('_').replace(/_/g, ' ');
			return { id, title };
		});
		return { conversations, nextToken };
	}
}
