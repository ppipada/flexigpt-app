/* ********************************************************************
 *   Declaration file for the API exposed over the context bridge
 *********************************************************************/
import { ISettingsAPI } from 'settingmodel';

export interface IBackendAPI {
	ping: () => string;
	log: (level: string, ...args: unknown[]) => void;
}

declare global {
	interface Window {
		BackendAPI: IBackendAPI;
		SettingsAPI: ISettingsAPI;
		loggerSet: booleanl;
	}
}
