import { log } from '@/logger';
import { parse, v7 as uuidv7 } from 'uuid';
import { Conversation, ConversationMessage, ConversationRoleEnum } from './conversation_types';

export function initConversation(title: string = 'New Conversation'): Conversation {
	return {
		id: uuidv7(),
		title: title.substring(0, 64),
		createdAt: new Date(),
		modifiedAt: new Date(),
		messages: [],
	};
}

export function initConversationMessage(role: ConversationRoleEnum, content: string): ConversationMessage {
	const d = new Date();
	return {
		id: d.toISOString(),
		createdAt: new Date(),
		role: role,
		content: content,
	};
}

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
