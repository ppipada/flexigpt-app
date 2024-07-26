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
};

async function getAllSettings(): Promise<SettingsSchema> {
	return await window.SettingsAPI.getAllSettings();
}

async function setSetting(key: string, value: any) {
	await window.SettingsAPI.setSetting(key, value);
}

export { getAllSettings, setSetting };
