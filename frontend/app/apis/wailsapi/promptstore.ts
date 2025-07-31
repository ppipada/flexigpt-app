import type {
	IPromptStoreAPI,
	MessageBlock,
	PreProcessorCall,
	PromptBundle,
	PromptTemplate,
	PromptTemplateListItem,
	PromptVariable,
} from '@/spec/prompt';

import {
	DeletePromptBundle,
	DeletePromptTemplate,
	GetPromptTemplate,
	ListPromptBundles,
	ListPromptTemplates,
	PatchPromptBundle,
	PatchPromptTemplate,
	PutPromptBundle,
	PutPromptTemplate,
	SearchPromptTemplates,
} from '../wailsjs/go/main/PromptTemplateStoreWrapper';
import type { spec } from '../wailsjs/go/models';

export class WailsPromptStoreAPI implements IPromptStoreAPI {
	// --- Bundle Operations ---

	async listPromptBundles(
		bundleIDs?: string[],
		includeDisabled?: boolean,
		pageSize?: number,
		pageToken?: string
	): Promise<{ promptBundles: PromptBundle[]; nextPageToken?: string }> {
		const req = {
			BundleIDs: bundleIDs,
			IncludeDisabled: includeDisabled,
			PageSize: pageSize,
			PageToken: pageToken,
		};
		const resp = await ListPromptBundles(req as spec.ListPromptBundlesRequest);
		return {
			promptBundles: (resp.Body?.promptBundles ?? []) as PromptBundle[],
			nextPageToken: resp.Body?.nextPageToken ?? undefined,
		};
	}

	async putPromptBundle(
		bundleID: string,
		slug: string,
		displayName: string,
		isEnabled: boolean,
		description?: string
	): Promise<void> {
		const req = {
			BundleID: bundleID,
			Body: {
				slug: slug,
				displayName: displayName,
				isEnabled: isEnabled,
				description: description,
			} as spec.PutPromptBundleRequestBody,
		};
		await PutPromptBundle(req as spec.PutPromptBundleRequest);
	}

	async patchPromptBundle(bundleID: string, isEnabled: boolean): Promise<void> {
		const req = {
			BundleID: bundleID,
			Body: {
				isEnabled: isEnabled,
			},
		};
		await PatchPromptBundle(req as spec.PatchPromptBundleRequest);
	}

	async deletePromptBundle(bundleID: string): Promise<void> {
		const req: spec.DeletePromptBundleRequest = {
			BundleID: bundleID,
		};
		await DeletePromptBundle(req);
	}

	// --- Template Operations ---

	async listPromptTemplates(
		bundleIDs?: string[],
		tags?: string[],
		includeDisabled?: boolean,
		pageSize?: number,
		pageToken?: string
	): Promise<{ promptTemplateListItems: PromptTemplateListItem[]; nextPageToken?: string }> {
		const req = {
			BundleIDs: bundleIDs,
			Tags: tags,
			IncludeDisabled: includeDisabled,
			RecommendedPageSize: pageSize,
			PageToken: pageToken,
		};
		const resp = await ListPromptTemplates(req as spec.ListPromptTemplatesRequest);
		return {
			promptTemplateListItems: (resp.Body?.promptTemplateListItems ?? []) as PromptTemplateListItem[],
			nextPageToken: resp.Body?.nextPageToken ?? undefined,
		};
	}

	async searchPromptTemplates(
		query: string,
		pageToken?: string,
		pageSize?: number,
		includeDisabled?: boolean
	): Promise<{ promptTemplateListItems: PromptTemplateListItem[]; nextPageToken?: string }> {
		const req = {
			Query: query,
			PageToken: pageToken,
			PageSize: pageSize,
			IncludeDisabled: includeDisabled,
		};
		const resp = await SearchPromptTemplates(req as spec.SearchPromptTemplatesRequest);
		return {
			promptTemplateListItems: (resp.Body?.promptTemplateListItems ?? []) as PromptTemplateListItem[],
			nextPageToken: resp.Body?.nextPageToken ?? undefined,
		};
	}

	async putPromptTemplate(
		bundleID: string,
		templateSlug: string,
		displayName: string,
		isEnabled: boolean,
		blocks: MessageBlock[],
		version: string,
		description?: string,
		tags?: string[],
		variables?: PromptVariable[],
		preProcessors?: PreProcessorCall[]
	): Promise<void> {
		const req = {
			BundleID: bundleID,
			TemplateSlug: templateSlug,
			Version: version,
			Body: {
				displayName: displayName,
				isEnabled: isEnabled,
				description: description,
				tags: tags,
				blocks: blocks,
				variables: variables,
				preProcessors: preProcessors,
			},
		};
		await PutPromptTemplate(req as spec.PutPromptTemplateRequest);
	}

	async patchPromptTemplate(
		bundleID: string,
		templateSlug: string,
		version: string,
		isEnabled: boolean
	): Promise<void> {
		const req = {
			BundleID: bundleID,
			TemplateSlug: templateSlug,
			Version: version,
			Body: {
				isEnabled: isEnabled,
			},
		};
		await PatchPromptTemplate(req as spec.PatchPromptTemplateRequest);
	}

	async deletePromptTemplate(bundleID: string, templateSlug: string, version: string): Promise<void> {
		const req = {
			BundleID: bundleID,
			TemplateSlug: templateSlug,
			Version: version,
		};
		await DeletePromptTemplate(req as spec.DeletePromptTemplateRequest);
	}

	async getPromptTemplate(
		bundleID: string,
		templateSlug: string,
		version: string
	): Promise<PromptTemplate | undefined> {
		const req = {
			BundleID: bundleID,
			TemplateSlug: templateSlug,
			Version: version,
		};
		const resp = await GetPromptTemplate(req as spec.GetPromptTemplateRequest);
		return resp.Body as PromptTemplate | undefined;
	}
}
