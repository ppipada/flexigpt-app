import type { ILogger } from '@/apis/interface';

export const noopLogger: ILogger = {
	log: (..._args: unknown[]) => {},
	error: (..._args: unknown[]) => {},
	info: (..._args: unknown[]) => {},
	debug: (..._args: unknown[]) => {},
	warn: (..._args: unknown[]) => {},
};

export const consoleLogger: ILogger = {
	log: (...args: unknown[]) => {
		console.log(...args);
	},
	error: (...args: unknown[]) => {
		console.error(...args);
	},
	info: (...args: unknown[]) => {
		console.info(...args);
	},
	debug: (...args: unknown[]) => {
		console.debug(...args);
	},
	warn: (...args: unknown[]) => {
		console.warn(...args);
	},
};
