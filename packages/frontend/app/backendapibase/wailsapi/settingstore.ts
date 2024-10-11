import { GetAllSettings, SetSetting } from '@/backendapibase/wailsjs/go/settingstore/SettingStore';
import { ISettingStoreAPI, SettingsSchema } from '@/models/settingmodel';

export class WailsSettingStoreAPI implements ISettingStoreAPI {
	// Implement the getAllSettings method
	async getAllSettings(): Promise<SettingsSchema> {
		const forceFetch = false;
		const s = await GetAllSettings(forceFetch);
		return s as SettingsSchema;
	}

	async setSetting(key: string, value: any): Promise<void> {
		await SetSetting(key, value);
	}
}
