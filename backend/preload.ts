/* eslint-disable @typescript-eslint/no-var-requires */
const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('BackendAPI', {
	ping: () => ipcRenderer.invoke('backend:ping'),
});

contextBridge.exposeInMainWorld('SettingsAPI', {
	getAllSettings: async () => {
		return await ipcRenderer.invoke('settings-store:getall');
	},
	setSetting: async (key: string, value: any) => {
		await ipcRenderer.invoke('settings-store:set', key, value);
	},
});
