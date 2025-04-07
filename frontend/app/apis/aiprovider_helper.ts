import { providerSetAPI } from '@/apis/baseapi';
import type { ProviderInfo, ProviderName } from '@/models/aiprovidermodel';

// Helper function to create a dictionary of provider info
export async function CreateProviderInfoDict() {
	const configInfo = await providerSetAPI.getConfigurationInfo();
	// Validate the presence of necessary configuration properties
	if (!('configuredProviders' in configInfo) || !('defaultProvider' in configInfo)) {
		return;
	}
	// Extract default provider and initialize provider info dictionary
	const configDefaultProvider = configInfo['defaultProvider'] as ProviderName;

	const providerInfoDict: Record<ProviderName, ProviderInfo> = {};
	for (const providerInfo of configInfo['configuredProviders'] as ProviderInfo[]) {
		providerInfoDict[providerInfo.name] = providerInfo;
	}
	return { defaultProvider: configDefaultProvider, providerInfo: providerInfoDict };
}
