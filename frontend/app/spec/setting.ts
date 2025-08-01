export enum ThemeType {
	System = 'system',
	Light = 'light',
	Dark = 'dark',
	Other = 'other',
}
export interface AppTheme {
	type: ThemeType;
	name: string;
}

export type AuthKeyType = string;
export const AuthKeyTypeProvider = 'provider';

export type AuthKeyName = string;

export interface AuthKey {
	secret: string;
	sha256: string;
	nonEmpty: boolean;
}

export interface AuthKeyMeta {
	type: AuthKeyType;
	keyName: AuthKeyName;
	sha256: string;
	nonEmpty: boolean;
}

export interface SettingsSchema {
	appTheme: AppTheme;
	authKeys: AuthKeyMeta[];
}

export interface ISettingStoreAPI {
	setAppTheme: (theme: AppTheme) => Promise<void>;
	getAuthKey: (type: AuthKeyType, keyName: AuthKeyName) => Promise<AuthKey>;
	deleteAuthKey: (type: AuthKeyType, keyName: AuthKeyName) => Promise<void>;
	setAuthKey: (type: AuthKeyType, keyName: AuthKeyName, secret: string) => Promise<void>;
	getSettings: (forceFetch?: boolean) => Promise<SettingsSchema>;
}
