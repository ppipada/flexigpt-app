import { FileFilter, IBackendAPI } from '@/models/backendmodel';
import { ILogger } from '@/models/loggermodel';

export class ElectronLogger implements ILogger {
	log(...args: unknown[]): void {
		if (window.BackendAPI) {
			window.BackendAPI.log('log', ...args);
		} else {
			console.log(...args);
		}
	}

	error(...args: unknown[]): void {
		if (window.BackendAPI) {
			window.BackendAPI.log('error', ...args);
		} else {
			console.error(...args);
		}
	}

	info(...args: unknown[]): void {
		if (window.BackendAPI) {
			window.BackendAPI.log('info', ...args);
		} else {
			console.info(...args);
		}
	}

	debug(...args: unknown[]): void {
		if (window.BackendAPI) {
			window.BackendAPI.log('debug', ...args);
		} else {
			console.debug(...args);
		}
	}

	warn(...args: unknown[]): void {
		if (window.BackendAPI) {
			window.BackendAPI.log('warn', ...args);
		} else {
			console.warn(...args);
		}
	}
}

export class ElectronBackendAPI implements IBackendAPI {
	// Implement the ping method
	ping(): Promise<string> {
		// Assuming the ping method is synchronous and returns a string
		return window.BackendAPI.ping();
	}

	// Implement the log method
	log(level: string, ...args: unknown[]): void {
		window.BackendAPI.log(level, ...args);
	}

	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	async savefile(defaultFilename: string, contentBase64: string, filters: Array<FileFilter>): Promise<void> {
		return await window.BackendAPI.savefile(defaultFilename, contentBase64, filters);
	}
}
