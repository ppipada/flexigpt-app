import type {
	ModelPresetID,
	ProviderName,
	ProviderPreset,
	PutModelPresetPayload,
	PutProviderPresetPayload,
} from '@/spec/modelpreset';

import type { IModelPresetStoreAPI } from '@/apis/interface';
import {
	DeleteModelPreset,
	DeleteProviderPreset,
	GetDefaultProvider,
	ListProviderPresets,
	PatchDefaultProvider,
	PatchModelPreset,
	PatchProviderPreset,
	PutModelPreset,
	PutProviderPreset,
} from '@/apis/wailsjs/go/main/ModelPresetStoreWrapper';
import type { spec } from '@/apis/wailsjs/go/models';

/**
 * @public
 */
export class WailsModelPresetStoreAPI implements IModelPresetStoreAPI {
	async getDefaultProvider(): Promise<ProviderName> {
		const resp = await GetDefaultProvider({});
		return resp.Body?.DefaultProvider ?? '';
	}

	async patchDefaultProvider(providerName: ProviderName): Promise<void> {
		if (!providerName) throw new Error('Missing providerName or payload');
		const r = {
			Body: { defaultProvider: providerName } as spec.PatchDefaultProviderRequestBody,
		};
		await PatchDefaultProvider(r as spec.PatchDefaultProviderRequest);
	}

	async putProviderPreset(providerName: ProviderName, payload: PutProviderPresetPayload): Promise<void> {
		if (!providerName) throw new Error('Missing providerName or payload');
		const r = {
			ProviderName: providerName,
			Body: payload as spec.PutProviderPresetRequestBody,
		};
		await PutProviderPreset(r as spec.PutProviderPresetRequest);
	}

	async patchProviderPreset(
		providerName: ProviderName,
		isEnabled?: boolean,
		defaultModelPresetID?: ModelPresetID
	): Promise<void> {
		if (!providerName) throw new Error('Missing providerName');
		const r = {
			ProviderName: providerName,
			Body: {
				isEnabled: isEnabled ?? undefined,
				defaultModelPresetID: defaultModelPresetID ?? undefined,
			},
		};
		await PatchProviderPreset(r as spec.PatchProviderPresetRequest);
	}

	async deleteProviderPreset(providerName: ProviderName): Promise<void> {
		if (!providerName) throw new Error('Missing providerName');
		const r = {
			ProviderName: providerName,
		};
		await DeleteProviderPreset(r as spec.DeleteProviderPresetRequest);
	}

	async putModelPreset(
		providerName: ProviderName,
		modelPresetID: ModelPresetID,
		payload: PutModelPresetPayload
	): Promise<void> {
		if (!providerName || !modelPresetID) throw new Error('Missing arguments');
		const r = {
			ProviderName: providerName,
			ModelPresetID: modelPresetID,
			Body: payload,
		};
		await PutModelPreset(r as spec.PutModelPresetRequest);
	}

	async patchModelPreset(providerName: ProviderName, modelPresetID: ModelPresetID, isEnabled: boolean): Promise<void> {
		if (!providerName || !modelPresetID) throw new Error('Missing arguments');
		const r = {
			ProviderName: providerName,
			ModelPresetID: modelPresetID,
			Body: { isEnabled: isEnabled } as spec.PatchModelPresetRequestBody,
		} as spec.PatchModelPresetRequest;
		await PatchModelPreset(r);
	}

	async deleteModelPreset(providerName: ProviderName, modelPresetID: ModelPresetID): Promise<void> {
		if (!providerName || !modelPresetID) throw new Error('Missing arguments');
		const r = {
			ProviderName: providerName,
			ModelPresetID: modelPresetID,
		};
		await DeleteModelPreset(r as spec.DeleteModelPresetRequest);
	}

	async listProviderPresets(
		names?: ProviderName[],
		includeDisabled?: boolean,
		pageSize?: number,
		pageToken?: string
	): Promise<{ providers: ProviderPreset[]; nextPageToken?: string }> {
		const r: spec.ListProviderPresetsRequest = {
			Names: names ?? [],
			IncludeDisabled: includeDisabled ?? false,
			PageSize: pageSize ?? 256,
			PageToken: pageToken ?? '',
		};
		const resp = await ListProviderPresets(r);
		return {
			providers: (resp.Body?.providers ?? []) as ProviderPreset[],
			nextPageToken: resp.Body?.nextPageToken ?? undefined,
		};
	}
}
