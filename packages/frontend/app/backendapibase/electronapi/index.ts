/* ********************************************************************
 *   Declaration file for the API exposed over the context bridge
 *********************************************************************/
import { IProviderSetAPI } from '@/models/aiprovidermodel';
import { IBackendAPI } from '@/models/backendmodel';
import { IConversationStoreAPI } from '@/models/conversationmodel';
import { ISettingStoreAPI } from '@/models/settingmodel';

declare global {
	interface Window {
		BackendAPI: IBackendAPI;
		SettingStoreAPI: ISettingStoreAPI;
		ConversationStoreAPI: IConversationStoreAPI;
		ProviderSetAPI: IProviderSetAPI;
	}
}

export * from './aiprovider';
export * from './backend';
export * from './conversationstore';
export * from './settingstore';
