import type {
	AppTheme,
	AuthKey,
	AuthKeyMeta,
	AuthKeyName,
	AuthKeyType,
	ISettingStoreAPI,
	SettingsSchema,
	ThemeType,
} from '@/spec/setting';

import {
	DeleteAuthKey,
	GetAuthKey,
	GetSettings,
	SetAppTheme,
	SetAuthKey,
} from '@/apis/wailsjs/go/main/SettingStoreWrapper';
import { type spec as wailsSpec } from '@/apis/wailsjs/go/models';

/**
 * @public
 */
export class WailsSettingStoreAPI implements ISettingStoreAPI {
	async setAppTheme(theme: AppTheme): Promise<void> {
		const r = {
			Body: {
				type: theme.type,
				name: theme.name,
			} as wailsSpec.SetAppThemeRequestBody,
		};
		await SetAppTheme(r as wailsSpec.SetAppThemeRequest);
	}

	async getAuthKey(type: AuthKeyType, keyName: AuthKeyName): Promise<AuthKey> {
		const r = {
			Type: type,
			KeyName: keyName,
		};
		const resp = await GetAuthKey(r as wailsSpec.GetAuthKeyRequest);
		return { secret: resp.Body?.secret ?? '', sha256: resp.Body?.sha256 ?? '', nonEmpty: resp.Body?.nonEmpty ?? false };
	}

	async deleteAuthKey(type: AuthKeyType, keyName: AuthKeyName): Promise<void> {
		const r = {
			Type: type,
			KeyName: keyName,
		};
		await DeleteAuthKey(r as wailsSpec.DeleteAuthKeyRequest);
	}

	async setAuthKey(type: AuthKeyType, keyName: AuthKeyName, secret: string): Promise<void> {
		const r = {
			Type: type,
			KeyName: keyName,
			Body: {
				secret: secret,
			},
		};
		await SetAuthKey(r as wailsSpec.SetAuthKeyRequest);
	}

	async getSettings(forceFetch?: boolean): Promise<SettingsSchema> {
		const r: wailsSpec.GetSettingsRequest = {
			ForceFetch: !!forceFetch,
		};
		const resp = await GetSettings(r);
		return {
			appTheme: {
				type: resp.Body?.appTheme.type as ThemeType,
				name: resp.Body?.appTheme.name ?? '',
			},
			authKeys: resp.Body?.authKeys as AuthKeyMeta[],
		};
	}
}
