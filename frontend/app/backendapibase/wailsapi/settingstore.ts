import { GetAllSettings, SetSetting } from '@/backendapibase/wailsjs/go/main/SettingStoreWrapper';
import { ISettingStoreAPI, SettingsSchema } from '@/models/settingmodel';
import { spec } from '../wailsjs/go/models';

export class WailsSettingStoreAPI implements ISettingStoreAPI {
	// Implement the getAllSettings method
	async getAllSettings(): Promise<SettingsSchema> {
		const r: spec.GetAllSettingsRequest = { ForceFetch: false };
		const s = await GetAllSettings(r);
		return s.Body as SettingsSchema;
	}

	async setSetting(key: string, value: any): Promise<void> {
		const r = { Key: key, Body: { value: value } };
		await SetSetting(r as spec.SetSettingRequest);
	}
}
