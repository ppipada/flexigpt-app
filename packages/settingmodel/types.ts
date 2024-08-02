export type SettingsSchema = {
	app: {
		defaultProvider: string;
	};
	openai: {
		apiKey: string;
		defaultModel: string;
		defaultTemperature: number;
		defaultOrigin: string;
		additionalSettings: Record<string, any>;
	};
	anthropic: {
		apiKey: string;
		defaultModel: string;
		defaultTemperature: number;
		defaultOrigin: string;
		additionalSettings: Record<string, any>;
	};
	huggingface: {
		apiKey: string;
		defaultModel: string;
		defaultTemperature: number;
		defaultOrigin: string;
		additionalSettings: Record<string, any>;
	};
	google: {
		apiKey: string;
		defaultModel: string;
		defaultTemperature: number;
		defaultOrigin: string;
		additionalSettings: Record<string, any>;
	};
	llamacpp: {
		apiKey: string;
		defaultModel: string;
		defaultTemperature: number;
		defaultOrigin: string;
		additionalSettings: Record<string, any>;
	};
};

export interface ISettingsAPI {
	getAllSettings: () => Promise<SettingsSchema>;
	setSetting: (key: string, value: any) => Promise<void>;
}
