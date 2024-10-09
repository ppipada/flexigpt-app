/* ********************************************************************
 *   Declaration file for the API exposed over the context bridge
 *********************************************************************/
import { IProviderSetAPI } from '@/models/aiprovidermodel';
import { IConversationAPI } from '@/models/conversationmodel';
import { ISettingsAPI } from '@/models/settingmodel';

export interface IBackendAPI {
	ping: () => string;
	log: (level: string, ...args: unknown[]) => void;
}

declare global {
	interface Window {
		BackendAPI: IBackendAPI;
		SettingsAPI: ISettingsAPI;
		ConversationAPI: IConversationAPI;
		ProviderSetAPI: IProviderSetAPI;
		loggerSet: boolean;
	}
}
