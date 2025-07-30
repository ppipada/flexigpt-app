import { type AddProviderRequest } from '@/spec/aiprovider';
import { type ProviderName } from '@/spec/modelpreset';
import { type AISetting, type AISettingAttrs } from '@/spec/setting';

import { providerSetAPI, settingstoreAPI } from '@/apis/baseapi';

export async function SetAppSettings(defaultProvider: ProviderName) {
	await settingstoreAPI.setAppSettings(defaultProvider);
	await providerSetAPI.setDefaultProvider(defaultProvider);
}

export async function AddAISetting(providerName: ProviderName, aiSetting: AISetting) {
	// Persist to setting store
	await settingstoreAPI.addAISetting(providerName, aiSetting);
	const req: AddProviderRequest = {
		provider: providerName,
		apiKey: aiSetting.apiKey,
		origin: aiSetting.origin,
		chatCompletionPathPrefix: aiSetting.chatCompletionPathPrefix,
	};
	await providerSetAPI.addProvider(req);
}

/**
 * @public
 */
export async function DeleteAISetting(providerName: ProviderName) {
	await settingstoreAPI.deleteAISetting(providerName);
	await providerSetAPI.deleteProvider(providerName);
}

export async function SetAISettingAPIKey(providerName: ProviderName, apiKey: string) {
	await settingstoreAPI.setAISettingAPIKey(providerName, apiKey);
	await providerSetAPI.setProviderAPIKey(providerName, apiKey);
}

export async function SetAISettingAttrs(providerName: ProviderName, aiSettingAttrs: AISettingAttrs) {
	await settingstoreAPI.setAISettingAttrs(providerName, aiSettingAttrs);
	if (typeof aiSettingAttrs.origin !== 'undefined' || typeof aiSettingAttrs.chatCompletionPathPrefix !== 'undefined') {
		await providerSetAPI.setProviderAttribute(
			providerName,
			aiSettingAttrs.origin,
			aiSettingAttrs.chatCompletionPathPrefix
		);
	}
}
