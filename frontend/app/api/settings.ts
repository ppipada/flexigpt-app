import { SettingsSchema } from 'sharedpkg/settings/types';
export async function getAllSettings(): Promise<SettingsSchema> {
	return await window.SettingsAPI.getAllSettings();
}

export async function setSetting(key: string, value: any) {
	await window.SettingsAPI.setSetting(key, value);
}
