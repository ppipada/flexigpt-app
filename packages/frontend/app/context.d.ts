/* ********************************************************************
 *   Declaration file for the API exposed over the context bridge
 *********************************************************************/
import { IBackendAPI, ISettingsAPI } from 'settings';

declare global {
	interface Window {
		BackendAPI: IBackendAPI;
		SettingsAPI: ISettingsAPI;
	}
}
