import {
	anthropicProviderInfo,
	googleProviderInfo,
	huggingfaceProviderInfo,
	llamacppProviderInfo,
	openaiProviderInfo,
	ProviderName,
} from 'aiprovider';
import { SettingsSchema } from './types';

export const defaultAISettings = {
	[ProviderName.ANTHROPIC]: {
		apiKey: anthropicProviderInfo.apiKey,
		defaultModel: anthropicProviderInfo.defaultModel,
		defaultOrigin: anthropicProviderInfo.defaultOrigin,
		defaultTemperature: anthropicProviderInfo.defaultTemperature,
		additionalSettings: anthropicProviderInfo.additionalSettings,
	},
	[ProviderName.GOOGLE]: {
		apiKey: googleProviderInfo.apiKey,
		defaultModel: googleProviderInfo.defaultModel,
		defaultOrigin: googleProviderInfo.defaultOrigin,
		defaultTemperature: googleProviderInfo.defaultTemperature,
		additionalSettings: googleProviderInfo.additionalSettings,
	},
	[ProviderName.HUGGINGFACE]: {
		apiKey: huggingfaceProviderInfo.apiKey,
		defaultModel: huggingfaceProviderInfo.defaultModel,
		defaultOrigin: huggingfaceProviderInfo.defaultOrigin,
		defaultTemperature: huggingfaceProviderInfo.defaultTemperature,
		additionalSettings: huggingfaceProviderInfo.additionalSettings,
	},
	[ProviderName.LLAMACPP]: {
		apiKey: llamacppProviderInfo.apiKey,
		defaultModel: llamacppProviderInfo.defaultModel,
		defaultOrigin: llamacppProviderInfo.defaultOrigin,
		defaultTemperature: llamacppProviderInfo.defaultTemperature,
		additionalSettings: llamacppProviderInfo.additionalSettings,
	},
	[ProviderName.OPENAI]: {
		apiKey: openaiProviderInfo.apiKey,
		defaultModel: openaiProviderInfo.defaultModel,
		defaultOrigin: openaiProviderInfo.defaultOrigin,
		defaultTemperature: openaiProviderInfo.defaultTemperature,
		additionalSettings: openaiProviderInfo.additionalSettings,
	},
};

export const defaultSettingsData: SettingsSchema = {
	app: {
		defaultProvider: ProviderName.OPENAI,
	},
	aiSettings: defaultAISettings,
};

export const sensitiveKeys = [
	'openai.apiKey',
	'anthropic.apiKey',
	'huggingface.apiKey',
	'google.apiKey',
	'llamacpp.apiKey',
];
