export enum ToolImplType {
	Go = 'go',
	HTTP = 'http',
}

export enum ToolImplOutputKind {
	Text = 'text',
	Blob = 'blob',
	None = 'none',
}

/**
 * @public
 *
 */
export type JSONValue = string | number | boolean | null | JSONValue[] | { [key: string]: JSONValue };
export type JSONSchema = JSONValue;

export type JSONRawString = string;

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
	successCodes?: number[]; // default: 2xx
	errorMode?: string; // "fail"(dflt) | "empty"
}

/**
 * @public
 */
export interface HTTPToolImpl {
	request: HTTPRequest;
	response: HTTPResponse;
}

export enum ToolStoreChoiceType {
	Function = 'function',
	Custom = 'custom',
	WebSearch = 'webSearch',
}

export interface ToolStoreChoice {
	choiceID: string;
	bundleID: string;
	bundleSlug?: string;

	toolID?: string;
	toolSlug: string;
	toolVersion: string;
	toolType: ToolStoreChoiceType;
	displayName?: string;
	description?: string;
}

export interface Tool {
	schemaVersion: string;

	id: string;
	slug: string;
	version: string;

	displayName: string;
	description?: string;
	tags?: string[];

	userCallable: boolean;
	llmCallable: boolean;
	outputKind: ToolImplOutputKind;

	argSchema: JSONSchema;

	type: ToolImplType;
	goImpl?: GoToolImpl;
	httpImpl?: HTTPToolImpl;

	isEnabled: boolean;
	isBuiltIn: boolean;
	createdAt: string;
	modifiedAt: string;
}

export interface ToolBundle {
	schemaVersion: string;

	id: string;
	slug: string;

	displayName?: string;
	description?: string;
	isEnabled: boolean;
	isBuiltIn: boolean;
	createdAt: string;
	modifiedAt: string;
	softDeletedAt?: string;
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
	output: JSONRawString;
	meta?: Record<string, any>;
	isBuiltIn: boolean;
	isError?: boolean;
	errorMessage?: string;
}

export interface UIToolStoreChoice extends ToolStoreChoice {
	selectionID: string;
}

/**
 * @public
 * Status for a tool-call chip in the composer/history.
 */
export type UIToolCallStatus = 'pending' | 'running' | 'succeeded' | 'failed' | 'discarded';

/**
 * UI representation of a tool call (for chips).
 * `type` is the inference ToolType (`function` | `custom` | `webSearch`),
 * but kept as string here to avoid a circular import.
 */
export interface UIToolCall {
	id: string;
	callID: string;
	name: string;
	arguments?: string;
	webSearchToolCallItems?: any;
	type: ToolStoreChoiceType;
	choiceID: string;
	status: UIToolCallStatus;
	toolStoreChoice: ToolStoreChoice;
	errorMessage?: string;
}

export interface UIToolOutput {
	id: string;
	callID: string;
	name: string;

	type: ToolStoreChoiceType;
	choiceID: string;

	/** Short human-readable label used in chips. */
	summary: string;

	/** Raw JSON/text output as returned by the tool store. */
	rawOutput: JSONRawString;

	toolStoreChoice: ToolStoreChoice;

	isError?: boolean;
	errorMessage?: string;

	arguments?: string;
	webSearchToolCallItems?: any;
}
