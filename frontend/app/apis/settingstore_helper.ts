import { providerSetAPI, settingstoreAPI } from '@/apis/baseapi';
import {
	DefaultModelParams,
	type ModelInfo,
	type ModelName,
	type ModelParams,
	type ProviderInfo,
	type ProviderName,
} from '@/models/aiprovidermodel';
import { DefaultModelSetting, type AISetting, type ModelSetting, type SettingsSchema } from '@/models/settingmodel';

export interface ModelOption extends ModelParams {
	title: string;
	provider: ProviderName;
}

export const DefaultModelOption: ModelOption = {
	...DefaultModelParams,
	provider: 'No Provider',
	name: 'No Model',
	title: 'No Model configured',
};

export interface ChatOptions {
	modelInfo: ModelOption;
	disablePreviousMessages: boolean;
}

export function UpdateProviderAISettings(provider: ProviderName, settings: AISetting) {
	providerSetAPI.setAttribute(provider, settings.apiKey, settings.origin, settings.chatCompletionPathPrefix);
}

export async function loadProviderSettings(): Promise<SettingsSchema> {
	const settings = await settingstoreAPI.getAllSettings();

	const defaultProvider = settings.app.defaultProvider;
	await providerSetAPI.setDefaultProvider(defaultProvider);
	// Iterate over each entry in settings.aiSettings
	for (const [providerName, aiSettings] of Object.entries(settings.aiSettings)) {
		UpdateProviderAISettings(providerName, aiSettings);
	}

	return settings;
}

function mergeModelSettingModelInfoToOption(
	providerName: ProviderName,
	modelName: ModelName,
	providerInfoDict: Record<ProviderName, ProviderInfo>,
	modelSetting: ModelSetting
): ModelOption {
	let modelInfo: ModelInfo | undefined = undefined;
	if (providerName in providerInfoDict && modelName in providerInfoDict[providerName].models) {
		modelInfo = providerInfoDict[providerName].models[modelName];
	}

	let temperature = DefaultModelParams.temperature ?? 0.1;
	if (typeof modelSetting.temperature !== 'undefined') {
		temperature = modelSetting.temperature;
	} else if (modelInfo) {
		temperature = modelInfo.defaultTemperature;
	}

	let stream = DefaultModelParams.stream;
	if (typeof modelSetting.stream !== 'undefined') {
		stream = modelSetting.stream;
	} else if (modelInfo) {
		stream = modelInfo.streamingSupport;
	}

	let maxPromptLength = DefaultModelParams.maxPromptLength;
	if (typeof modelSetting.maxPromptLength !== 'undefined') {
		maxPromptLength = modelSetting.maxPromptLength;
	} else if (modelInfo) {
		maxPromptLength = modelInfo.maxPromptLength;
	}

	let maxOutputLength = DefaultModelParams.maxOutputLength;
	if (typeof modelSetting.maxOutputLength !== 'undefined') {
		maxOutputLength = modelSetting.maxOutputLength;
	} else if (modelInfo) {
		maxOutputLength = modelInfo.maxOutputLength;
	}

	let reasoningSupport = DefaultModelParams.reasoningSupport;
	if (typeof modelSetting.reasoningSupport !== 'undefined') {
		reasoningSupport = modelSetting.reasoningSupport;
	} else if (modelInfo) {
		reasoningSupport = modelInfo.reasoningSupport;
	}

	let systemPrompt = DefaultModelParams.systemPrompt;
	if (typeof modelSetting.systemPrompt !== 'undefined') {
		systemPrompt = modelSetting.systemPrompt;
	} else if (modelInfo) {
		systemPrompt = modelInfo.defaultSystemPrompt;
	}

	let timeout = DefaultModelParams.timeout;
	if (typeof modelSetting.timeout !== 'undefined') {
		timeout = modelSetting.timeout;
	} else if (modelInfo) {
		timeout = modelInfo.timeout;
	}

	let additionalParameters = { ...DefaultModelParams.additionalParameters };
	if (modelSetting.additionalParameters) {
		additionalParameters = { ...additionalParameters, ...modelSetting.additionalParameters };
	}

	return {
		title: modelSetting.displayName,
		provider: providerName,
		name: modelName,
		stream: stream,
		maxPromptLength: maxPromptLength,
		maxOutputLength: maxOutputLength,
		temperature: temperature,
		reasoningSupport: reasoningSupport,
		systemPrompt: systemPrompt,
		timeout: timeout,
		additionalParameters: additionalParameters,
	};
}

export async function GetChatInputOptions() {
	try {
		// Fetch configuration info and settings
		const info = await providerSetAPI.getConfigurationInfo();
		if (info.defaultProvider === '' || Object.keys(info.configuredProviders).length === 0) {
			return { allOptions: [DefaultModelOption], default: DefaultModelOption };
		}
		const configDefaultProvider = info.defaultProvider;
		const providerInfoDict = info.configuredProviders;

		const settings = await settingstoreAPI.getAllSettings();
		// Initialize default option and input models array
		let defaultOption: ModelOption | undefined;
		const inputModels: ModelOption[] = [];

		for (const providerName of Object.keys(providerInfoDict)) {
			const aiSetting = settings.aiSettings[providerName];
			if (aiSetting.isEnabled) {
				const settingsDefaultModelName = aiSetting.defaultModel;
				for (const [modelName, modelSetting] of Object.entries(aiSetting.modelSettings)) {
					if (!modelSetting.isEnabled) {
						continue;
					}
					const modelOption = mergeModelSettingModelInfoToOption(
						providerName,
						modelName,
						providerInfoDict,
						modelSetting
					);
					inputModels.push(modelOption);
					if (modelName === settingsDefaultModelName && providerName === configDefaultProvider) {
						defaultOption = modelOption;
					}
				}
			}
		}

		if (defaultOption === undefined || inputModels.length === 0) {
			throw Error('No input option found !!!');
		}
		return { allOptions: inputModels, default: defaultOption };
	} catch (error) {
		console.error('Error fetching chat input options:', error);
		return { allOptions: [DefaultModelOption], default: DefaultModelOption };
	}
}

// Create a Model setting from default or ModelInfo if available
export async function PopulateModelSettingDefaults(
	providerName: ProviderName,
	modelName: ModelName,
	existingData?: ModelSetting
): Promise<ModelSetting> {
	const info = await providerSetAPI.getConfigurationInfo();
	if (info.defaultProvider === '' || Object.keys(info.configuredProviders).length === 0) {
		return DefaultModelSetting;
	}

	const providerInfoDict = info.configuredProviders;

	// locate provider + model info in that config
	if (!(providerName in providerInfoDict)) {
		return DefaultModelSetting;
	}
	if (!(modelName in providerInfoDict[providerName].models)) {
		return DefaultModelSetting;
	}
	const modelInfo = providerInfoDict[providerName].models[modelName];

	// if modelInfo is present, use its known defaults
	const defaultsFromModelInfo: ModelSetting = {
		displayName: modelInfo.displayName || '',
		isEnabled: true,
		stream: modelInfo.streamingSupport,
		maxPromptLength: modelInfo.maxPromptLength,
		maxOutputLength: modelInfo.maxOutputLength,
		temperature: modelInfo.defaultTemperature,
		reasoningSupport: modelInfo.reasoningSupport,
		systemPrompt: modelInfo.defaultSystemPrompt,
		timeout: modelInfo.timeout,
	};
	// combine fallback -> modelInfo defaults -> existingData
	return {
		...defaultsFromModelInfo,
		...existingData,
	};
}
