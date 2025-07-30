// --- Enums ---
/**
 * @public
 */
export enum PromptRoleEnum {
	System = 'system',
	Developer = 'developer',
	User = 'user',
	Assistant = 'assistant',
}

/**
 * @public
 */
export enum VarType {
	String = 'string',
	Number = 'number',
	Boolean = 'boolean',
	Enum = 'enum',
	Date = 'date',
}

/**
 * @public
 */
export enum VarSource {
	User = 'user',
	Static = 'static',
	Tool = 'tool',
}

/**
 * @public
 */
export enum PreProcessorOnError {
	OnErrorEmpty = 'empty',
	OnErrorFail = 'fail',
}

// --- Core Types ---

export interface MessageBlock {
	id: string;
	role: PromptRoleEnum;
	content: string;
}

export interface PromptVariable {
	name: string;
	type: VarType;
	required: boolean;
	source: VarSource;
	description?: string;
	staticVal?: string;
	toolID?: string;
	enumValues?: string[];
	default?: string;
}

export interface PreProcessorCall {
	id: string;
	toolID: string;
	args?: Record<string, any>;
	saveAs: string;
	pathExpr?: string;
	onError?: PreProcessorOnError;
}

export interface PromptTemplate {
	id: string;
	displayName: string;
	slug: string;
	isEnabled: boolean;
	description?: string;
	tags?: string[];
	blocks: MessageBlock[];
	variables?: PromptVariable[];
	preProcessors?: PreProcessorCall[];
	version: string;
	createdAt: string;
	modifiedAt: string;
	isBuiltIn: boolean;
}

export interface PromptBundle {
	id: string;
	slug: string;
	displayName?: string;
	description?: string;
	isEnabled: boolean;
	createdAt: string;
	modifiedAt: string;
	isBuiltIn: boolean;
}

export interface PromptTemplateListItem {
	bundleID: string;
	bundleSlug: string;
	templateSlug: string;
	templateVersion: string;
	isBuiltIn: boolean;
}

// --- API Interface ---

export interface IPromptStoreAPI {
	// --- Bundle Operations ---

	/** List bundles, optionally filtered by IDs, disabled, and paginated. */
	listPromptBundles(
		bundleIDs?: string[],
		includeDisabled?: boolean,
		pageSize?: number,
		pageToken?: string
	): Promise<{ promptBundles: PromptBundle[]; nextPageToken?: string }>;

	/** Create or update a bundle. */
	putPromptBundle(
		bundleID: string,
		slug: string,
		displayName: string,
		isEnabled: boolean,
		description?: string
	): Promise<void>;

	/** Patch (enable/disable) a bundle. */
	patchPromptBundle(bundleID: string, isEnabled: boolean): Promise<void>;

	/** Delete a bundle. */
	deletePromptBundle(bundleID: string): Promise<void>;

	// --- Template Operations ---

	/** List templates, optionally filtered by bundleIDs, tags, etc. */
	listPromptTemplates(
		bundleIDs?: string[],
		tags?: string[],
		includeDisabled?: boolean,
		recommendedPageSize?: number,
		pageToken?: string
	): Promise<{ promptTemplateListItems: PromptTemplateListItem[]; nextPageToken?: string }>;

	/** Search templates by query. */
	searchPromptTemplates(
		query: string,
		pageToken?: string,
		pageSize?: number,
		includeDisabled?: boolean
	): Promise<{ promptTemplateListItems: PromptTemplateListItem[]; nextPageToken?: string }>;

	/** Create or update a template. */
	putPromptTemplate(
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
	): Promise<void>;

	/** Patch (enable/disable) a template version. */
	patchPromptTemplate(bundleID: string, templateSlug: string, version: string, isEnabled: boolean): Promise<void>;

	/** Delete a template version. */
	deletePromptTemplate(bundleID: string, templateSlug: string, version: string): Promise<void>;

	/** Get a template version (or latest if version omitted). */
	getPromptTemplate(bundleID: string, templateSlug: string, version: string): Promise<PromptTemplate | undefined>;
}
