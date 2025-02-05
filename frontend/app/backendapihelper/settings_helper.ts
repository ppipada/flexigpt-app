import { providerSetAPI, settingstoreAPI } from '@/backendapibase';
import { ProviderName } from '@/models/aiprovidermodel';
import { AISetting, SettingsSchema } from '@/models/settingmodel';

export function updateProviderAISettings(provider: ProviderName, settings: AISetting) {
	providerSetAPI.setAttribute(
		provider,
		settings.apiKey,
		settings.defaultModel,
		settings.defaultTemperature,
		settings.defaultOrigin
	);
}

export async function loadProviderSettings(): Promise<SettingsSchema> {
	const settings = await settingstoreAPI.getAllSettings();
	if (settings) {
		const defaultProvider = settings.app.defaultProvider as ProviderName;
		providerSetAPI.setDefaultProvider(defaultProvider);
		// Iterate over each entry in settings.aiSettings
		for (const [providerName, aiSettings] of Object.entries(settings.aiSettings)) {
			updateProviderAISettings(providerName as ProviderName, aiSettings);
		}
	}
	return settings;
}
