/* eslint-disable @typescript-eslint/no-var-requires */
const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('BackendAPI', {
	ping: async () => await ipcRenderer.invoke('backend:ping'),
	log: async (level: string, ...args: unknown[]) => {
		ipcRenderer.invoke('backend:log', level, ...args);
	},
});

contextBridge.exposeInMainWorld('SettingStoreAPI', {
	getAllSettings: async () => {
		return await ipcRenderer.invoke('settingstore:getall');
	},
	setSetting: async (key: string, value: any) => {
		await ipcRenderer.invoke('settingstore:set', key, value);
	},
});

contextBridge.exposeInMainWorld('ConversationStoreAPI', {
	saveConversation: async (conversation: any) => {
		await ipcRenderer.invoke('conversationstore:save', conversation);
	},
	deleteConversation: async (id: string, title: string) => {
		await ipcRenderer.invoke('conversationstore:delete', id, title);
	},
	getConversation: async (id: string, title: string) => {
		return await ipcRenderer.invoke('conversationstore:get', id, title);
	},
	listConversations: async (token?: string) => {
		return await ipcRenderer.invoke('conversationstore:list', token);
	},
	addMessageToConversation: async (id: string, title: string, newMessage: any) => {
		return await ipcRenderer.invoke('conversationstore:addMessage', id, title, newMessage);
	},
});

contextBridge.exposeInMainWorld('ProviderSetAPI', {
	getDefaultProvider: async () => {
		const provider = await ipcRenderer.invoke('providerset:getDefaultProvider');
		// console.log('preload', provider);
		return provider;
	},
	setDefaultProvider: async (provider: any) => {
		return await ipcRenderer.invoke('providerset:setDefaultProvider', provider);
	},
	getProviderInfo: async (provider: any) => {
		return await ipcRenderer.invoke('providerset:getProviderInfo', provider);
	},
	getConfigurationInfo: async () => {
		return await ipcRenderer.invoke('providerset:getConfigurationInfo');
	},

	setAttribute: async (
		provider: any,
		apiKey?: string,
		defaultModel?: any,
		defaultTemperature?: number,
		defaultOrigin?: string
	) => {
		return await ipcRenderer.invoke(
			'providerset:setAttribute',
			provider,
			apiKey,
			defaultModel,
			defaultTemperature,
			defaultOrigin
		);
	},
	getCompletionRequest: async (
		provider: any,
		prompt: string,
		prevMessages?: any,
		inputParams?: any,
		stream?: boolean
	) => {
		return await ipcRenderer.invoke(
			'providerset:getCompletionRequest',
			provider,
			prompt,
			prevMessages,
			inputParams,
			stream
		);
	},
	completion: async (provider: any, input: any, onStreamData: any) => {
		const callbackId = `stream-data-callback-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
		if (onStreamData) {
			ipcRenderer.on(callbackId, async (_event, data) => await onStreamData(data));
		}
		return await ipcRenderer.invoke('providerset:completion', provider, input, callbackId);
	},
});
