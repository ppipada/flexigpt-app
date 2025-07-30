import type { IProviderSetAPI } from '@/spec/aiprovider';
import type { IBackendAPI } from '@/spec/backend';
import type { IConversationStoreAPI } from '@/spec/conversation';
import type { ILogger } from '@/spec/logger';
import type { IModelPresetStoreAPI } from '@/spec/modelpreset';
import type { IPromptStoreAPI } from '@/spec/prompt';
import type { ISettingStoreAPI } from '@/spec/setting';
import type { IToolStoreAPI } from '@/spec/tool';

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
