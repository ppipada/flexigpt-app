/* eslint-disable @typescript-eslint/no-unused-vars */
import * as fs from 'fs';
import { ILogger } from 'logger';
import * as path from 'path';
import { createLogger, format, transports, Logger as WinstonLogger } from 'winston';

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

export const createWinstonLogger = (logType: 'console' | 'file', logLevel: string, logDir?: string): WinstonLogger => {
	const loggerTransports = [];

	if (logType === 'console') {
		loggerTransports.push(new transports.Console());
	} else if (logType === 'file') {
		// Use provided log directory or default to 'logs'
		const directory = logDir || path.join(__dirname, 'logs');
		if (!fs.existsSync(directory)) {
			fs.mkdirSync(directory);
		}
		loggerTransports.push(
			new transports.File({ filename: path.join(directory, 'error.log'), level: 'error' }),
			new transports.File({ filename: path.join(directory, 'combined.log') })
		);
	}

	return createLogger({
		level: logLevel,
		format: format.combine(
			format.timestamp(),
			format.printf(({ timestamp, level, message }) => `${timestamp} [${level.toUpperCase()}]: ${message}`)
		),
		transports: loggerTransports,
	});
};

// Example ILogger implementation using the created logger
export const createILogger = (logType: 'console' | 'file', logLevel: string, logDir?: string): ILogger => {
	const winstonLogger = createWinstonLogger(logType, logLevel, logDir);

	return {
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
};
