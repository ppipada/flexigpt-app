import { providerSetAPI, settingstoreAPI } from '@/backendapibase';
import { ModelInfo, ModelName, ProviderInfo, ProviderName } from '@/models/aiprovidermodel';
import { SettingsSchema } from '@/models/settingmodel';

export interface ModelOption {
	title: string;
	provider?: ProviderName;
	name?: ModelName;
	temperature: number;
}

export async function GetChatInputOptions() {
	try {
		// Fetch configuration info and settings
		const configInfo = await providerSetAPI.getConfigurationInfo();
		const settings = await settingstoreAPI.getAllSettings();

		// Initialize default option and input models array
		let defaultOption: ModelOption | undefined;
		const inputModels: ModelOption[] = [];

		// Validate the presence of necessary configuration properties
		if (
			!('configuredModels' in configInfo) ||
			!('configuredProviders' in configInfo) ||
			!('defaultProvider' in configInfo)
		) {
			return handleNoConfiguredModels();
		}

		// Extract default provider and initialize provider info dictionary
		const configDefaultProvider = configInfo['defaultProvider'] as ProviderName;
		const providerInfoDict = createProviderInfoDict(configInfo['configuredProviders']);

		// Process configured models
		const configuredModels = configInfo['configuredModels'] as ModelInfo[];
		for (const modelInfo of configuredModels) {
			const modelOption = createModelOption(modelInfo, providerInfoDict, settings);
			if (modelOption) {
				inputModels.push(modelOption);
				if (modelInfo.name === providerInfoDict[configDefaultProvider]?.defaultModel) {
					defaultOption = modelOption;
				}
			}
		}

		return { allOptions: inputModels, default: defaultOption };
	} catch (error) {
		console.error('Error fetching chat input options:', error);
		return handleNoConfiguredModels();
	}
}

// Helper function to handle cases with no configured models
function handleNoConfiguredModels() {
	const noModel: ModelOption = {
		title: 'No Model configured',
		temperature: 0.1,
	};
	return { allOptions: [noModel], default: noModel };
}

// Helper function to create a dictionary of provider info
function createProviderInfoDict(configuredProviders: ProviderInfo[]) {
	const providerInfoDict: Record<string, ProviderInfo> = {};
	for (const providerInfo of configuredProviders) {
		providerInfoDict[providerInfo.name] = providerInfo;
	}
	return providerInfoDict;
}

// Helper function to create a model option if the provider is enabled
function createModelOption(
	modelInfo: ModelInfo,
	providerInfoDict: Record<string, ProviderInfo>,
	settings: SettingsSchema
): ModelOption | null {
	const providerName = modelInfo.provider;
	const aiSetting = settings.aiSettings[providerName];

	if (aiSetting && aiSetting.isEnabled) {
		return {
			title: modelInfo.displayName,
			provider: providerName,
			temperature: providerInfoDict[providerName]?.defaultTemperature || 0.1,
			name: modelInfo.name,
		};
	}
	return null;
}
