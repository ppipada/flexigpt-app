/* ********************************************************************
 *   Declaration file for the API exposed over the context bridge
 *********************************************************************/
import { IBackendAPI, ISettingsAPI } from 'sharedpkg/settings/types';

declare global {
	interface Window {
		BackendAPI: IBackendAPI;
		SettingsAPI: ISettingsAPI;
	}
}
