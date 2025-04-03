import { providerSetAPI, settingstoreAPI } from '@/backendapibase';
import type { ProviderName } from '@/models/aiprovidermodel';
import type { AISetting, SettingsSchema } from '@/models/settingmodel';

export function updateProviderAISettings(provider: ProviderName, settings: AISetting) {
	providerSetAPI.setAttribute(
		provider,
		settings.apiKey,
		settings.defaultModel,
		settings.defaultTemperature,
		settings.origin
	);
}

export async function loadProviderSettings(): Promise<SettingsSchema> {
	const settings = await settingstoreAPI.getAllSettings();

	const defaultProvider = settings.app.defaultProvider;
	await providerSetAPI.setDefaultProvider(defaultProvider);
	// Iterate over each entry in settings.aiSettings
	for (const [providerName, aiSettings] of Object.entries(settings.aiSettings)) {
		updateProviderAISettings(providerName, aiSettings);
	}

	return settings;
}
