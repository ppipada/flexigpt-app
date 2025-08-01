import type { ModelParams } from '@/spec/aiprovider';
import {
	type ModelDisplayName,
	type ModelPreset,
	type ModelPresetID,
	type ProviderDisplayName,
	type ProviderName,
	ReasoningLevel,
	ReasoningType,
} from '@/spec/modelpreset';
import { AuthKeyTypeProvider, type SettingsSchema } from '@/spec/setting';

import { modelPresetStoreAPI, settingstoreAPI } from '@/apis/baseapi';
import { getAllProviderPresetsMap } from '@/apis/modelpreset_helper';

export interface ChatOption extends ModelParams {
	providerName: ProviderName;
	modelPresetID: ModelPresetID;
	providerDisplayName: ProviderDisplayName;
	modelDisplayName: ModelDisplayName;
	disablePreviousMessages: boolean;
}

const DefaultModelParams: ModelParams = {
	name: '',
	stream: false,
	maxPromptLength: 2048,
	maxOutputLength: 1024,
	temperature: 0.1,
	reasoning: {
		type: ReasoningType.SingleWithLevels,
		level: ReasoningLevel.Medium,
		tokens: 1024,
	},
	systemPrompt: '',
	timeout: 300,
	additionalParametersRawJSON: undefined,
};

export const DefaultChatOptions: ChatOption = {
	...DefaultModelParams,
	providerName: 'no-provider',
	modelPresetID: 'no-model',
	providerDisplayName: 'No Provider',
	modelDisplayName: 'No Model configured',
	disablePreviousMessages: false,
};

function hasApiKey(settings: SettingsSchema, providerName: ProviderName): boolean {
	return settings.authKeys.some(k => k.type === AuthKeyTypeProvider && k.keyName === providerName && k.nonEmpty);
}

//  Ensure every optional field of `ModelParams` is filled.
//  Falls back to hard-coded defaults when the model preset does not specify a value.
function buildModelParams(modelPreset: ModelPreset): ModelParams {
	const o: ModelParams = {
		name: modelPreset.name,
		stream: modelPreset.stream ?? DefaultModelParams.stream,
		maxPromptLength: modelPreset.maxPromptLength ?? DefaultModelParams.maxPromptLength,
		maxOutputLength: modelPreset.maxOutputLength ?? DefaultModelParams.maxOutputLength,
		systemPrompt: modelPreset.systemPrompt ?? DefaultModelParams.systemPrompt,
		timeout: modelPreset.timeout ?? DefaultModelParams.timeout,

		// Optional fields in modelparams
		temperature: modelPreset.temperature,
		reasoning: modelPreset.reasoning,
		additionalParametersRawJSON: modelPreset.additionalParametersRawJSON,
	};
	if (o.temperature === undefined && o.reasoning === undefined) {
		o.temperature = DefaultModelParams.temperature;
	}
	return o;
}

export async function getChatInputOptions(): Promise<{
	allOptions: ChatOption[];
	default: ChatOption;
}> {
	try {
		/* fetch everything in parallel */
		const [allProviderPresets, settings, defaultProviderName] = await Promise.all([
			getAllProviderPresetsMap(), // contains built-ins + user presets merged
			settingstoreAPI.getSettings(),
			modelPresetStoreAPI.getDefaultProvider(),
		]);

		const allOptions: ChatOption[] = [];
		let defaultOption: ChatOption | undefined;

		for (const [providerName, providerPreset] of Object.entries(allProviderPresets)) {
			/* provider disabled or no key → skip */
			if (!providerPreset.isEnabled || !hasApiKey(settings, providerName)) {
				continue;
			}

			for (const [modelPresetID, modelPreset] of Object.entries(providerPreset.modelPresets)) {
				if (!modelPreset.isEnabled) continue;

				const modelParams = buildModelParams(modelPreset);

				const option: ChatOption = {
					...modelParams,
					providerName: providerName,
					modelPresetID: modelPresetID,
					providerDisplayName: providerPreset.displayName,
					modelDisplayName: modelPreset.displayName,
					disablePreviousMessages: false,
				};

				allOptions.push(option);

				if (providerName === defaultProviderName && modelPresetID === providerPreset.defaultModelPresetID) {
					defaultOption = option;
				}
			}
		}

		if (!defaultOption) {
			if (allOptions.length > 0) {
				defaultOption = allOptions[0];
			} else {
				defaultOption = DefaultChatOptions;
				allOptions.push(DefaultChatOptions);
			}
		}

		return { allOptions, default: defaultOption };
	} catch (error) {
		console.error('Error while building chat input options:', error);
		return { allOptions: [DefaultChatOptions], default: DefaultChatOptions };
	}
}
