import { SecureSchema } from './secureSchema';

export type SettingsSchema = {
	openai: {
		apiKey: string;
		defaultModel: string;
		defaultTemperature: number;
	};
	anthropic: {
		apiKey: string;
		defaultModel: string;
		defaultTemperature: number;
	};
} & SecureSchema;

export const defaultSettingsData: SettingsSchema = {
	openai: {
		apiKey: '',
		defaultModel: 'gpt-4o-mini',
		defaultTemperature: 0.1,
	},
	anthropic: {
		apiKey: '',
		defaultModel: 'claude-3-haiku-20240307',
		defaultTemperature: 0.1,
	},
	_sensitiveKeys: ['openai.apiKey', 'anthropic.apiKey'],
};
