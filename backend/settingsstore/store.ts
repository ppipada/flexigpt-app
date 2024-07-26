import { Low } from 'lowdb';
import { PathLike } from 'node:fs';
import { SecureJSONFile } from './secureAdapter';
import { SettingsSchema, defaultSettingsData } from './settingsSchema';

export async function getSettingsDB(filename: PathLike): Promise<Low<SettingsSchema>> {
	const adapter = new SecureJSONFile<SettingsSchema>(filename);
	const db = new Low<SettingsSchema>(adapter, defaultSettingsData);
	await db.read();
	return db;
}

export class SettingsStore {
	private db!: Low<SettingsSchema>;

	constructor(private filename: PathLike) {}

	async initialize() {
		this.db = await getSettingsDB(this.filename);
	}

	async getAllSettings(): Promise<SettingsSchema> {
		await this.db.read();
		return this.db.data;
	}

	async setSetting(key: string, value: any): Promise<void> {
		await this.db.read();

		const keys = key.split('.');
		const lastKey = keys.pop();
		const target = keys.reduce((o, k) => (o[k] = o[k] || {}), this.db.data);

		if (lastKey) target[lastKey] = value;

		await this.db.write();
	}
}
