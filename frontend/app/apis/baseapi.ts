import type { IModelPresetStoreAPI } from '@/models/aimodelmodel';
import type { IProviderSetAPI } from '@/models/aiprovidermodel';
import type { IBackendAPI } from '@/models/backendmodel';
import type { IConversationStoreAPI } from '@/models/conversationmodel';
import type { ILogger } from '@/models/loggermodel';
import type { ISettingStoreAPI } from '@/models/settingmodel';

import { IS_WAILS_PLATFORM } from '@/lib/features';

// Static imports at the top - Vite will tree-shake unused imports
import * as wailsImpl from './wailsapi';

export let log: ILogger;
export let backendAPI: IBackendAPI;
export let conversationStoreAPI: IConversationStoreAPI;
export let providerSetAPI: IProviderSetAPI;
export let settingstoreAPI: ISettingStoreAPI;
export let modelPresetStoreAPI: IModelPresetStoreAPI;

// Conditional initialization
if (IS_WAILS_PLATFORM) {
	// Initialize with Wails implementations
	log = new wailsImpl.WailsLogger();
	backendAPI = new wailsImpl.WailsBackendAPI();
	conversationStoreAPI = new wailsImpl.WailsConversationStoreAPI();
	providerSetAPI = new wailsImpl.WailsProviderSetAPI();
	settingstoreAPI = new wailsImpl.WailsSettingStoreAPI();
	modelPresetStoreAPI = new wailsImpl.WailsModelPresetStoreAPI();
} else {
	// Error for unsupported platforms
	throw new Error('Unsupported platform');
}
