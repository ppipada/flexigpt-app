import type { ModelName, ProviderName } from '@/models/aiprovidermodel';

export interface AISetting {
	isEnabled: boolean;
	apiKey: string;
	defaultModel: ModelName;
	defaultTemperature: number;
	origin: string;
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
