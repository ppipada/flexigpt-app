import { IS_WAILS_PLATFORM } from '@/lib/features';
import type { IProviderSetAPI } from '@/models/aiprovidermodel';
import type { IBackendAPI } from '@/models/backendmodel';
import type { IConversationStoreAPI } from '@/models/conversationmodel';
import type { ILogger } from '@/models/loggermodel';
import type { ISettingStoreAPI } from '@/models/settingmodel';

let LoggerImpl: new () => ILogger;
let BackendAPIImpl: new () => IBackendAPI;
let ConversationStoreAPIImpl: new () => IConversationStoreAPI;
let ProviderSetAPIImpl: new () => IProviderSetAPI;
let SettingStoreAPIImpl: new () => ISettingStoreAPI;

(async () => {
	if (IS_WAILS_PLATFORM) {
		// Use dynamic import instead of require
		const wailsAPI = await import('./wailsapi');
		LoggerImpl = wailsAPI.WailsLogger;
		BackendAPIImpl = wailsAPI.WailsBackendAPI;
		ConversationStoreAPIImpl = wailsAPI.WailsConversationStoreAPI;
		ProviderSetAPIImpl = wailsAPI.WailsProviderSetAPI;
		SettingStoreAPIImpl = wailsAPI.WailsSettingStoreAPI;

		// Initialize exports
		initializeExports();
	} else {
		throw new Error('Unsupported platform');
	}
})().catch((error: unknown) => {
	console.error('Failed to initialize API:', error);
});

// Define exports but initialize them later
export let log: ILogger;
export let backendAPI: IBackendAPI;
export let conversationStoreAPI: IConversationStoreAPI;
export let providerSetAPI: IProviderSetAPI;
export let settingstoreAPI: ISettingStoreAPI;

function initializeExports() {
	log = new LoggerImpl();
	backendAPI = new BackendAPIImpl();
	conversationStoreAPI = new ConversationStoreAPIImpl();
	providerSetAPI = new ProviderSetAPIImpl();
	settingstoreAPI = new SettingStoreAPIImpl();
}
