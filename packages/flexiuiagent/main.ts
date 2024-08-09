import {
	app,
	BrowserWindow,
	CallbackResponse,
	globalShortcut,
	ipcMain,
	OnBeforeRequestListenerDetails,
	protocol,
	session,
} from 'electron';
import electronIsDev from 'electron-is-dev';
// import electronUpdater from 'electron-updater';
import path from 'node:path';
import { dirname } from 'path';
import { fileURLToPath, format as urlformat } from 'url';

import { ILogger, log } from 'logger';
import { createILogger } from 'winstonlogger';

// const logLevel = "info";
const logLevel = 'debug';
if (!electronIsDev) {
	const wlog = createILogger('file', logLevel, path.join(app.getPath('userData'), 'logs'));
	log.setLogger(wlog);
	log.info('Backend: Running in production');
} else {
	const wlog = createILogger('file', logLevel, path.join(app.getPath('userData'), 'logs'));
	log.setLogger(wlog);
	log.info('Backend: Running in dev');
}

import { ProviderSet } from 'aiproviderimpl';
import { ChatCompletionRequestMessage, IProviderSetAPI, ModelName, ProviderName } from 'aiprovidermodel';
import { Conversation, ConversationMessage } from 'conversationmodel';
import { ConversationCollection } from 'conversationstore';
import { SettingsStore } from 'settingstore';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const FRONTEND_PATH_PREFIX = '/frontend/build';
const PUBLIC_FILES_PATHS = ['/icon.png', '/favicon.ico'];
const HANDLE_FILES_PREFIXES = [`file://${FRONTEND_PATH_PREFIX}`, ...PUBLIC_FILES_PATHS.map(path => `file://${path}`)];
// const ICON_PATH = path.resolve(__dirname, `../../${FRONTEND_PATH_PREFIX}/favicon.ico`);
const PRELOAD_PATH = path.join(__dirname, 'preload.js');

// const { autoUpdater } = electronUpdater;
let appWindow: BrowserWindow | null = null;
let settingsManager: SettingsStore;
let conversationManager: ConversationCollection;
let providerSetManager: IProviderSetAPI;

// class AppUpdater {
// 	constructor() {
// 		log.transports.file.level = 'info';
// 		autoUpdater.logger = log;
// 		autoUpdater.checkForUpdatesAndNotify();
// 	}
// }

const installExtensions = async () => {
	// Install extensions if in development mode
};

const spawnAppWindow = async () => {
	if (electronIsDev) await installExtensions();

	appWindow = new BrowserWindow({
		width: 800,
		height: 600,
		// icon: ICON_PATH,
		// show: false,
		webPreferences: {
			preload: PRELOAD_PATH,
		},
	});

	let loadurl = 'http://localhost:3000';
	if (!electronIsDev) {
		loadurl = urlformat({
			pathname: path.join(__dirname, `../..${FRONTEND_PATH_PREFIX}/index.html`),
			protocol: 'file:',
			slashes: true,
		});
	}
	// log.info('Window loading URL: ', loadurl);
	appWindow.loadURL(loadurl);
	appWindow.maximize();
	appWindow.setMenu(null);
	globalShortcut.register('Control+Shift+I', () => {
		if (appWindow) {
			appWindow.webContents.openDevTools();
		}
	});
	appWindow.show();

	if (electronIsDev) appWindow.webContents.openDevTools({ mode: 'right' });

	appWindow.on('closed', () => {
		appWindow = null;
	});
};

const initializeSettingsManager = async () => {
	const settingsFilePath = path.join(app.getPath('userData'), 'settings.json');
	log.info(`Settings file url: ${settingsFilePath}`);
	settingsManager = new SettingsStore(settingsFilePath);
	await settingsManager.initialize();
};

const initializeConversationManager = async () => {
	const conversationDir = path.join(app.getPath('userData'), 'conversations');
	log.info(`Conversation directory: ${conversationDir}`);
	conversationManager = new ConversationCollection(conversationDir);
};

const initializeProviderSetManager = async () => {
	providerSetManager = new ProviderSet(ProviderName.OPENAI);
};

const getActualURL = (origurl: string) => {
	let callurl = origurl;

	if (HANDLE_FILES_PREFIXES.some(prefix => callurl.startsWith(prefix))) {
		// Remove the file://
		let actualURL = callurl.substring(7);
		// For public files add the frontend prefix
		if (PUBLIC_FILES_PATHS.some(pfile => actualURL === pfile)) {
			actualURL = FRONTEND_PATH_PREFIX + actualURL;
		}
		const percentIndex = actualURL.indexOf('%');
		if (percentIndex !== -1) {
			actualURL = actualURL.substring(0, percentIndex);
		}
		const qIndex = actualURL.indexOf('?');
		if (qIndex !== -1) {
			actualURL = actualURL.substring(0, qIndex);
		}
		// Handle if its a page request
		if (actualURL.endsWith('/')) {
			actualURL += 'index.html';
		}
		// Create a absolute url from the actual url
		callurl = urlformat({
			pathname: path.join(__dirname, `../..${actualURL}`),
			protocol: 'file:',
			slashes: true,
		});
	}
	// log.debug(`Input URL: ${origurl} Callpath: ${callurl}`);
	return callurl;
};

const handleAccessRequest = (
	details: OnBeforeRequestListenerDetails,
	callback: (response: CallbackResponse) => void
) => {
	const callurl = getActualURL(details.url);
	if (callurl !== details.url) {
		callback({ redirectURL: callurl });
	} else {
		callback({});
	}
};

const interceptFileProtocol = () => {
	protocol.interceptFileProtocol('file', (request, callback) => {
		const callurl = getActualURL(request.url);
		callback({ path: callurl.substring(8) });
	});
};

// This is done to fix: [ERROR:gl_surface_presentation_helper.cc(260)] GetVSyncParametersIfAvailable() failed in linux.
app.disableHardwareAcceleration();

app.on('ready', async () => {
	// new AppUpdater();
	await initializeSettingsManager();
	await initializeConversationManager();
	await initializeProviderSetManager();
	interceptFileProtocol();
	spawnAppWindow();
	session.defaultSession.webRequest.onBeforeRequest(handleAccessRequest);
});

app.on('window-all-closed', () => {
	if (process.platform !== 'darwin') {
		app.quit();
	}
});

/*
 * ======================================================================================
 *                                IPC Main Events
 * ======================================================================================
 */

ipcMain.handle('settingstore:getall', async () => {
	return await settingsManager.getAllSettings();
});

ipcMain.handle('settingstore:set', async (_event, key: string, value: any) => {
	await settingsManager.setSetting(key, value);
});

ipcMain.handle('backend:ping', async () => {
	return 'pong';
});

ipcMain.handle('backend:log', async (_event, level: string, ...args: unknown[]) => {
	log[level as keyof ILogger](...args);
});

ipcMain.handle('conversation:save', async (_event, conversation: Conversation) => {
	await conversationManager.saveConversation(conversation);
});

ipcMain.handle('conversation:delete', async (_event, id: string, title: string) => {
	await conversationManager.deleteConversation(id, title);
});

ipcMain.handle('conversation:get', async (_event, id: string, title: string) => {
	return await conversationManager.getConversation(id, title);
});

ipcMain.handle('conversation:list', async (_event, token?: string) => {
	return await conversationManager.listConversations(token);
});

ipcMain.handle(
	'conversation:addMessage',
	async (_event, id: string, title: string, newMessage: ConversationMessage) => {
		return await conversationManager.addMessageToConversation(id, title, newMessage);
	}
);

// eslint-disable-next-line @typescript-eslint/no-unused-vars
ipcMain.handle('providerset:getDefaultProvider', async _event => {
	return await providerSetManager.getDefaultProvider();
	// log.info('Main provider', provider);
});

ipcMain.handle('providerset:setDefaultProvider', async (_event, provider: ProviderName) => {
	return await providerSetManager.setDefaultProvider(provider);
});

ipcMain.handle('providerset:getProviderInfo', async (_event, provider: ProviderName) => {
	return await providerSetManager.getProviderInfo(provider);
});

// eslint-disable-next-line @typescript-eslint/no-unused-vars
ipcMain.handle('providerset:getConfigurationInfo', async _event => {
	return await providerSetManager.getConfigurationInfo();
});

ipcMain.handle(
	'providerset:setAttribute',
	async (
		_event,
		provider: ProviderName,
		apiKey?: string,
		defaultModel?: ModelName,
		defaultTemperature?: number,
		defaultOrigin?: string
	) => {
		return await providerSetManager.setAttribute(provider, apiKey, defaultModel, defaultTemperature, defaultOrigin);
	}
);

ipcMain.handle(
	'providerset:getCompletionRequest',
	async (
		_event,
		provider: ProviderName,
		prompt: string,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		inputParams?: {
			[key: string]: any;
		},
		stream?: boolean
	) => {
		return await providerSetManager.getCompletionRequest(provider, prompt, prevMessages, inputParams, stream);
	}
);

ipcMain.handle('providerset:completion', async (event, provider, input, callbackId) => {
	const onStreamData = async (data: string) => {
		await event.sender.send(`${callbackId}`, data);
	};

	return await providerSetManager.completion(provider, input, onStreamData);
});
