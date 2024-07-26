/* ********************************************************************
 *   Declaration file for the API exposed over the context bridge
 *********************************************************************/
import { SettingsSchema } from './app/api/settings';

export interface IBackendAPI {
	ping: () => Promise<string>;
}

export interface ISettingsAPI {
	getAllSettings: () => Promise<SettingsSchema>;
	setSetting: (key: string, value: any) => Promise<void>;
}

declare global {
	interface Window {
		BackendAPI: IBackendAPI;
		SettingsAPI: ISettingsAPI;
	}
}
