import {
	DefaultModelParams,
	type ModelName,
	type ModelParams,
	type ModelPreset,
	type ProviderName,
} from '@/models/aiprovidermodel';

export interface AISetting {
	isEnabled: boolean;
	apiKey: string;
	origin: string;
	chatCompletionPathPrefix: string;
	defaultModel: ModelName;
	modelPresets: Record<ModelName, ModelPreset>;
}

export interface AISettingAttrs {
	isEnabled?: boolean;
	origin?: string;
	chatCompletionPathPrefix?: string;
	defaultModel?: ModelName;
}

export interface ChatOptions extends ModelParams {
	title: string;
	provider: ProviderName;
	disablePreviousMessages: boolean;
}

export const DefaultChatOptions: ChatOptions = {
	...DefaultModelParams,
	provider: 'No Provider',
	name: 'No Model',
	title: 'No Model configured',
	disablePreviousMessages: false,
};

export type SettingsSchema = {
	aiSettings: Record<ProviderName, AISetting>;
	app: {
		defaultProvider: ProviderName;
	};
};

export interface ISettingStoreAPI {
	getAllSettings: () => Promise<SettingsSchema>;
	// setSetting: (key: string, value: any) => Promise<void>;
	setAppSettings: (defaultProvider: ProviderName) => Promise<void>;
	addAISetting: (providerName: ProviderName, aiSetting: AISetting) => Promise<void>;
	deleteAISetting: (providerName: ProviderName) => Promise<void>;
	setAISettingAPIKey: (providerName: ProviderName, apiKey: string) => Promise<void>;
	setAISettingAttrs: (providerName: ProviderName, aiSettingAttrs: AISettingAttrs) => Promise<void>;
	addModelPreset: (providerName: ProviderName, modelName: ModelName, modelPreset: ModelPreset) => Promise<void>;
	deleteModelPreset: (providerName: ProviderName, modelName: ModelName) => Promise<void>;
}
