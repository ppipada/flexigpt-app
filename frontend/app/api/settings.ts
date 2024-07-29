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
};

export const providers = ['openai', 'anthropic', 'huggingface', 'googlegl', 'llamacpp'] as const;

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
};

export const aiSettingsDescriptions: Record<string, string> = {
	'openai.apiKey': 'Your openAI API key. Can be seen at https://beta.openai.com/account/api-keys',
	'openai.defaultModel': 'Default model to use for chat requests',
	'openai.defaultOrigin':
		'Default origin to use for requests. This can be used to talk to any server that serves a compatible API',
	'openai.defaultTemperature': 'Default temperature setting for chat requests',
	'openai.additionalSettings': 'Any additional settings to pass to the model. Input as a JSON object',

	'anthropic.apiKey': 'Your anthropic API key.',
	'anthropic.defaultModel': 'Default model to use for chat requests',
	'anthropic.defaultOrigin':
		'Default origin to use for requests. This can be used to talk to any server that serves a compatible API',
	'anthropic.defaultTemperature': 'Default temperature setting for chat requests',
	'anthropic.additionalSettings': 'Any additional settings to pass to the model. Input as a JSON object',

	'huggingface.apiKey': 'Your huggingface API key.',
	'huggingface.defaultModel': 'Default model to use for chat completion requests',
	'huggingface.defaultOrigin':
		'Default origin to use for requests. This can be used to talk to any server that serves a compatible API',
	'huggingface.defaultTemperature': 'Default temperature setting for chat requests',
	'huggingface.additionalSettings': 'Any additional settings to pass to the model. Input as a JSON object',

	'googlegl.apiKey': 'Your googlegl API key.',
	'googlegl.defaultModel': 'Default model to use for chat completion requests',
	'googlegl.defaultOrigin':
		'Default origin to use for requests. This can be used to talk to any server that serves a compatible API',
	'googlegl.defaultTemperature': 'Default temperature setting for chat requests',
	'googlegl.additionalSettings': 'Any additional settings to pass to the model. Input as a JSON object',

	'llamacpp.apiKey': 'Your llamacpp API key.',
	'llamacpp.defaultModel': 'Default model to use for chat completion requests',
	'llamacpp.defaultOrigin':
		'Default origin to use for requests. This can be used to talk to any server that serves a compatible API',
	'llamacpp.defaultTemperature': 'Default temperature setting for chat requests',
	'llamacpp.additionalSettings': 'Any additional settings to pass to the model. Input as a JSON object',
};

export async function getAllSettings(): Promise<SettingsSchema> {
	return await window.SettingsAPI.getAllSettings();
}

export async function setSetting(key: string, value: any) {
	await window.SettingsAPI.setSetting(key, value);
}
