import { DefaultModelParams, type ModelName, type ProviderName } from '@/models/aiprovidermodel';

export interface AISetting {
	isEnabled: boolean;
	apiKey: string;
	origin: string;
	defaultModel: ModelName;
	modelSettings: Record<ModelName, ModelSetting>;
}

export interface ModelSetting {
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

export const DefaultModelSetting: ModelSetting = {
	displayName: '',
	isEnabled: true,
	stream: DefaultModelParams.stream,
	promptLength: DefaultModelParams.maxPromptLength,
	outputLength: DefaultModelParams.maxOutputLength,
	temperature: DefaultModelParams.temperature ?? 0.1,
	reasoningSupport: DefaultModelParams.reasoningSupport,
	systemPrompt: DefaultModelParams.systemPrompt,
	timeout: DefaultModelParams.timeout,
	additionalParameters: DefaultModelParams.additionalParameters,
};

export type SettingsSchema = {
	aiSettings: Record<ProviderName, AISetting>;
	app: {
		defaultProvider: ProviderName;
	};
};

export interface ISettingStoreAPI {
	getAllSettings: () => Promise<SettingsSchema>;
	setSetting: (key: string, value: any) => Promise<void>;
}
