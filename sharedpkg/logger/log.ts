/* eslint-disable @typescript-eslint/no-unused-vars */
import { createLogger, format, transports, Logger as WinstonLogger } from 'winston';

/* eslint-disable @typescript-eslint/no-empty-function */
export interface ILogger {
	log(...args: unknown[]): void;
	error(...args: unknown[]): void;
	info(...args: unknown[]): void;
	debug(...args: unknown[]): void;
	warn(...args: unknown[]): void;
}

// Utility function to convert unknown to string
const stringifyArgs = (args: unknown[]): string => {
	return args
		.map(arg => {
			if (typeof arg === 'object' && arg !== null) {
				return JSON.stringify(arg);
			}
			return String(arg);
		})
		.join(' ');
};

// Winston logger instance
const winstonLogger: WinstonLogger = createLogger({
	level: 'debug', // Set the logging level
	format: format.combine(
		format.timestamp(),
		format.printf(({ timestamp, level, message }) => `${timestamp} [${level.toUpperCase()}]: ${message}`)
	),
	transports: [new transports.Console()],
});

export const noopLogger: ILogger = {
	log: (..._args: unknown[]) => {},
	error: (..._args: unknown[]) => {},
	info: (..._args: unknown[]) => {},
	debug: (..._args: unknown[]) => {},
	warn: (..._args: unknown[]) => {},
};

export const winstonBackedLogger: ILogger = {
	log: (...args: unknown[]) => {
		winstonLogger.log('info', stringifyArgs(args));
	},
	error: (...args: unknown[]) => {
		winstonLogger.error(stringifyArgs(args));
	},
	info: (...args: unknown[]) => {
		winstonLogger.info(stringifyArgs(args));
	},
	debug: (...args: unknown[]) => {
		winstonLogger.debug(stringifyArgs(args));
	},
	warn: (...args: unknown[]) => {
		winstonLogger.warn(stringifyArgs(args));
	},
};

let globalLogger: ILogger = winstonBackedLogger;

export function setGlobalLogger(logger: ILogger): void {
	globalLogger = logger;
}

export const log = {
	log: (...args: unknown[]) => globalLogger.log(...args),
	error: (...args: unknown[]) => globalLogger.error(...args),
	info: (...args: unknown[]) => globalLogger.info(...args),
	debug: (...args: unknown[]) => globalLogger.debug(...args),
	warn: (...args: unknown[]) => globalLogger.warn(...args),
};
