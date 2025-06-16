import type { ModelName, ModelPreset, ProviderName } from '@/models/aiprovidermodel';
import type { AISetting, AISettingAttrs, ISettingStoreAPI, SettingsSchema } from '@/models/settingmodel';

import {
	AddAISetting,
	AddModelPreset,
	DeleteAISetting,
	DeleteModelPreset,
	GetAllSettings,
	SetAISettingAPIKey,
	SetAISettingAttrs,
	SetAppSettings,
} from '@/apis/wailsjs/go/main/SettingStoreWrapper';

import type { spec } from '../wailsjs/go/models';

/**
 * @public
 */
export class WailsSettingStoreAPI implements ISettingStoreAPI {
	// Implement the getAllSettings method
	async getAllSettings(): Promise<SettingsSchema> {
		const r: spec.GetAllSettingsRequest = { ForceFetch: false };
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
		await SetAppSettings(r as spec.SetAppSettingsRequest);
	}

	async addAISetting(providerName: ProviderName, aiSetting: AISetting): Promise<void> {
		const r = {
			ProviderName: providerName,
			Body: aiSetting,
		};
		await AddAISetting(r as spec.AddAISettingRequest);
	}

	async deleteAISetting(providerName: ProviderName): Promise<void> {
		const r = {
			ProviderName: providerName,
		};
		await DeleteAISetting(r as spec.DeleteAISettingRequest);
	}

	async setAISettingAPIKey(providerName: ProviderName, apiKey: string): Promise<void> {
		const r = {
			ProviderName: providerName,
			Body: {
				apiKey: apiKey,
			},
		};
		await SetAISettingAPIKey(r as spec.SetAISettingAPIKeyRequest);
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
		await SetAISettingAttrs(r as spec.SetAISettingAttrsRequest);
	}

	async addModelPreset(providerName: ProviderName, modelName: ModelName, modelPreset: ModelPreset): Promise<void> {
		const r = {
			ProviderName: providerName,
			ModelName: modelName,
			Body: modelPreset,
		};
		await AddModelPreset(r as spec.AddModelPresetRequest);
	}

	async deleteModelPreset(providerName: ProviderName, modelName: ModelName): Promise<void> {
		const r = {
			ProviderName: providerName,
			ModelName: modelName,
		};
		await DeleteModelPreset(r as spec.DeleteModelPresetRequest);
	}
}
