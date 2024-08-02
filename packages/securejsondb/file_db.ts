import { safeStorage } from 'electron';
import { PathLike } from 'fs';
import { Low } from 'lowdb';
import { DataFile } from 'lowdb/node';
import { SecureSchema } from './schema';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function getNestedProperty(obj: any, path: string) {
	return path.split('.').reduce((o, p) => (o ? o[p] : undefined), obj);
}
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function setNestedProperty(obj: any, path: string, value: any) {
	const keys = path.split('.');
	const lastKey = keys.pop();
	const lastObj = keys.reduce((o, k) => (o[k] = o[k] || {}), obj);
	if (lastKey) lastObj[lastKey] = value;
}

export class SecureJSONFileAdapter<T extends SecureSchema> extends DataFile<T> {
	constructor(filename: PathLike) {
		super(filename, {
			parse: (text: string) => {
				const parsedData = JSON.parse(text) as T;
				const sensitiveKeys = parsedData._sensitiveKeys || [];

				// Decrypt keys
				sensitiveKeys.forEach(key => {
					const value = getNestedProperty(parsedData, key);
					if (value) {
						setNestedProperty(parsedData, key, safeStorage.decryptString(Buffer.from(value, 'base64')));
					}
				});

				return parsedData;
			},
			stringify: (data: T) => {
				const encryptedData = { ...data };
				const sensitiveKeys = encryptedData._sensitiveKeys || [];

				// Encrypt keys
				sensitiveKeys.forEach(key => {
					const value = getNestedProperty(encryptedData, key);
					if (value) {
						setNestedProperty(encryptedData, key, safeStorage.encryptString(value).toString('base64'));
					}
				});

				return JSON.stringify(encryptedData, null, 2);
			},
		});
	}
}

export class SecureJSONFileDB<T extends SecureSchema> {
	private db!: Low<T>;

	constructor(
		private filename: PathLike,
		private defaultData: T
	) {}

	async initialize(): Promise<void> {
		const adapter = new SecureJSONFileAdapter<T>(this.filename);
		this.db = new Low<T>(adapter, this.defaultData);
		await this.db.read();
	}

	async getAllData(): Promise<T> {
		await this.db.read();
		return this.db.data;
	}

	async setData(dotSeparatedKey: string, value: any): Promise<void> {
		await this.db.read();

		const keys = dotSeparatedKey.split('.');
		const lastKey = keys.pop();
		// Use type assertion here to assure TypeScript that we're working with an object.
		let target = this.db.data as Record<string, any>;

		for (const key of keys) {
			if (!target[key]) {
				target[key] = {};
			}
			target = target[key];
		}

		if (lastKey) {
			target[lastKey] = value;
		}

		await this.db.write();
	}

	async overwriteData(newData: T): Promise<void> {
		this.db.data = newData;
		await this.db.write();
	}
}
