import { SecureSchema } from './secureSchema';

export type SettingsSchema = {
	app: {
		defaultProvider: string;
	};
	openai: {
		apiKey: string;
		defaultModel: string;
		defaultTemperature: number;
		defaultOrigin: string;
		additionalSettings: string;
	};
	anthropic: {
		apiKey: string;
		defaultModel: string;
		defaultTemperature: number;
		defaultOrigin: string;
		additionalSettings: string;
	};
	huggingface: {
		apiKey: string;
		defaultModel: string;
		defaultTemperature: number;
		defaultOrigin: string;
		additionalSettings: string;
	};
	googlegl: {
		apiKey: string;
		defaultModel: string;
		defaultTemperature: number;
		defaultOrigin: string;
		additionalSettings: string;
	};
	llamacpp: {
		apiKey: string;
		defaultModel: string;
		defaultTemperature: number;
		defaultOrigin: string;
		additionalSettings: string;
	};
} & SecureSchema;

export const defaultSettingsData: SettingsSchema = {
	app: {
		defaultProvider: 'openai',
	},
	openai: {
		apiKey: '',
		defaultModel: 'gpt-4o-mini',
		defaultOrigin: 'https://api.openai.com',
		defaultTemperature: 0.1,
		additionalSettings: '{}',
	},
	anthropic: {
		apiKey: '',
		defaultModel: 'claude-3-haiku-20240307',
		defaultOrigin: 'https://api.anthropic.com',
		defaultTemperature: 0.1,
		additionalSettings: '{}',
	},
	huggingface: {
		apiKey: '',
		defaultModel: 'deepseek-ai/deepseek-coder-1.3b-instruct',
		defaultOrigin: 'https://api-inference.huggingface.co',
		defaultTemperature: 0.1,
		additionalSettings: '{}',
	},
	googlegl: {
		apiKey: '',
		defaultModel: 'gemini-1.0-pro',
		defaultOrigin: 'https://generativelanguage.googleapis.com',
		defaultTemperature: 0.1,
		additionalSettings: '{}',
	},
	llamacpp: {
		apiKey: '',
		defaultModel: '',
		defaultOrigin: 'http://127.0.0.1:8080',
		defaultTemperature: 0.1,
		additionalSettings: '{}',
	},

	_sensitiveKeys: ['openai.apiKey', 'anthropic.apiKey', 'huggingface.apiKey', 'googlegl.apiKey', 'llamacpp.apiKey'],
};
