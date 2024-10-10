import { ModelName, ProviderName } from '@/models/aiprovidermodel';

export interface AISetting {
	apiKey: string;
	defaultModel: ModelName;
	defaultTemperature: number;
	defaultOrigin: string;
	additionalSettings: Record<string, any>;
}

export type SettingsSchema = {
	aiSettings: { [key: string]: AISetting };
	app: {
		defaultProvider: ProviderName;
	};
};

export interface ISettingStoreAPI {
	getAllSettings: () => Promise<SettingsSchema>;
	setSetting: (key: string, value: any) => Promise<void>;
}
