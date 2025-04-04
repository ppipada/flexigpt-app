import type { ModelName, ProviderName } from '@/models/aiprovidermodel';

export interface AISetting {
	isEnabled: boolean;
	apiKey: string;
	origin: string;
	defaultModel: ModelName;
	modelSettings: ModelSetting[];
}

export interface ModelSetting {
	name: ModelName;
	displayName: string;
	isEnabled: boolean;
	stream?: boolean;
	promptLength?: number;
	outputLength?: number;
	temperature?: number;
	reasoningSupport?: boolean;
	systemPrompt?: string;
	timeout?: number;
	additionalParameters?: Record<string, any>;
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
