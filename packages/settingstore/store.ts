import { Low } from 'lowdb';
import { PathLike } from 'node:fs';
import { SecureJSONFile, SecureSchema } from 'securejsondb';
import { defaultSettingsData, SettingsSchema } from 'settingmodel';

export type SecureSettingsSchema = SettingsSchema & SecureSchema;

export const defaultSecureSettingsData: SecureSettingsSchema = {
	...defaultSettingsData,
	_sensitiveKeys: ['openai.apiKey', 'anthropic.apiKey', 'huggingface.apiKey', 'googlegl.apiKey', 'llamacpp.apiKey'],
};

export async function getSettingsDB(filename: PathLike): Promise<Low<SecureSettingsSchema>> {
	const adapter = new SecureJSONFile<SecureSettingsSchema>(filename);
	const db = new Low<SecureSettingsSchema>(adapter, defaultSecureSettingsData);
	await db.read();
	return db;
}

export class SettingsStore {
	private db!: Low<SecureSettingsSchema>;

	constructor(private filename: PathLike) {}

	async initialize() {
		this.db = await getSettingsDB(this.filename);
	}

	async getAllSettings(): Promise<SecureSettingsSchema> {
		await this.db.read();
		return this.db.data;
	}

	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	async setSetting(key: string, value: any): Promise<void> {
		await this.db.read();

		const keys = key.split('.');
		const lastKey = keys.pop();
		const target = keys.reduce((o, k) => (o[k] = o[k] || {}), this.db.data);

		if (lastKey) target[lastKey] = value;

		await this.db.write();
	}
}
