/* ********************************************************************
 *   Declaration file for the API exposed over the context bridge
 *********************************************************************/
import { IConversationAPI } from 'conversationmodel';
import { ISettingsAPI } from 'settingmodel';

export interface IBackendAPI {
	ping: () => string;
	log: (level: string, ...args: unknown[]) => void;
}

declare global {
	interface Window {
		BackendAPI: IBackendAPI;
		SettingsAPI: ISettingsAPI;
		ConversationAPI: IConversationAPI;
		loggerSet: booleanl;
	}
}
