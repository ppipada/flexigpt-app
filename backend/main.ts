import { BrowserWindow, CallbackResponse, OnBeforeRequestListenerDetails, app, ipcMain, session } from 'electron';
import electronIsDev from 'electron-is-dev';
import log from 'electron-log';
import electronUpdater from 'electron-updater';
import path from 'node:path';
import { dirname } from 'path';
import { fileURLToPath, format as urlformat } from 'url';
import { SettingsStore } from './settingsstore/store';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const FRONTEND_PATH_PREFIX = '/frontend/build';
const PUBLIC_FILES_PATHS = ['/icon.png', '/favicon.ico'];
const HANDLE_FILES_PREFIXES = [`file://${FRONTEND_PATH_PREFIX}`, ...PUBLIC_FILES_PATHS.map(path => `file://${path}`)];
const ICON_PATH = path.resolve(__dirname, `../../${FRONTEND_PATH_PREFIX}/favicon.ico`);
const PRELOAD_PATH = path.join(__dirname, 'preload.js');

const { autoUpdater } = electronUpdater;
let appWindow: BrowserWindow | null = null;
let settingsManager: SettingsStore;

class AppUpdater {
	constructor() {
		log.transports.file.level = 'info';
		autoUpdater.logger = log;
		autoUpdater.checkForUpdatesAndNotify();
	}
}

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
	console.log('Window loading URL: ', loadurl);
	appWindow.loadURL(loadurl);
	appWindow.maximize();
	appWindow.setMenu(null);
	appWindow.show();

	if (electronIsDev) appWindow.webContents.openDevTools({ mode: 'right' });

	appWindow.on('closed', () => {
		appWindow = null;
	});
};

const initializeSettingsManager = async () => {
	const settingsFilePath = path.join(app.getPath('userData'), 'settings.json');
	console.log(`Settings file url: ${settingsFilePath}`);
	settingsManager = new SettingsStore(settingsFilePath);
	await settingsManager.initialize();
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
	// console.log(`Input URL: ${origurl} Callpath: ${callurl}`);
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

// This is done to fix: [ERROR:gl_surface_presentation_helper.cc(260)] GetVSyncParametersIfAvailable() failed in linux.
app.disableHardwareAcceleration();

app.on('ready', async () => {
	// new AppUpdater();
	await initializeSettingsManager();
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

ipcMain.handle('settings-store:getall', async () => {
	return await settingsManager.getAllSettings();
});

ipcMain.handle('settings-store:set', async (_event, key: string, value: any) => {
	await settingsManager.setSetting(key, value);
});

ipcMain.handle('backend:ping', () => {
	return 'pong';
});
