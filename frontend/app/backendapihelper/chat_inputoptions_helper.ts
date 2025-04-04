import { providerSetAPI, settingstoreAPI } from '@/backendapibase';
import type { ModelName, ProviderInfo, ProviderName } from '@/models/aiprovidermodel';

export interface ModelOption {
	title: string;
	provider: ProviderName;
	name: ModelName;
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
		if (!('configuredProviders' in configInfo) || !('defaultProvider' in configInfo)) {
			return handleNoConfiguredModels();
		}

		// Extract default provider and initialize provider info dictionary
		const configDefaultProvider = configInfo['defaultProvider'] as ProviderName;
		const providerInfoDict = createProviderInfoDict(configInfo['configuredProviders']);

		for (const providerName of Object.keys(providerInfoDict)) {
			const aiSetting = settings.aiSettings[providerName];
			if (aiSetting.isEnabled) {
				const settingsDefaultModelName = aiSetting.defaultModel;
				for (const modelParams of aiSetting.modelSettings) {
					if (!modelParams.isEnabled) {
						continue;
					}
					const modelOption = {
						title: modelParams.displayName,
						provider: providerName,
						temperature: modelParams.temperature || 0.1,
						name: modelParams.name,
					};
					inputModels.push(modelOption);
					if (modelParams.name === settingsDefaultModelName && providerName === configDefaultProvider) {
						defaultOption = modelOption;
					}
				}
			}
		}

		if (defaultOption === undefined || inputModels.length === 0) {
			throw Error('No input option found !!!');
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
		provider: 'No Provider',
		name: 'No Model',
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
