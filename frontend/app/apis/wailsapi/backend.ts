import { Ping, SaveFile } from '@/apis/wailsjs/go/main/App';
import { BrowserOpenURL, LogDebug, LogError, LogInfo, LogWarning } from '@/apis/wailsjs/runtime/runtime';
import type { FileFilter, IBackendAPI } from '@/models/backendmodel';
import type { ILogger } from '@/models/loggermodel';
import { sprintf } from 'sprintf-js';

function formatMessage(args: unknown[]): string {
	if (args.length > 0 && typeof args[0] === 'string') {
		try {
			return sprintf(args[0], ...args.slice(1));
		} catch (error) {
			console.error(`Error formatting message: ${error}`);
			return '';
		}
	} else {
		console.error(`Invalid format string or arguments: ${args}`);
		return '';
	}
}

function isRuntimeAvailable(): boolean {
	return typeof window !== 'undefined' && 'runtime' in window;
}

/**
 * @public
 */
export class WailsLogger implements ILogger {
	log(...args: unknown[]): void {
		const msg = formatMessage(args);
		if (isRuntimeAvailable()) {
			if (msg) {
				LogInfo(msg);
			}
		} else {
			console.log(...args);
		}
	}

	error(...args: unknown[]): void {
		const msg = formatMessage(args);
		if (isRuntimeAvailable()) {
			if (msg) {
				LogError(msg);
			}
		} else {
			console.error(...args);
		}
	}

	info(...args: unknown[]): void {
		const msg = formatMessage(args);
		if (isRuntimeAvailable()) {
			if (msg) {
				LogInfo(msg);
			}
		} else {
			console.info(...args);
		}
	}

	debug(...args: unknown[]): void {
		const msg = formatMessage(args);
		if (isRuntimeAvailable()) {
			if (msg) {
				LogDebug(msg);
			}
		} else {
			console.debug(...args);
		}
	}

	warn(...args: unknown[]): void {
		const msg = formatMessage(args);
		if (isRuntimeAvailable()) {
			if (msg) {
				LogWarning(msg);
			}
		} else {
			console.warn(...args);
		}
	}
}

/**
 * @public
 */
export class WailsBackendAPI implements IBackendAPI {
	// Implement the ping method
	ping(): Promise<string> {
		return Ping();
	}

	// Implement the log method
	log(level: string, ...args: unknown[]): void {
		const msg = formatMessage(args);
		if (isRuntimeAvailable()) {
			if (msg) {
				switch (level) {
					case 'info':
						LogInfo(msg);
						break;
					case 'error':
						LogError(msg);
						break;
					case 'debug':
						LogDebug(msg);
						break;
					case 'warn':
						LogWarning(msg);
						break;
				}
			}
		} else {
			console.log(...args);
		}
	}

	async savefile(defaultFilename: string, contentBase64: string, filters: Array<FileFilter>): Promise<void> {
		// Call the Go backend method to save the file
		try {
			await SaveFile(defaultFilename, contentBase64, filters);
		} catch (err) {
			console.error('Error saving file:', err);
		}
		return;
	}

	openurl(url: string): void {
		BrowserOpenURL(url);
	}
}
