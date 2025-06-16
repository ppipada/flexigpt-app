import type {
	IModelPresetStoreAPI,
	ModelName,
	ModelPreset,
	ModelPresetsSchema,
	ProviderModelPresets,
	ProviderName,
} from '@/models/aimodelmodel';

import {
	AddModelPreset,
	CreateModelPresets,
	DeleteModelPreset,
	GetAllModelPresets,
} from '../wailsjs/go/main/ModelPresetStoreWrapper';
import type { spec } from '../wailsjs/go/models';

/**
 * @public
 */
export class WailsModelPresetStoreAPI implements IModelPresetStoreAPI {
	// Implement the getAllSettings method
	async getAllModelPresets(): Promise<ModelPresetsSchema> {
		const r: spec.GetAllModelPresetsRequest = { ForceFetch: false };
		const s = await GetAllModelPresets(r);
		return s.Body as ModelPresetsSchema;
	}

	async createModelPresets(providerName: ProviderName, providerModelPresets: ProviderModelPresets): Promise<void> {
		const r = {
			ProviderName: providerName,
			Body: providerModelPresets,
		};
		await CreateModelPresets(r as spec.CreateModelPresetsRequest);
	}

	async deleteModelPresets(providerName: ProviderName): Promise<void> {
		const r = {
			ProviderName: providerName,
		};
		await DeleteModelPreset(r as spec.DeleteModelPresetRequest);
	}

	async addModelPreset(providerName: ProviderName, modelName: ModelName, modelPreset: ModelPreset): Promise<void> {
		const r = {
			ProviderName: providerName,
			ModelName: modelName,
			Body: modelPreset,
		};
		await AddModelPreset(r as spec.AddModelPresetRequest);
	}

	async deleteModelPreset(providerName: ProviderName, modelName: ModelName): Promise<void> {
		const r = {
			ProviderName: providerName,
			ModelName: modelName,
		};
		await DeleteModelPreset(r as spec.DeleteModelPresetRequest);
	}
}
