import { type ProviderName } from '@/models/aimodelmodel';

export interface AISetting {
	isEnabled: boolean;
	apiKey: string;
	origin: string;
	chatCompletionPathPrefix: string;
}

export interface AISettingAttrs {
	isEnabled?: boolean;
	origin?: string;
	chatCompletionPathPrefix?: string;
}

export type SettingsSchema = {
	aiSettings: Record<ProviderName, AISetting>;
	app: {
		defaultProvider: ProviderName;
	};
};

export interface ISettingStoreAPI {
	getAllSettings: () => Promise<SettingsSchema>;
	setAppSettings: (defaultProvider: ProviderName) => Promise<void>;
	addAISetting: (providerName: ProviderName, aiSetting: AISetting) => Promise<void>;
	deleteAISetting: (providerName: ProviderName) => Promise<void>;
	setAISettingAPIKey: (providerName: ProviderName, apiKey: string) => Promise<void>;
	setAISettingAttrs: (providerName: ProviderName, aiSettingAttrs: AISettingAttrs) => Promise<void>;
}
