import { PathLike } from 'node:fs';
import { SecureJSONFileDB, SecureSchema } from 'securejsondb';
import { defaultSettingsData, ISettingsAPI, sensitiveKeys, SettingsSchema } from 'settingmodel';

export type SecureSettingsSchema = SettingsSchema & SecureSchema;

export const defaultSecureSettingsData: SecureSettingsSchema = {
	...defaultSettingsData,
	_sensitiveKeys: sensitiveKeys,
};

export class SettingsStore implements ISettingsAPI {
	private db: SecureJSONFileDB<SecureSettingsSchema>;

	constructor(filename: PathLike) {
		this.db = new SecureJSONFileDB<SecureSettingsSchema>(filename, defaultSecureSettingsData);
	}

	async initialize() {
		await this.db.initialize();
	}

	async getAllSettings(): Promise<SecureSettingsSchema> {
		return await this.db.getAllData();
	}

	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	async setSetting(key: string, value: any): Promise<void> {
		await this.db.setData(key, value);
	}
}
