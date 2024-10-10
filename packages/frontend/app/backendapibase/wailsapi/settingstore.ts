import { ISettingStoreAPI, SettingsSchema } from '@/models/settingmodel';

export class WailsSettingStoreAPI implements ISettingStoreAPI {
	// Implement the getAllSettings method
	async getAllSettings(): Promise<SettingsSchema> {
		return await window.SettingStoreAPI.getAllSettings();
	}

	// Implement the setSetting method
	async setSetting(key: string, value: any): Promise<void> {
		await window.SettingStoreAPI.setSetting(key, value);
	}
}
