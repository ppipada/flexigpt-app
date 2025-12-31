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
