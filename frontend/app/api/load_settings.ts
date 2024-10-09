import { setAttribute, setDefaultProvider } from '@/api/base_aiproviderimpl';
import { getAllSettings } from '@/api/base_settings';
import { ProviderName } from '@/models/aiprovidermodel';
import { AISetting, SettingsSchema } from '@/models/settingmodel';

export function updateProviderAISettings(provider: ProviderName, settings: AISetting) {
	setAttribute(provider, settings.apiKey, settings.defaultModel, settings.defaultTemperature, settings.defaultOrigin);
}

export async function loadProviderSettings(): Promise<SettingsSchema> {
	const settings = await getAllSettings();
	if (settings) {
		const defaultProvider = settings.app.defaultProvider as ProviderName;
		setDefaultProvider(defaultProvider);
		updateProviderAISettings(ProviderName.ANTHROPIC, settings[ProviderName.ANTHROPIC]);
		updateProviderAISettings(ProviderName.GOOGLE, settings[ProviderName.GOOGLE]);
		updateProviderAISettings(ProviderName.HUGGINGFACE, settings[ProviderName.HUGGINGFACE]);
		updateProviderAISettings(ProviderName.LLAMACPP, settings[ProviderName.LLAMACPP]);
		updateProviderAISettings(ProviderName.OPENAI, settings[ProviderName.OPENAI]);
	}
	return settings;
}
