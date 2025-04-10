import {
	type AddProviderRequest,
	DefaultModelParams,
	type ModelName,
	type ModelParams,
	type ProviderName,
} from '@/models/aiprovidermodel';
import { type AISetting, type AISettingAttrs, DefaultModelSetting, type ModelSetting } from '@/models/settingmodel';

import { providerSetAPI, settingstoreAPI } from '@/apis/baseapi';

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

export async function SetAppSettings(defaultProvider: ProviderName) {
	await settingstoreAPI.setAppSettings(defaultProvider);
	await providerSetAPI.setDefaultProvider(defaultProvider);
}

export async function AddAISetting(providerName: ProviderName, aiSetting: AISetting) {
	// Persist to setting store
	await settingstoreAPI.addAISetting(providerName, aiSetting);
	const req: AddProviderRequest = {
		provider: providerName,
		apiKey: aiSetting.apiKey,
		origin: aiSetting.origin,
		chatCompletionPathPrefix: aiSetting.chatCompletionPathPrefix,
	};
	await providerSetAPI.addProvider(req);
}

/**
 * @public
 */
export async function DeleteAISetting(providerName: ProviderName) {
	await settingstoreAPI.deleteAISetting(providerName);
	await providerSetAPI.deleteProvider(providerName);
}

export async function SetAISettingAPIKey(providerName: ProviderName, apiKey: string) {
	await settingstoreAPI.setAISettingAPIKey(providerName, apiKey);
	await providerSetAPI.setProviderAPIKey(providerName, apiKey);
}

export async function SetAISettingAttrs(providerName: ProviderName, aiSettingAttrs: AISettingAttrs) {
	await settingstoreAPI.setAISettingAttrs(providerName, aiSettingAttrs);
	if (typeof aiSettingAttrs.origin !== 'undefined' || typeof aiSettingAttrs.chatCompletionPathPrefix !== 'undefined') {
		await providerSetAPI.setProviderAttribute(
			providerName,
			aiSettingAttrs.origin,
			aiSettingAttrs.chatCompletionPathPrefix
		);
	}
}

function mergeDefaultsModelSettingAndInbuilt(
	providerName: ProviderName,
	modelName: ModelName,
	inbuiltProviderModels: Record<ProviderName, Record<ModelName, ModelParams>>,
	modelSetting: ModelSetting
): ModelParams {
	let inbuiltModelParams: ModelParams | undefined = undefined;
	if (providerName in inbuiltProviderModels && modelName in inbuiltProviderModels[providerName]) {
		inbuiltModelParams = inbuiltProviderModels[providerName][modelName];
	}

	let temperature = DefaultModelParams.temperature ?? 0.1;
	if (typeof modelSetting.temperature !== 'undefined') {
		temperature = modelSetting.temperature;
	} else if (inbuiltModelParams) {
		temperature = inbuiltModelParams.temperature ?? 0.1;
	}

	let stream = DefaultModelParams.stream;
	if (typeof modelSetting.stream !== 'undefined') {
		stream = modelSetting.stream;
	} else if (inbuiltModelParams) {
		stream = inbuiltModelParams.stream;
	}

	let maxPromptLength = DefaultModelParams.maxPromptLength;
	if (typeof modelSetting.maxPromptLength !== 'undefined') {
		maxPromptLength = modelSetting.maxPromptLength;
	} else if (inbuiltModelParams) {
		maxPromptLength = inbuiltModelParams.maxPromptLength;
	}

	let maxOutputLength = DefaultModelParams.maxOutputLength;
	if (typeof modelSetting.maxOutputLength !== 'undefined') {
		maxOutputLength = modelSetting.maxOutputLength;
	} else if (inbuiltModelParams) {
		maxOutputLength = inbuiltModelParams.maxOutputLength;
	}

	let reasoningSupport = DefaultModelParams.reasoningSupport;
	if (typeof modelSetting.reasoningSupport !== 'undefined') {
		reasoningSupport = modelSetting.reasoningSupport;
	} else if (inbuiltModelParams) {
		reasoningSupport = inbuiltModelParams.reasoningSupport;
	}

	let systemPrompt = DefaultModelParams.systemPrompt;
	if (typeof modelSetting.systemPrompt !== 'undefined') {
		systemPrompt = modelSetting.systemPrompt;
	} else if (inbuiltModelParams) {
		systemPrompt = inbuiltModelParams.systemPrompt;
	}

	let timeout = DefaultModelParams.timeout;
	if (typeof modelSetting.timeout !== 'undefined') {
		timeout = modelSetting.timeout;
	} else if (inbuiltModelParams) {
		timeout = inbuiltModelParams.timeout;
	}

	let additionalParameters = { ...DefaultModelParams.additionalParameters };
	if (modelSetting.additionalParameters) {
		additionalParameters = { ...additionalParameters, ...modelSetting.additionalParameters };
	}

	return {
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
		const inbuiltProviderModels = info.inbuiltProviderModels;

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
					const mergedModelParam = mergeDefaultsModelSettingAndInbuilt(
						providerName,
						modelName,
						inbuiltProviderModels,
						modelSetting
					);
					const modelOption: ModelOption = {
						title: modelSetting.displayName,
						provider: providerName,
						name: modelName,
						stream: mergedModelParam.stream,
						maxPromptLength: mergedModelParam.maxPromptLength,
						maxOutputLength: mergedModelParam.maxOutputLength,
						temperature: mergedModelParam.temperature,
						reasoningSupport: mergedModelParam.reasoningSupport,
						systemPrompt: mergedModelParam.systemPrompt,
						timeout: mergedModelParam.timeout,
						additionalParameters: mergedModelParam.additionalParameters,
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
		return { allOptions: [DefaultModelOption], default: DefaultModelOption };
	}
}

// Create a Model setting from default or inbuilt ModelParams if available
export async function PopulateModelSettingDefaults(
	providerName: ProviderName,
	modelName: ModelName,
	existingData?: ModelSetting
): Promise<ModelSetting> {
	const info = await providerSetAPI.getConfigurationInfo();
	if (
		info.defaultProvider === '' ||
		Object.keys(info.configuredProviders).length === 0 ||
		!(providerName in info.configuredProviders)
	) {
		return DefaultModelSetting;
	}

	let modelSetting = existingData;
	if (typeof modelSetting === 'undefined') {
		modelSetting = {
			displayName: modelName,
			isEnabled: true,
		};
	}

	const mergedModelParam = mergeDefaultsModelSettingAndInbuilt(
		providerName,
		modelName,
		info.inbuiltProviderModels,
		modelSetting
	);

	return {
		displayName: modelSetting.displayName,
		isEnabled: modelSetting.isEnabled,
		stream: mergedModelParam.stream,
		maxPromptLength: mergedModelParam.maxPromptLength,
		maxOutputLength: mergedModelParam.maxOutputLength,
		temperature: mergedModelParam.temperature,
		reasoningSupport: mergedModelParam.reasoningSupport,
		systemPrompt: mergedModelParam.systemPrompt,
		timeout: mergedModelParam.timeout,
	};
}
