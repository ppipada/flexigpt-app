import { ModelName, ProviderName } from 'aiprovider';

export interface AISetting {
	apiKey: string;
	defaultModel: ModelName;
	defaultTemperature: number;
	defaultOrigin: string;
	additionalSettings: Record<string, any>;
}
export type SettingsSchema = {
	app: {
		defaultProvider: ProviderName;
	};
	aiSettings: { [key: string]: AISetting };
};

export interface ISettingStoreAPI {
	getAllSettings: () => Promise<SettingsSchema>;
	setSetting: (key: string, value: any) => Promise<void>;
}
