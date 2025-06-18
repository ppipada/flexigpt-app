import {
	type ChatOptions,
	DefaultChatOptions,
	DefaultModelParams,
	DefaultModelPreset,
	type ModelParams,
	type ModelPreset,
	type ModelPresetID,
	type ProviderName,
	type ProviderPreset,
	type ReasoningParams,
} from '@/models/modelpresetsmodel';

import { modelPresetStoreAPI, providerSetAPI, settingstoreAPI } from '@/apis/baseapi';

function mergeDefaultsModelPresetAndInbuilt(
	providerName: ProviderName,
	modelPresetID: ModelPresetID,
	inbuiltProviderPresets: Record<ProviderName, ProviderPreset>,
	modelPreset: ModelPreset
): ModelParams {
	let inbuiltModelPreset: ModelPreset | undefined = undefined;
	if (providerName in inbuiltProviderPresets && modelPresetID in inbuiltProviderPresets[providerName].modelPresets) {
		inbuiltModelPreset = inbuiltProviderPresets[providerName].modelPresets[modelPresetID];
	}

	let temperature = DefaultModelParams.temperature ?? 0.1;
	if (typeof modelPreset.temperature !== 'undefined') {
		temperature = modelPreset.temperature;
	} else if (inbuiltModelPreset && inbuiltModelPreset.temperature) {
		temperature = inbuiltModelPreset.temperature;
	}

	let stream = DefaultModelParams.stream;
	if (typeof modelPreset.stream !== 'undefined') {
		stream = modelPreset.stream;
	} else if (inbuiltModelPreset && inbuiltModelPreset.stream) {
		stream = inbuiltModelPreset.stream;
	}

	let maxPromptLength = DefaultModelParams.maxPromptLength;
	if (typeof modelPreset.maxPromptLength !== 'undefined') {
		maxPromptLength = modelPreset.maxPromptLength;
	} else if (inbuiltModelPreset && inbuiltModelPreset.maxPromptLength) {
		maxPromptLength = inbuiltModelPreset.maxPromptLength;
	}

	let maxOutputLength = DefaultModelParams.maxOutputLength;
	if (typeof modelPreset.maxOutputLength !== 'undefined') {
		maxOutputLength = modelPreset.maxOutputLength;
	} else if (inbuiltModelPreset && inbuiltModelPreset.maxOutputLength) {
		maxOutputLength = inbuiltModelPreset.maxOutputLength;
	}

	let reasoning: ReasoningParams | undefined = undefined;
	if (typeof modelPreset.reasoning !== 'undefined') {
		reasoning = modelPreset.reasoning;
	} else if (inbuiltModelPreset) {
		reasoning = inbuiltModelPreset.reasoning;
	}

	let systemPrompt = DefaultModelParams.systemPrompt;
	if (typeof modelPreset.systemPrompt !== 'undefined') {
		systemPrompt = modelPreset.systemPrompt;
	} else if (inbuiltModelPreset && inbuiltModelPreset.systemPrompt) {
		systemPrompt = inbuiltModelPreset.systemPrompt;
	}

	let timeout = DefaultModelParams.timeout;
	if (typeof modelPreset.timeout !== 'undefined') {
		timeout = modelPreset.timeout;
	} else if (inbuiltModelPreset && inbuiltModelPreset.timeout) {
		timeout = inbuiltModelPreset.timeout;
	}

	let additionalParametersRawJSON = DefaultModelParams.additionalParametersRawJSON;
	if (modelPreset.additionalParametersRawJSON) {
		additionalParametersRawJSON = modelPreset.additionalParametersRawJSON;
	}

	return {
		name: modelPreset.name,
		stream: stream,
		maxPromptLength: maxPromptLength,
		maxOutputLength: maxOutputLength,
		temperature: temperature,
		reasoning: reasoning,
		systemPrompt: systemPrompt,
		timeout: timeout,
		additionalParametersRawJSON: additionalParametersRawJSON,
	};
}

export async function GetChatInputOptions(): Promise<{ allOptions: ChatOptions[]; default: ChatOptions }> {
	try {
		// Fetch configuration info and settings
		const info = await providerSetAPI.getConfigurationInfo();
		if (info.defaultProvider === '' || Object.keys(info.configuredProviders).length === 0) {
			return { allOptions: [DefaultChatOptions], default: DefaultChatOptions };
		}
		const configDefaultProvider = info.defaultProvider;
		const providerInfoDict = info.configuredProviders;
		const inbuiltProviderPresets = info.inbuiltProviderModels;

		const onlySettings = await settingstoreAPI.getAllSettings();
		const modelPresetsSchema = await modelPresetStoreAPI.getAllModelPresets();
		const mergedPresets = MergeInbuiltModelsWithPresets(modelPresetsSchema.providerPresets, inbuiltProviderPresets);
		// Initialize default option and input models array
		let defaultOption: ChatOptions | undefined;
		const inputModels: ChatOptions[] = [];

		for (const providerName of Object.keys(providerInfoDict)) {
			if (
				!(providerName in onlySettings.aiSettings) ||
				!onlySettings.aiSettings[providerName].isEnabled ||
				!(providerName in mergedPresets)
			) {
				// Skip disabled providers or ones without any data in mergedPresets.
				continue;
			}

			const settingsDefaultModelPresetID = mergedPresets[providerName].defaultModelPresetID;

			for (const [modelPresetID, modelPreset] of Object.entries(mergedPresets[providerName].modelPresets)) {
				if (modelPresetID === '') {
					console.warn('Got empty model preset id in merged models. skipping. provider: ', providerName);
					continue;
				}
				// console.log('Processing', JSON.stringify(modelPreset, null, 2));
				if (!modelPreset.isEnabled) {
					continue;
				}
				const mergedModelParam = mergeDefaultsModelPresetAndInbuilt(
					providerName,
					modelPresetID,
					inbuiltProviderPresets,
					modelPreset
				);
				const chatOption: ChatOptions = {
					id: modelPresetID,
					title: modelPreset.displayName,
					provider: providerName,
					name: mergedModelParam.name,
					stream: mergedModelParam.stream,
					maxPromptLength: mergedModelParam.maxPromptLength,
					maxOutputLength: mergedModelParam.maxOutputLength,
					temperature: mergedModelParam.temperature,
					reasoning: mergedModelParam.reasoning,
					systemPrompt: mergedModelParam.systemPrompt,
					timeout: mergedModelParam.timeout,
					additionalParametersRawJSON: mergedModelParam.additionalParametersRawJSON,
					disablePreviousMessages: false,
				};
				inputModels.push(chatOption);
				if (modelPreset.id === settingsDefaultModelPresetID && providerName === configDefaultProvider) {
					defaultOption = chatOption;
				}
			}
		}

		if (defaultOption === undefined || inputModels.length === 0) {
			throw Error('No input option found !!!');
		}
		return { allOptions: inputModels, default: defaultOption };
	} catch (error) {
		console.error('Error fetching chat input options:', error);
		return { allOptions: [DefaultChatOptions], default: DefaultChatOptions };
	}
}

// Create a Model setting from default or inbuilt ModelParams if available
export async function PopulateModelPresetDefaults(
	providerName: ProviderName,
	modelPresetID: ModelPresetID,
	existingData?: ModelPreset
): Promise<ModelPreset> {
	const info = await providerSetAPI.getConfigurationInfo();
	if (
		info.defaultProvider === '' ||
		Object.keys(info.configuredProviders).length === 0 ||
		!(providerName in info.configuredProviders)
	) {
		return DefaultModelPreset;
	}

	let modelPreset = existingData;
	if (typeof modelPreset === 'undefined') {
		modelPreset = {
			id: modelPresetID,
			name: modelPresetID,
			displayName: modelPresetID,
			isEnabled: true,
			shortCommand: modelPresetID,
		};
	}

	const mergedModelParam = mergeDefaultsModelPresetAndInbuilt(
		providerName,
		modelPresetID,
		info.inbuiltProviderModels,
		modelPreset
	);

	return {
		id: modelPreset.name,
		name: modelPreset.name,
		displayName: modelPreset.displayName,
		isEnabled: modelPreset.isEnabled,
		shortCommand: modelPreset.shortCommand,
		stream: mergedModelParam.stream,
		maxPromptLength: mergedModelParam.maxPromptLength,
		maxOutputLength: mergedModelParam.maxOutputLength,
		temperature: mergedModelParam.temperature,
		reasoning: mergedModelParam.reasoning,
		systemPrompt: mergedModelParam.systemPrompt,
		timeout: mergedModelParam.timeout,
	};
}

export function MergeInbuiltModelsWithPresets(
	modelPresets: Record<ProviderName, ProviderPreset>,
	inbuiltProviderPresets: Record<ProviderName, ProviderPreset>
): Record<ProviderName, ProviderPreset> {
	const newPresets = { ...modelPresets };

	for (const provider in inbuiltProviderPresets) {
		if (!(provider in newPresets)) {
			continue;
		}

		// For each model in inbuiltProviderInfo, ensure it exists in modelPresets
		for (const modelPresetID in inbuiltProviderPresets[provider].modelPresets) {
			if (modelPresetID === '') {
				console.warn('Got empty model preset id in inbuilt models. skipping. provider: ', provider);
				continue;
			}
			if (modelPresetID in newPresets[provider]) {
				continue;
			}

			newPresets[provider].modelPresets[modelPresetID] = {
				id: inbuiltProviderPresets[provider].modelPresets[modelPresetID].id,
				name: inbuiltProviderPresets[provider].modelPresets[modelPresetID].name,
				displayName: inbuiltProviderPresets[provider].modelPresets[modelPresetID].displayName,
				isEnabled: inbuiltProviderPresets[provider].modelPresets[modelPresetID].isEnabled,
				shortCommand: inbuiltProviderPresets[provider].modelPresets[modelPresetID].shortCommand,
				stream: inbuiltProviderPresets[provider].modelPresets[modelPresetID].stream,
				maxPromptLength: inbuiltProviderPresets[provider].modelPresets[modelPresetID].maxPromptLength,
				maxOutputLength: inbuiltProviderPresets[provider].modelPresets[modelPresetID].maxOutputLength,
				temperature: inbuiltProviderPresets[provider].modelPresets[modelPresetID].temperature,
				reasoning: inbuiltProviderPresets[provider].modelPresets[modelPresetID].reasoning,
				systemPrompt: inbuiltProviderPresets[provider].modelPresets[modelPresetID].systemPrompt,
				timeout: inbuiltProviderPresets[provider].modelPresets[modelPresetID].timeout,
				additionalParametersRawJSON:
					inbuiltProviderPresets[provider].modelPresets[modelPresetID].additionalParametersRawJSON,
			};
		}
	}

	// console.log('Inbuilt', JSON.stringify(inbuiltProviderPresets, null, 2));
	// console.log('New', JSON.stringify(newPresets, null, 2));

	return newPresets;
}
