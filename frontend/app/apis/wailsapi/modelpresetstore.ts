import type {
	IModelPresetStoreAPI,
	ModelPreset,
	ModelPresetID,
	PresetsSchema,
	ProviderName,
	ProviderPreset,
} from '@/models/aimodelmodel';

import {
	AddModelPreset,
	CreateProviderPreset,
	DeleteModelPreset,
	GetAllModelPresets,
	SetDefaultModelPreset,
} from '../wailsjs/go/main/ModelPresetStoreWrapper';
import type { spec } from '../wailsjs/go/models';

/**
 * @public
 */
export class WailsModelPresetStoreAPI implements IModelPresetStoreAPI {
	// Implement the getAllSettings method
	async getAllModelPresets(): Promise<PresetsSchema> {
		const r: spec.GetAllModelPresetsRequest = { ForceFetch: false };
		const s = await GetAllModelPresets(r);
		return s.Body as PresetsSchema;
	}

	async createProviderPreset(providerName: ProviderName, providerPreset: ProviderPreset): Promise<void> {
		const r = {
			ProviderName: providerName,
			Body: providerPreset,
		};
		await CreateProviderPreset(r as spec.CreateProviderPresetRequest);
	}

	async deleteProviderPreset(providerName: ProviderName): Promise<void> {
		const r = {
			ProviderName: providerName,
		};
		await DeleteModelPreset(r as spec.DeleteModelPresetRequest);
	}

	async addModelPreset(
		providerName: ProviderName,
		modelPresetID: ModelPresetID,
		modelPreset: ModelPreset
	): Promise<void> {
		const r = {
			ProviderName: providerName,
			ModelPresetID: modelPresetID,
			Body: modelPreset,
		};
		await AddModelPreset(r as spec.AddModelPresetRequest);
	}

	async deleteModelPreset(providerName: ProviderName, modelPresetID: ModelPresetID): Promise<void> {
		const r = {
			ProviderName: providerName,
			ModelPresetID: modelPresetID,
		};
		await DeleteModelPreset(r as spec.DeleteModelPresetRequest);
	}

	async setDefaultModelPreset(providerName: ProviderName, modelPresetID: ModelPresetID): Promise<void> {
		const r = {
			ProviderName: providerName,
			Body: {
				ModelPresetID: modelPresetID,
			},
		};
		await SetDefaultModelPreset(r as spec.SetDefaultModelPresetRequest);
	}
}
