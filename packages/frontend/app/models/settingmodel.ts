import { ModelName, ProviderName } from '@/models/aiprovidermodel';

export interface AISetting {
	apiKey: string;
	defaultModel: ModelName;
	defaultTemperature: number;
	defaultOrigin: string;
	additionalSettings: Record<string, any>;
}
export interface AISettingsSchema {
	[ProviderName.ANTHROPIC]: AISetting;
	[ProviderName.GOOGLE]: AISetting;
	[ProviderName.HUGGINGFACE]: AISetting;
	[ProviderName.LLAMACPP]: AISetting;
	[ProviderName.OPENAI]: AISetting;
}

export type SettingsSchema = AISettingsSchema & {
	app: {
		defaultProvider: ProviderName;
	};
};

export interface ISettingStoreAPI {
	getAllSettings: () => Promise<SettingsSchema>;
	setSetting: (key: string, value: any) => Promise<void>;
}
