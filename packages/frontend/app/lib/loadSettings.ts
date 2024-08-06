import { getAllSettings } from '@/api/settings';
import { ProviderName, providerSet } from 'aiprovider';
import { AISetting, SettingsSchema } from 'settingmodel';

export function updateProviderAISettings(provider: ProviderName, settings: AISetting) {
	const providerAPI = providerSet.getProviderAPI(provider);
	providerAPI.setAttribute(settings.apiKey, settings.defaultModel, settings.defaultTemperature, settings.defaultOrigin);
}

export async function loadProviderSettings(): Promise<SettingsSchema> {
	const settings = await getAllSettings();
	if (settings) {
		const defaultProvider = settings.app.defaultProvider as ProviderName;
		providerSet.setDefaultProvider(defaultProvider);
		updateProviderAISettings(ProviderName.ANTHROPIC, settings[ProviderName.ANTHROPIC]);
		updateProviderAISettings(ProviderName.GOOGLE, settings[ProviderName.GOOGLE]);
		updateProviderAISettings(ProviderName.HUGGINGFACE, settings[ProviderName.HUGGINGFACE]);
		updateProviderAISettings(ProviderName.LLAMACPP, settings[ProviderName.LLAMACPP]);
		updateProviderAISettings(ProviderName.OPENAI, settings[ProviderName.OPENAI]);
	}
	return settings;
}
