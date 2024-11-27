export interface DocStoreServer {
	id: string;
	name: string;
	url: string;
	collections: Collection[];
	status: 'online' | 'offline';
	description: string;
	dbName: string;
}

export interface Collection {
	id: string;
	name: string;
	command: string;
	documentCount: number;
	chunkCount: number;
}
