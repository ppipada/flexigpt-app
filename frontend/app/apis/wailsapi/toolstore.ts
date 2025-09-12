import type {
	GoToolImpl,
	HTTPToolImpl,
	InvokeGoOptions,
	InvokeHTTPOptions,
	InvokeToolResponse,
	IToolStoreAPI,
	JSONRawString,
	Tool,
	ToolBundle,
	ToolListItem,
	ToolType,
} from '@/spec/tool';

import {
	DeleteTool,
	DeleteToolBundle,
	GetTool,
	InvokeTool,
	ListToolBundles,
	ListTools,
	PatchTool,
	PatchToolBundle,
	PutTool,
	PutToolBundle,
	SearchTools,
} from '@/apis/wailsjs/go/main/ToolStoreWrapper';
import type { spec } from '@/apis/wailsjs/go/models';

export class WailsToolStoreAPI implements IToolStoreAPI {
	// --- Bundle Operations ---

	async listToolBundles(
		bundleIDs?: string[],
		includeDisabled?: boolean,
		pageSize?: number,
		pageToken?: string
	): Promise<{ toolBundles: ToolBundle[]; nextPageToken?: string }> {
		const req = {
			BundleIDs: bundleIDs,
			IncludeDisabled: includeDisabled,
			PageSize: pageSize,
			PageToken: pageToken,
		};
		const resp = await ListToolBundles(req as spec.ListToolBundlesRequest);
		return {
			toolBundles: (resp.Body?.toolBundles ?? []) as ToolBundle[],
			nextPageToken: resp.Body?.nextPageToken ?? undefined,
		};
	}

	async putToolBundle(
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
			} as spec.PutToolBundleRequestBody,
		};
		await PutToolBundle(req as spec.PutToolBundleRequest);
	}

	async patchToolBundle(bundleID: string, isEnabled: boolean): Promise<void> {
		const req = {
			BundleID: bundleID,
			Body: {
				isEnabled: isEnabled,
			},
		};
		await PatchToolBundle(req as spec.PatchToolBundleRequest);
	}

	async deleteToolBundle(bundleID: string): Promise<void> {
		const req: spec.DeleteToolBundleRequest = {
			BundleID: bundleID,
		};
		await DeleteToolBundle(req);
	}

	// --- Tool Operations ---

	async listTools(
		bundleIDs?: string[],
		tags?: string[],
		includeDisabled?: boolean,
		pageSize?: number,
		pageToken?: string
	): Promise<{ toolListItems: ToolListItem[]; nextPageToken?: string }> {
		const req = {
			BundleIDs: bundleIDs,
			Tags: tags,
			IncludeDisabled: includeDisabled,
			RecommendedPageSize: pageSize,
			PageToken: pageToken,
		};
		const resp = await ListTools(req as spec.ListToolsRequest);
		return {
			toolListItems: (resp.Body?.toolListItems ?? []) as ToolListItem[],
			nextPageToken: resp.Body?.nextPageToken ?? undefined,
		};
	}

	async searchTools(
		query: string,
		pageToken?: string,
		pageSize?: number,
		includeDisabled?: boolean
	): Promise<{ toolListItems: ToolListItem[]; nextPageToken?: string }> {
		const req = {
			Query: query,
			PageToken: pageToken,
			PageSize: pageSize,
			IncludeDisabled: includeDisabled,
		};
		const resp = await SearchTools(req as spec.SearchToolsRequest);
		return {
			toolListItems: (resp.Body?.toolListItems ?? []) as ToolListItem[],
			nextPageToken: resp.Body?.nextPageToken ?? undefined,
		};
	}

	async putTool(
		bundleID: string,
		toolSlug: string,
		version: string,
		displayName: string,
		isEnabled: boolean,
		argSchema: JSONRawString,
		outputSchema: JSONRawString,
		type: ToolType,
		goImpl?: GoToolImpl,
		httpImpl?: HTTPToolImpl,
		description?: string,
		tags?: string[]
	): Promise<void> {
		const req = {
			BundleID: bundleID,
			ToolSlug: toolSlug,
			Version: version,
			Body: {
				displayName: displayName,
				isEnabled: isEnabled,
				description: description,
				tags: tags,
				argSchema: argSchema,
				outputSchema: outputSchema,
				type: type,
				goImpl: goImpl,
				httpImpl: httpImpl,
			} as spec.PutToolRequestBody,
		};
		await PutTool(req as spec.PutToolRequest);
	}

	async patchTool(bundleID: string, toolSlug: string, version: string, isEnabled: boolean): Promise<void> {
		const req = {
			BundleID: bundleID,
			ToolSlug: toolSlug,
			Version: version,
			Body: {
				isEnabled: isEnabled,
			},
		};
		await PatchTool(req as spec.PatchToolRequest);
	}

	async deleteTool(bundleID: string, toolSlug: string, version: string): Promise<void> {
		const req: spec.DeleteToolRequest = {
			BundleID: bundleID,
			ToolSlug: toolSlug,
			Version: version,
		};
		await DeleteTool(req);
	}

	async invokeTool(
		bundleID: string,
		toolSlug: string,
		version: string,
		args?: JSONRawString,
		httpOptions?: InvokeHTTPOptions,
		goOptions?: InvokeGoOptions
	): Promise<any> {
		const req = {
			BundleID: bundleID,
			ToolSlug: toolSlug,
			Version: version,
			Body: {
				args: args ?? {},
				httpOptions: httpOptions as spec.InvokeHTTPOptions,
				goOptions: goOptions as spec.InvokeGoOptions,
			} as spec.InvokeToolRequestBody,
		} as spec.InvokeToolRequest;
		const resp = await InvokeTool(req);
		return resp.Body as InvokeToolResponse;
	}

	async getTool(bundleID: string, toolSlug: string, version: string): Promise<Tool | undefined> {
		const req: spec.GetToolRequest = {
			BundleID: bundleID,
			ToolSlug: toolSlug,
			Version: version,
		};
		const resp = await GetTool(req);
		return resp.Body as Tool | undefined;
	}
}
