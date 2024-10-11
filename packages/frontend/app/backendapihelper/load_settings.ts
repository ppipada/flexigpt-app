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
		updateProviderAISettings(ProviderName.ANTHROPIC, settings.aiSettings[ProviderName.ANTHROPIC]);
		updateProviderAISettings(ProviderName.GOOGLE, settings.aiSettings[ProviderName.GOOGLE]);
		updateProviderAISettings(ProviderName.HUGGINGFACE, settings.aiSettings[ProviderName.HUGGINGFACE]);
		updateProviderAISettings(ProviderName.LLAMACPP, settings.aiSettings[ProviderName.LLAMACPP]);
		updateProviderAISettings(ProviderName.OPENAI, settings.aiSettings[ProviderName.OPENAI]);
	}
	return settings;
}
