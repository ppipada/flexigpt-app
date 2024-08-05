import { ConversationMessage, IConversationAPI } from 'conversationmodel';
import { basename } from 'node:path';
import { CollectionMonthPartitioned, PATHORDER_DESC } from 'securejsondb';
// import { getSampleConversations } from './message_samples';
// import { log } from 'logger';
import { Conversation } from './store_types';

export class ConversationCollection implements IConversationAPI {
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

	async listConversations(
		token?: string
	): Promise<{ conversations: { id: string; title: string }[]; nextToken?: string }> {
		const { files, nextToken } = await this.partitionedCollection.listFiles(PATHORDER_DESC, token);
		const conversations = files.map(file => {
			const filename = basename(file);
			const [id, ...titleParts] = filename.replace('.json', '').split('_');
			const title = titleParts.join('_').replace(/_/g, ' ');
			return { id, title };
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
