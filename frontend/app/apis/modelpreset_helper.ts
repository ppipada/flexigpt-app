import { type ProviderName, type ProviderPreset } from '@/spec/modelpreset';

import { modelPresetStoreAPI } from '@/apis/baseapi';

export async function getAllProviderPresetsMap(
	includeDisabled?: boolean
): Promise<Record<ProviderName, ProviderPreset>> {
	let pageToken: string | undefined = undefined;
	const result: Record<ProviderName, ProviderPreset> = {};
	let pageCount = 0;
	const MAX_PAGES = 20;

	do {
		const { providers, nextPageToken } = await modelPresetStoreAPI.listProviderPresets(
			undefined,
			includeDisabled,
			undefined,
			pageToken
		);
		for (const preset of providers) {
			result[preset.name] = preset;
		}
		pageToken = nextPageToken;
		pageCount++;
		if (pageCount >= MAX_PAGES) break;
	} while (pageToken);

	return result;
}
