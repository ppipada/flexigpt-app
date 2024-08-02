/* eslint-disable @typescript-eslint/no-unused-vars */
import { ILogger, log, setGlobalLogger } from 'logger';

// Define an IPC logger
export const ipcLogger: ILogger = {
	log: (...args: unknown[]) => window.BackendAPI.log('log', ...args),
	error: (...args: unknown[]) => window.BackendAPI.log('error', ...args),
	info: (...args: unknown[]) => window.BackendAPI.log('info', ...args),
	debug: (...args: unknown[]) => window.BackendAPI.log('debug', ...args),
	warn: (...args: unknown[]) => window.BackendAPI.log('warn', ...args),
};

export function setLogger() {
	if (window.loggerSet) {
		return;
	}
	setGlobalLogger(ipcLogger);
	log.debug('Frontend: DOM loaded');
	window.loggerSet = true;
}
