import { safeStorage } from 'electron';
import { PathLike } from 'fs';
import { DataFile, DataFileSync } from 'lowdb/node';
import { SecureSchema } from './schema';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function getNestedProperty(obj: any, path: string) {
	return path.split('.').reduce((o, p) => (o ? o[p] : undefined), obj);
}
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function setNestedProperty(obj: any, path: string, value: any) {
	const keys = path.split('.');
	const lastKey = keys.pop();
	const lastObj = keys.reduce((o, k) => (o[k] = o[k] || {}), obj);
	if (lastKey) lastObj[lastKey] = value;
}

export class SecureJSONFile<T extends SecureSchema> extends DataFile<T> {
	constructor(filename: PathLike) {
		super(filename, {
			parse: (text: string) => {
				const parsedData = JSON.parse(text) as T;
				const sensitiveKeys = parsedData._sensitiveKeys || [];

				// Decrypt keys
				sensitiveKeys.forEach(key => {
					const value = getNestedProperty(parsedData, key);
					if (value) {
						setNestedProperty(parsedData, key, safeStorage.decryptString(Buffer.from(value, 'base64')));
					}
				});

				return parsedData;
			},
			stringify: (data: T) => {
				const encryptedData = { ...data };
				const sensitiveKeys = encryptedData._sensitiveKeys || [];

				// Encrypt keys
				sensitiveKeys.forEach(key => {
					const value = getNestedProperty(encryptedData, key);
					if (value) {
						setNestedProperty(encryptedData, key, safeStorage.encryptString(value).toString('base64'));
					}
				});

				return JSON.stringify(encryptedData, null, 2);
			},
		});
	}
}

export class SecureJSONFileSync<T extends SecureSchema> extends DataFileSync<T> {
	constructor(filename: PathLike) {
		super(filename, {
			parse: (text: string) => {
				const parsedData = JSON.parse(text) as T;
				const sensitiveKeys = parsedData._sensitiveKeys || [];

				// Decrypt keys
				sensitiveKeys.forEach(key => {
					const value = getNestedProperty(parsedData, key);
					if (value) {
						setNestedProperty(parsedData, key, safeStorage.decryptString(Buffer.from(value, 'base64')));
					}
				});

				return parsedData;
			},
			stringify: (data: T) => {
				const encryptedData = { ...data };
				const sensitiveKeys = encryptedData._sensitiveKeys || [];

				// Encrypt keys
				sensitiveKeys.forEach(key => {
					const value = getNestedProperty(encryptedData, key);
					if (value) {
						setNestedProperty(encryptedData, key, safeStorage.encryptString(value).toString('base64'));
					}
				});

				return JSON.stringify(encryptedData, null, 2);
			},
		});
	}
}
