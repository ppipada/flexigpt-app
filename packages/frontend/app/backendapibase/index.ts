'use client';

import { IProviderSetAPI } from '@/models/aiprovidermodel';
import { IBackendAPI } from '@/models/backendmodel';
import { IConversationStoreAPI } from '@/models/conversationmodel';
import { ILogger } from '@/models/loggermodel';
import { ISettingStoreAPI } from '@/models/settingmodel';

let LoggerImpl: new () => ILogger;
let BackendAPIImpl: new () => IBackendAPI;
let ConversationStoreAPIImpl: new () => IConversationStoreAPI;
let ProviderSetAPIImpl: new () => IProviderSetAPI;
let SettingStoreAPIImpl: new () => ISettingStoreAPI;

if (process.env.NEXT_PUBLIC_PLATFORM === 'electron') {
	// eslint-disable-next-line @typescript-eslint/no-var-requires
	const electronAPI = require('./electronapi');
	LoggerImpl = electronAPI.ElectronLogger;
	BackendAPIImpl = electronAPI.ElectronBackendAPI;
	ConversationStoreAPIImpl = electronAPI.ElectronConversationStoreAPI;
	ProviderSetAPIImpl = electronAPI.ElectronProviderSetAPI;
	SettingStoreAPIImpl = electronAPI.ElectronSettingStoreAPI;
} else if (process.env.NEXT_PUBLIC_PLATFORM === 'wails') {
	// eslint-disable-next-line @typescript-eslint/no-var-requires
	const wailsAPI = require('./wailsapi');
	LoggerImpl = wailsAPI.WailsLogger;
	BackendAPIImpl = wailsAPI.WailsBackendAPI;
	ConversationStoreAPIImpl = wailsAPI.WailsConversationStoreAPI;
	ProviderSetAPIImpl = wailsAPI.WailsProviderSetAPI;
	SettingStoreAPIImpl = wailsAPI.WailsSettingStoreAPI;
} else {
	throw new Error(`Unsupported platform: ${process.env.NEXT_PUBLIC_PLATFORM}`);
}

export const log = new LoggerImpl();
export const backendAPI = new BackendAPIImpl();
export const conversationStoreAPI = new ConversationStoreAPIImpl();
export const providerSetAPI = new ProviderSetAPIImpl();
export const settingstoreAPI = new SettingStoreAPIImpl();
