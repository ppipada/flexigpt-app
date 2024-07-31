import {
	anthropicProviderInfo,
	googleProviderInfo,
	huggingfaceProviderInfo,
	llamacppProviderInfo,
	openaiProviderInfo,
	ProviderName,
} from 'aiprovider';
import { SettingsSchema } from './types';

export const defaultSettingsData: SettingsSchema = {
	app: {
		defaultProvider: ProviderName.OPENAI,
	},
	openai: {
		apiKey: openaiProviderInfo.apiKey,
		defaultModel: openaiProviderInfo.defaultModel,
		defaultOrigin: openaiProviderInfo.defaultOrigin,
		defaultTemperature: openaiProviderInfo.defaultTemperature,
		additionalSettings: openaiProviderInfo.additionalSettings,
	},
	anthropic: {
		apiKey: anthropicProviderInfo.apiKey,
		defaultModel: anthropicProviderInfo.defaultModel,
		defaultOrigin: anthropicProviderInfo.defaultOrigin,
		defaultTemperature: anthropicProviderInfo.defaultTemperature,
		additionalSettings: anthropicProviderInfo.additionalSettings,
	},
	huggingface: {
		apiKey: huggingfaceProviderInfo.apiKey,
		defaultModel: huggingfaceProviderInfo.defaultModel,
		defaultOrigin: huggingfaceProviderInfo.defaultOrigin,
		defaultTemperature: huggingfaceProviderInfo.defaultTemperature,
		additionalSettings: huggingfaceProviderInfo.additionalSettings,
	},
	googlegl: {
		apiKey: googleProviderInfo.apiKey,
		defaultModel: googleProviderInfo.defaultModel,
		defaultOrigin: googleProviderInfo.defaultOrigin,
		defaultTemperature: googleProviderInfo.defaultTemperature,
		additionalSettings: googleProviderInfo.additionalSettings,
	},
	llamacpp: {
		apiKey: llamacppProviderInfo.apiKey,
		defaultModel: llamacppProviderInfo.defaultModel,
		defaultOrigin: llamacppProviderInfo.defaultOrigin,
		defaultTemperature: llamacppProviderInfo.defaultTemperature,
		additionalSettings: llamacppProviderInfo.additionalSettings,
	},
};
