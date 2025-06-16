import type { ProviderName } from '@/models/aimodelmodel';
import type { AISetting, AISettingAttrs, ISettingStoreAPI, SettingsSchema } from '@/models/settingmodel';

import {
	AddAISetting,
	DeleteAISetting,
	GetAllSettings,
	SetAISettingAPIKey,
	SetAISettingAttrs,
	SetAppSettings,
} from '@/apis/wailsjs/go/main/SettingStoreWrapper';

import type { settingstore } from '../wailsjs/go/models';

/**
 * @public
 */
export class WailsSettingStoreAPI implements ISettingStoreAPI {
	// Implement the getAllSettings method
	async getAllSettings(): Promise<SettingsSchema> {
		const r: settingstore.GetAllSettingsRequest = { ForceFetch: false };
		const s = await GetAllSettings(r);
		return s.Body as SettingsSchema;
	}

	// async setSetting(key: string, value: any): Promise<void> {
	// 	const r = { Key: key, Body: { value: value } };
	// 	await SetSetting(r as spec.SetSettingRequest);
	// }

	async setAppSettings(defaultProvider: ProviderName): Promise<void> {
		const r = {
			Body: {
				defaultProvider: defaultProvider,
			},
		};
		await SetAppSettings(r as settingstore.SetAppSettingsRequest);
	}

	async addAISetting(providerName: ProviderName, aiSetting: AISetting): Promise<void> {
		const r = {
			ProviderName: providerName,
			Body: aiSetting,
		};
		await AddAISetting(r as settingstore.AddAISettingRequest);
	}

	async deleteAISetting(providerName: ProviderName): Promise<void> {
		const r = {
			ProviderName: providerName,
		};
		await DeleteAISetting(r as settingstore.DeleteAISettingRequest);
	}

	async setAISettingAPIKey(providerName: ProviderName, apiKey: string): Promise<void> {
		const r = {
			ProviderName: providerName,
			Body: {
				apiKey: apiKey,
			},
		};
		await SetAISettingAPIKey(r as settingstore.SetAISettingAPIKeyRequest);
	}

	async setAISettingAttrs(providerName: ProviderName, aiSettingAttrs: AISettingAttrs): Promise<void> {
		const r = {
			ProviderName: providerName,
			Body: {
				isEnabled: aiSettingAttrs.isEnabled,
				origin: aiSettingAttrs.origin,
				chatCompletionPathPrefix: aiSettingAttrs.chatCompletionPathPrefix,
				defaultModel: aiSettingAttrs.defaultModel,
			},
		};
		await SetAISettingAttrs(r as settingstore.SetAISettingAttrsRequest);
	}
}
