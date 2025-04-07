import { providerSetAPI, settingstoreAPI } from '@/apis/baseapi';
import type { ModelName, ProviderName } from '@/models/aiprovidermodel';
import type { AISetting, ModelSetting, SettingsSchema } from '@/models/settingmodel';

export interface ModelOption {
	title: string;
	provider: ProviderName;
	name: ModelName;
	temperature: number;
}

export function UpdateProviderAISettings(provider: ProviderName, settings: AISetting) {
	providerSetAPI.setAttribute(provider, settings.apiKey, settings.defaultModel, settings.origin);
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

// Helper function to handle cases with no configured models
function handleNoConfiguredModels() {
	const noModel: ModelOption = {
		provider: 'No Provider',
		name: 'No Model',
		title: 'No Model configured',
		temperature: 0.1,
	};
	return { allOptions: [noModel], default: noModel };
}

export async function GetChatInputOptions() {
	try {
		// Fetch configuration info and settings
		const info = await providerSetAPI.getConfigurationInfo();
		if (!info || info.defaultProvider === '' || Object.keys(info.configuredProviders).length === 0) {
			return handleNoConfiguredModels();
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
				for (const [modelName, modelParams] of Object.entries(aiSetting.modelSettings)) {
					if (!modelParams.isEnabled) {
						continue;
					}
					let temperature = 0.1;
					if (typeof modelParams.temperature !== 'undefined') {
						temperature = modelParams.temperature;
					} else if (modelName in providerInfoDict[providerName].models) {
						temperature = providerInfoDict[providerName].models[modelName].defaultTemperature;
					}
					const modelOption = {
						title: modelParams.displayName,
						provider: providerName,
						temperature: temperature,
						name: modelName,
					};
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
		return handleNoConfiguredModels();
	}
}

// ----------------------------------------------------
// Utility: merges model defaults from providerInfo with
// an existing model setting (if any).
// ----------------------------------------------------
export async function PopulateModelSettingDefaults(
	providerName: ProviderName,
	modelName: ModelName,
	existingData?: ModelSetting
): Promise<ModelSetting> {
	// define some global fallback defaults
	const fallback: ModelSetting = {
		displayName: '',
		isEnabled: true,
		stream: false,
		promptLength: 2048,
		outputLength: 1024,
		temperature: 0.1,
		reasoningSupport: false,
		systemPrompt: '',
		timeout: 60,
	};
	// fetch overall config
	const info = await providerSetAPI.getConfigurationInfo();
	if (!info || info.defaultProvider === '' || Object.keys(info.configuredProviders).length === 0) {
		return fallback;
	}

	const providerInfoDict = info.configuredProviders;

	// locate provider + model info in that config
	if (!(providerName in providerInfoDict)) {
		return fallback;
	}
	if (!(modelName in providerInfoDict[providerName].models)) {
		return fallback;
	}
	const modelInfo = providerInfoDict[providerName].models[modelName];

	// if modelInfo is present, use its known defaults
	const defaultsFromModelInfo: ModelSetting = {
		displayName: modelInfo.displayName || '',
		isEnabled: true,
		stream: modelInfo.streamingSupport,
		promptLength: modelInfo.maxPromptLength,
		outputLength: modelInfo.maxOutputLength,
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
