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
		return await ipcRenderer.invoke('settingstore:getall');
	},
	setSetting: async (key: string, value: any) => {
		await ipcRenderer.invoke('settingstore:set', key, value);
	},
});

contextBridge.exposeInMainWorld('ConversationAPI', {
	saveConversation: async (conversation: any) => {
		await ipcRenderer.invoke('conversation:save', conversation);
	},
	createNewConversation: async (title: string) => {
		return await ipcRenderer.invoke('conversation:create', title);
	},
	deleteConversation: async (id: string, title: string) => {
		await ipcRenderer.invoke('conversation:delete', id, title);
	},
	getConversation: async (id: string, title: string) => {
		return await ipcRenderer.invoke('conversation:get', id, title);
	},
	listConversations: async (token?: string) => {
		return await ipcRenderer.invoke('conversation:list', token);
	},
	addMessageToConversation: async (id: string, title: string, newMessage: any) => {
		return await ipcRenderer.invoke('conversation:addMessage', id, title, newMessage);
	},
});
