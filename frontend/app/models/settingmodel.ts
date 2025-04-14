import {
	DefaultModelParams,
	type ModelName,
	type ModelParams,
	type ProviderName,
	type ReasoningParams,
} from '@/models/aiprovidermodel';

export interface AISetting {
	isEnabled: boolean;
	apiKey: string;
	origin: string;
	chatCompletionPathPrefix: string;
	defaultModel: ModelName;
	modelSettings: Record<ModelName, ModelSetting>;
}

export interface AISettingAttrs {
	isEnabled?: boolean;
	origin?: string;
	chatCompletionPathPrefix?: string;
	defaultModel?: ModelName;
}

export interface ModelSetting {
	displayName: string;
	isEnabled: boolean;
	stream?: boolean;
	maxPromptLength?: number;
	maxOutputLength?: number;
	temperature?: number;
	reasoning?: ReasoningParams;
	systemPrompt?: string;
	timeout?: number;
	additionalParameters?: Record<string, any>;
}

export const DefaultModelSetting: ModelSetting = {
	displayName: '',
	isEnabled: true,
	stream: DefaultModelParams.stream,
	maxPromptLength: DefaultModelParams.maxPromptLength,
	maxOutputLength: DefaultModelParams.maxOutputLength,
	temperature: DefaultModelParams.temperature ?? 0.1,
	reasoning: DefaultModelParams.reasoning,
	systemPrompt: DefaultModelParams.systemPrompt,
	timeout: DefaultModelParams.timeout,
	additionalParameters: DefaultModelParams.additionalParameters,
};

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
	addModelSetting: (providerName: ProviderName, modelName: ModelName, modelSetting: ModelSetting) => Promise<void>;
	deleteModelSetting: (providerName: ProviderName, modelName: ModelName) => Promise<void>;
}
