/**
 * @public
 */
export enum ToolType {
	Go = 'go',
	HTTP = 'http',
}

/**
 * @public
 */
export type JSONSchema = any; // You may want to use a stricter type or a library

/**
 * @public
 */
export interface GoToolImpl {
	/** Fully-qualified registration key, e.g. "github.com/acme/flexigpt/tools.Weather" */
	func: string;
}

/**
 * @public
 */
export interface HTTPAuth {
	type: string;
	in?: string; // "header" | "query" (apiKey only)
	name?: string;
	valueTemplate: string; // may contain ${SECRET}
}

/**
 * @public
 */
export interface HTTPRequest {
	method?: string; // default "GET"
	urlTemplate: string; // http(s)://… may contain ${var}
	query?: Record<string, string>; // k:${var}
	headers?: Record<string, string>; // k:${var}
	body?: string; // raw or template
	auth?: HTTPAuth;
	timeoutMs?: number; // default 10_000
}

/**
 * @public
 */
export interface HTTPResponse {
	successCodes?: number[]; // default: any 2xx
	encoding?: string; // "json"(dflt) | "text"
	selector?: string; // JSONPath / JMESPath / regexp
	errorMode?: string; // "fail"(dflt) | "empty"
}

/**
 * @public
 */
export interface HTTPToolImpl {
	request: HTTPRequest;
	response: HTTPResponse;
}

export interface Tool {
	id: string;
	slug: string;
	version: string;
	displayName: string;
	description?: string;
	tags?: string[];
	argSchema: JSONSchema;
	outputSchema: JSONSchema;
	type: ToolType;
	goImpl?: GoToolImpl;
	httpImpl?: HTTPToolImpl;
	isEnabled: boolean;
	isBuiltIn: boolean;
	createdAt: string;
	modifiedAt: string;
	schemaVersion: string;
}

export interface ToolBundle {
	id: string;
	slug: string;
	displayName?: string;
	description?: string;
	isEnabled: boolean;
	isBuiltIn: boolean;
	createdAt: string;
	modifiedAt: string;
	softDeletedAt?: string;
	schemaVersion: string;
}

export interface ToolListItem {
	bundleID: string;
	bundleSlug: string;
	toolSlug: string;
	toolVersion: string;
	isBuiltIn: boolean;
}

export interface InvokeHTTPOptions {
	timeoutMs?: number;
	extraHeaders?: Record<string, string>;
	secrets?: Record<string, string>;
}

export interface InvokeGoOptions {
	timeoutMs?: number;
}

export interface InvokeToolResponse {
	output: any;
	meta?: Record<string, any>;
	isBuiltIn: boolean;
}
export interface IToolStoreAPI {
	// --- Bundle Operations ---

	/** List tool bundles, optionally filtered by IDs, disabled, and paginated. */
	listToolBundles(
		bundleIDs?: string[],
		includeDisabled?: boolean,
		pageSize?: number,
		pageToken?: string
	): Promise<{ toolBundles: ToolBundle[]; nextPageToken?: string }>;

	/** Create or update a tool bundle. */
	putToolBundle(
		bundleID: string,
		slug: string,
		displayName: string,
		isEnabled: boolean,
		description?: string
	): Promise<void>;

	/** Patch (enable/disable) a tool bundle. */
	patchToolBundle(bundleID: string, isEnabled: boolean): Promise<void>;

	/** Delete a tool bundle. */
	deleteToolBundle(bundleID: string): Promise<void>;

	// --- Tool Operations ---

	/** List tools, optionally filtered by bundleIDs, tags, etc. */
	listTools(
		bundleIDs?: string[],
		tags?: string[],
		includeDisabled?: boolean,
		recommendedPageSize?: number,
		pageToken?: string
	): Promise<{ toolListItems: ToolListItem[]; nextPageToken?: string }>;

	/** Search tools by query. */
	searchTools(
		query: string,
		pageToken?: string,
		pageSize?: number,
		includeDisabled?: boolean
	): Promise<{ toolListItems: ToolListItem[]; nextPageToken?: string }>;

	/** Create or update a tool. */
	putTool(
		bundleID: string,
		toolSlug: string,
		version: string,
		displayName: string,
		isEnabled: boolean,
		argSchema: JSONSchema,
		outputSchema: JSONSchema,
		type: ToolType,
		goImpl?: GoToolImpl,
		httpImpl?: HTTPToolImpl,
		description?: string,
		tags?: string[]
	): Promise<void>;

	/** Patch (enable/disable) a tool version. */
	patchTool(bundleID: string, toolSlug: string, version: string, isEnabled: boolean): Promise<void>;

	/** Delete a tool version. */
	deleteTool(bundleID: string, toolSlug: string, version: string): Promise<void>;

	/** Invoke a tool version. */
	invokeTool(
		bundleID: string,
		toolSlug: string,
		version: string,
		args?: Record<string, any>,
		httpOptions?: InvokeHTTPOptions,
		goOptions?: InvokeGoOptions
	): Promise<InvokeToolResponse>;

	/** Get a tool version. */
	getTool(bundleID: string, toolSlug: string, version: string): Promise<Tool | undefined>;
}
