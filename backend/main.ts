import { BrowserWindow, app, ipcMain, net, protocol } from 'electron';
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

	const RESOURCES_PATH = electronIsDev
		? path.join(__dirname, '../../assets')
		: path.join(process.resourcesPath, 'assets');

	const getAssetPath = (...paths: string[]): string => {
		return path.join(RESOURCES_PATH, ...paths);
	};

	const PRELOAD_PATH = path.join(__dirname, 'preload.js');

	appWindow = new BrowserWindow({
		width: 800,
		height: 600,
		icon: getAssetPath('icon.png'),
		show: false,
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
	let callurl = decodeURI(origurl);

	if (HANDLE_FILES_PREFIXES.some(prefix => callurl.startsWith(prefix))) {
		// Remove the file://
		let actualURL = callurl.substring(7);
		if (PUBLIC_FILES_PATHS.some(pfile => actualURL === pfile)) {
			// For public files add the frontend prefix
			actualURL = FRONTEND_PATH_PREFIX + actualURL;
		}
		// Create a absolute url from the actual url
		callurl = urlformat({
			pathname: path.join(__dirname, `../..${actualURL}`),
			protocol: 'file:',
			slashes: true,
		});
	}
	console.log(`Input URL: ${origurl} Callpath: ${callurl}`);
	return callurl;
};

const interceptFileProtocol = () => {
	protocol.interceptFileProtocol('file', (request, callback) => {
		const callurl = getActualURL(request.url);
		callback({ path: callurl.substring(8) });
	});
};

const handleFileProtocol = () => {
	protocol.handle('file', (request: Request) => {
		const callurl = getActualURL(request.url);
		// fetch the new path, without reinvoking this handler
		return net.fetch(callurl, { bypassCustomProtocolHandlers: true });
	});
};

app.on('ready', async () => {
	// new AppUpdater();
	await initializeSettingsManager();
	spawnAppWindow();
});

app.on('window-all-closed', () => {
	if (process.platform !== 'darwin') {
		app.quit();
	}
});

app.whenReady().then(() => {
	// interceptFileProtocol();
	handleFileProtocol();
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
