import { app, BrowserWindow, ipcMain } from 'electron';
import electronIsDev from 'electron-is-dev';
import log from 'electron-log';
import electronUpdater from 'electron-updater';
import path from 'node:path';
import { dirname } from 'path';
import { fileURLToPath } from 'url';
import { SettingsStore } from './settingsstore/store';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
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
		width: 1920,
		height: 1080,
		icon: getAssetPath('icon.png'),
		show: false,
		webPreferences: {
			preload: PRELOAD_PATH,
		},
	});

	appWindow.loadURL(
		electronIsDev ? 'http://localhost:3000' : `file://${path.join(__dirname, '../../frontend/build/index.html')}`
	);
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
	console.log(settingsFilePath);
	settingsManager = new SettingsStore(settingsFilePath);
	await settingsManager.initialize();
};

app.on('ready', async () => {
	new AppUpdater();
	await initializeSettingsManager();
	spawnAppWindow();
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
