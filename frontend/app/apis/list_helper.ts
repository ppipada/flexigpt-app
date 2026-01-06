import type { ProviderName } from '@/spec/inference';
import type { ProviderPreset } from '@/spec/modelpreset';
import type { PromptBundle, PromptTemplateListItem } from '@/spec/prompt';
import type { ToolBundle, ToolListItem } from '@/spec/tool';

import { modelPresetStoreAPI, promptStoreAPI, toolStoreAPI } from '@/apis/baseapi';

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

export async function getAllPromptTemplates(
	bundleIDs?: string[],
	tags?: string[],
	includeDisabled?: boolean
): Promise<PromptTemplateListItem[]> {
	const all: PromptTemplateListItem[] = [];
	let pageToken: string | undefined;
	const recommendedPageSize = 25;
	do {
		const { promptTemplateListItems, nextPageToken } = await promptStoreAPI.listPromptTemplates(
			bundleIDs,
			tags,
			includeDisabled,
			recommendedPageSize,
			pageToken
		);

		all.push(...promptTemplateListItems);
		pageToken = nextPageToken;
	} while (pageToken);

	return all;
}

export async function getAllPromptBundles(bundleIDs?: string[], includeDisabled?: boolean): Promise<PromptBundle[]> {
	const all: PromptBundle[] = [];
	let pageToken: string | undefined;
	const pageSize = 25;
	do {
		const { promptBundles, nextPageToken } = await promptStoreAPI.listPromptBundles(
			bundleIDs,
			includeDisabled,
			pageSize,
			pageToken
		);

		all.push(...promptBundles);
		pageToken = nextPageToken;
	} while (pageToken);

	return all;
}

export async function getAllTools(
	bundleIDs?: string[],
	tags?: string[],
	includeDisabled?: boolean
): Promise<ToolListItem[]> {
	const all: ToolListItem[] = [];
	let pageToken: string | undefined;
	const recommendedPageSize = 25;
	do {
		const { toolListItems, nextPageToken } = await toolStoreAPI.listTools(
			bundleIDs,
			tags,
			includeDisabled,
			recommendedPageSize,
			pageToken
		);

		all.push(...toolListItems);
		pageToken = nextPageToken;
	} while (pageToken);

	return all;
}

export async function getAllToolBundles(bundleIDs?: string[], includeDisabled?: boolean): Promise<ToolBundle[]> {
	const all: ToolBundle[] = [];

	let pageToken: string | undefined;
	const pageSize = 25;
	do {
		const { toolBundles, nextPageToken } = await toolStoreAPI.listToolBundles(
			bundleIDs,
			includeDisabled,
			pageSize,
			pageToken
		);

		all.push(...toolBundles);
		pageToken = nextPageToken;
	} while (pageToken);

	return all;
}
