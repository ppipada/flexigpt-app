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

export interface IBackendAPI {
	ping: () => Promise<string>;
}

export interface ISettingsAPI {
	getAllSettings: () => Promise<SettingsSchema>;
	setSetting: (key: string, value: any) => Promise<void>;
}
