/* eslint-disable @typescript-eslint/no-var-requires */
const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('BackendAPI', {
	ping: async () => await ipcRenderer.invoke('backend:ping'),
	log: async (level: string, ...args: unknown[]) => {
		ipcRenderer.invoke('backend:log', level, ...args);
	},
});

contextBridge.exposeInMainWorld('SettingsAPI', {
	getAllSettings: async () => {
		return await ipcRenderer.invoke('settings-store:getall');
	},
	setSetting: async (key: string, value: any) => {
		await ipcRenderer.invoke('settings-store:set', key, value);
	},
});
