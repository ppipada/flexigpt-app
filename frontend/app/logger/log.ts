/* eslint-disable @typescript-eslint/no-unused-vars */
/* eslint-disable @typescript-eslint/no-empty-function */
export interface ILogger {
	log(...args: unknown[]): void;
	error(...args: unknown[]): void;
	info(...args: unknown[]): void;
	debug(...args: unknown[]): void;
	warn(...args: unknown[]): void;
}

export const noopLogger: ILogger = {
	log: (..._args: unknown[]) => {},
	error: (..._args: unknown[]) => {},
	info: (..._args: unknown[]) => {},
	debug: (..._args: unknown[]) => {},
	warn: (..._args: unknown[]) => {},
};

// Define a new console logger
export const consoleLogger: ILogger = {
	log: (...args: unknown[]) => console.log(...args),
	error: (...args: unknown[]) => console.error(...args),
	info: (...args: unknown[]) => console.info(...args),
	debug: (...args: unknown[]) => console.debug(...args),
	warn: (...args: unknown[]) => console.warn(...args),
};

class Logger {
	private logger;

	constructor(l: ILogger) {
		this.logger = l;
	}

	setLogger(logger: ILogger): void {
		this.logger = logger;
	}

	log(...args: unknown[]): void {
		this.logger.log(...args);
	}

	error(...args: unknown[]): void {
		this.logger.error(...args);
	}

	info(...args: unknown[]): void {
		this.logger.info(...args);
	}

	debug(...args: unknown[]): void {
		this.logger.debug(...args);
	}

	warn(...args: unknown[]): void {
		this.logger.warn(...args);
	}
}

export const log = new Logger(consoleLogger);
