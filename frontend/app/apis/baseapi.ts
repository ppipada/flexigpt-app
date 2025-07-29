import type { IProviderSetAPI } from '@/models/aiprovidermodel';
import type { IBackendAPI } from '@/models/backendmodel';
import type { IConversationStoreAPI } from '@/models/conversationmodel';
import type { ILogger } from '@/models/loggermodel';
import type { IModelPresetStoreAPI } from '@/models/modelpresetsmodel';
import type { IPromptStoreAPI } from '@/models/promptmodel';
import type { ISettingStoreAPI } from '@/models/settingmodel';
import type { IToolStoreAPI } from '@/models/toolmodel';

import { IS_WAILS_PLATFORM } from '@/lib/features';

import * as wailsImpl from './wailsapi';

export let log: ILogger;
export let backendAPI: IBackendAPI;
export let conversationStoreAPI: IConversationStoreAPI;
export let providerSetAPI: IProviderSetAPI;
export let settingstoreAPI: ISettingStoreAPI;
export let modelPresetStoreAPI: IModelPresetStoreAPI;
export let promptStoreAPI: IPromptStoreAPI;
export let toolStoreAPI: IToolStoreAPI;

// Conditional initialization
if (IS_WAILS_PLATFORM) {
	// Initialize with Wails implementations
	log = new wailsImpl.WailsLogger();
	backendAPI = new wailsImpl.WailsBackendAPI();
	conversationStoreAPI = new wailsImpl.WailsConversationStoreAPI();
	providerSetAPI = new wailsImpl.WailsProviderSetAPI();
	settingstoreAPI = new wailsImpl.WailsSettingStoreAPI();
	modelPresetStoreAPI = new wailsImpl.WailsModelPresetStoreAPI();
	promptStoreAPI = new wailsImpl.WailsPromptStoreAPI();
	toolStoreAPI = new wailsImpl.WailsToolStoreAPI();
} else {
	// Error for unsupported platforms
	throw new Error('Unsupported platform');
}
