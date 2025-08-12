import { CustomThemeDark, CustomThemeLight } from '@/spec/theme_consts';

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

export const toProviderName = (t: ThemeType | AppTheme): string => {
	const type = typeof t === 'string' ? t : t.type;
	if (type === ThemeType.System) return 'system';
	if (type === ThemeType.Light) return CustomThemeLight;
	if (type === ThemeType.Dark) return CustomThemeDark;
	return typeof t === 'string' ? t : t.name; /* ThemeType.Other */
};
export const toThemeType = (name: string): ThemeType => {
	if (name === CustomThemeLight || name === 'light') return ThemeType.Light;
	if (name === CustomThemeDark || name === 'dark') return ThemeType.Dark;
	if (name === 'system') return ThemeType.System;
	return ThemeType.Other;
};

export interface ISettingStoreAPI {
	setAppTheme: (theme: AppTheme) => Promise<void>;
	getAuthKey: (type: AuthKeyType, keyName: AuthKeyName) => Promise<AuthKey>;
	deleteAuthKey: (type: AuthKeyType, keyName: AuthKeyName) => Promise<void>;
	setAuthKey: (type: AuthKeyType, keyName: AuthKeyName, secret: string) => Promise<void>;
	getSettings: (forceFetch?: boolean) => Promise<SettingsSchema>;
}
