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
	async setSetting(dotSeparatedKey: string, value: any): Promise<void> {
		const keys = dotSeparatedKey.split('.');
		let currentSchema: any = defaultSecureSettingsData;

		for (let i = 0; i < keys.length; i++) {
			if (!(keys[i] in currentSchema)) {
				throw new Error(`Invalid key: ${keys.slice(0, i + 1).join('.')}`);
			}
			currentSchema = currentSchema[keys[i]];
		}

		const expectedType = typeof currentSchema;
		const valueType = typeof value;

		if (expectedType !== valueType) {
			throw new Error(`Type mismatch for key "${dotSeparatedKey}": expected ${expectedType}, got ${valueType}`);
		}

		// If types match, set the data
		await this.db.setData(dotSeparatedKey, value);
	}
}
