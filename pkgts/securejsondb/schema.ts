export interface SecureSchema {
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	[key: string]: any;
	_sensitiveKeys?: string[];
}
