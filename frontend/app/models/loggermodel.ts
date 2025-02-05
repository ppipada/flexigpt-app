/* eslint-disable @typescript-eslint/no-unused-vars */
/* eslint-disable @typescript-eslint/no-empty-function */
export interface ILogger {
	log(...args: unknown[]): void;
	error(...args: unknown[]): void;
	info(...args: unknown[]): void;
	debug(...args: unknown[]): void;
	warn(...args: unknown[]): void;
}

/**
 * @public
 */
export const noopLogger: ILogger = {
	log: (..._args: unknown[]) => {},
	error: (..._args: unknown[]) => {},
	info: (..._args: unknown[]) => {},
	debug: (..._args: unknown[]) => {},
	warn: (..._args: unknown[]) => {},
};

/**
 * @public
 * Define a new console logger
 */
export const consoleLogger: ILogger = {
	log: (...args: unknown[]) => console.log(...args),
	error: (...args: unknown[]) => console.error(...args),
	info: (...args: unknown[]) => console.info(...args),
	debug: (...args: unknown[]) => console.debug(...args),
	warn: (...args: unknown[]) => console.warn(...args),
};
