import type { ToolStoreChoice } from '@/spec/tool';

export enum RoleEnum {
	System = 'system',
	Developer = 'developer',
	User = 'user',
	Assistant = 'assistant',
	Function = 'function',
	Tool = 'tool',
}

export enum ReasoningType {
	HybridWithTokens = 'hybridWithTokens',
	SingleWithLevels = 'singleWithLevels',
}

export enum ReasoningLevel {
	None = 'none',
	Minimal = 'minimal',
	Low = 'low',
	Medium = 'medium',
	High = 'high',
	XHigh = 'xhigh',
}

export enum Status {
	InProgress = 'inProgress',
	Completed = 'completed',
	Incomplete = 'incomplete',
	Failed = 'failed',
	Cancelled = 'cancelled',
	Queued = 'queued',
	Searching = 'searching',
}

export interface ReasoningParam {
	type: ReasoningType;
	level: ReasoningLevel;
	tokens: number;
}

export interface ModelParam {
	name: string;
	stream: boolean;
	maxPromptLength: number;
	maxOutputLength: number;
	temperature?: number;
	reasoning?: ReasoningParam;
	systemPrompt: string;
	timeout: number;
	additionalParametersRawJSON?: string;
}

export interface InferenceError {
	code: string;
	message: string;
}

export interface InferenceUsage {
	inputTokensTotal: number;
	inputTokensCached: number;
	inputTokensUncached: number;
	outputTokens: number;
	reasoningTokens: number;
}

export enum CacheControlKind {
	Ephemeral = 'ephemeral',
}

export interface CacheControlEphemeral {
	ttl?: string;
}

export interface CacheControl {
	kind: CacheControlKind;
	cacheControlEphemeral?: CacheControlEphemeral;
}

export enum CitationKind {
	URL = 'urlCitation',
}

export interface CitationConfig {
	enabled: boolean;
}

export interface URLCitation {
	url: string;
	title?: string;
	citedText?: string;
	startIndex?: number;
	endIndex?: number;
	encryptedIndex?: string;
}

export interface Citation {
	kind: CitationKind;
	urlCitation?: URLCitation;
}

export enum ContentItemKind {
	Text = 'text',
	Image = 'image',
	File = 'file',
	Refusal = 'refusal',
}

export interface ContentItemRefusal {
	refusal: string;
}

export interface ContentItemFile {
	id?: string;
	fileName?: string;
	fileMIME?: string;
	fileURL?: string;
	fileData?: string;
	additionalContext?: string;
	citationConfig?: CitationConfig;
}

export enum ImageDetail {
	High = 'high',
	Low = 'low',
	Auto = 'auto',
}

export interface ContentItemImage {
	id?: string;
	detail?: ImageDetail;
	imageName?: string;
	imageMIME?: string;
	imageURL?: string;
	imageData?: string;
}

export interface ContentItemText {
	text: string;
	citations?: Citation[];
}

export interface InputOutputContentItemUnion {
	kind: ContentItemKind;
	textItem?: ContentItemText;
	refusalItem?: ContentItemRefusal;
	imageItem?: ContentItemImage;
	fileItem?: ContentItemFile;
}

export interface InputOutputContent {
	id: string;
	role: RoleEnum;
	status?: Status;
	cacheControl?: CacheControl;
	contents?: InputOutputContentItemUnion[];
}

export interface ReasoningContent {
	id: string;
	role: RoleEnum;
	status?: Status;
	cacheControl?: CacheControl;
	signature?: string;
	summary?: string[];
	thinking?: string[];
	redactedThinking?: string[];
	encryptedContent?: string[];
}

export interface ToolOutputItemUnion {
	kind: ContentItemKind;
	textItem?: ContentItemText;
	imageItem?: ContentItemImage;
	fileItem?: ContentItemFile;
}

export enum WebSearchToolCallKind {
	Search = 'search',
	OpenPage = 'openPage',
	Find = 'find',
}

export interface WebSearchToolCallFind {
	url: string;
	pattern: string;
}
export interface WebSearchToolCallOpenPage {
	url: string;
}
export interface WebSearchToolCallSearchSource {
	url: string;
}
export interface WebSearchToolCallSearch {
	query: string;
	sources?: WebSearchToolCallSearchSource[];
	input?: Record<string, any>;
}
export interface WebSearchToolCallItemUnion {
	kind: WebSearchToolCallKind;
	searchItem?: WebSearchToolCallSearch;
	openPageItem?: WebSearchToolCallOpenPage;
	findItem?: WebSearchToolCallFind;
}
export interface WebSearchToolChoiceItemUserLocation {
	city?: string;
	country?: string;
	region?: string;
	timezone?: string;
}
export interface WebSearchToolChoiceItem {
	max_uses?: number;
	searchContextSize?: string;
	allowed_domains?: string[];
	blocked_domains?: string[];
	user_location?: WebSearchToolChoiceItemUserLocation;
}

export enum WebSearchToolOutputKind {
	Search = 'search',
	Error = 'error',
}

export interface WebSearchToolOutputError {
	code: string;
}
export interface WebSearchToolOutputSearch {
	url: string;
	title?: string;
	encryptedContent?: string;
	renderedContent?: string;
	page_age?: string;
}
export interface WebSearchToolOutputItemUnion {
	kind: WebSearchToolOutputKind;
	searchItem?: WebSearchToolOutputSearch;
	errorItem?: WebSearchToolOutputError;
}

export enum ToolType {
	Function = 'function',
	Custom = 'custom',
	WebSearch = 'webSearch',
}
export interface ToolCall {
	type: ToolType;
	choiceID: string;
	id: string;
	role: RoleEnum;
	status?: Status;
	cacheControl?: CacheControl;
	callID: string;
	name: string;
	arguments?: string;
	webSearchToolCallItems?: WebSearchToolCallItemUnion[];
}

export interface ToolOutput {
	type: ToolType;
	choiceID: string;
	id: string;
	role: RoleEnum;
	status?: Status;
	cacheControl?: CacheControl;
	callID: string;
	name: string;
	isError: boolean;
	signature?: string;
	contents?: ToolOutputItemUnion[];
	webSearchToolOutputItems?: WebSearchToolOutputItemUnion[];
}

export interface ToolChoice {
	type: ToolType;
	id: string;
	cacheControl?: CacheControl;
	name: string;
	description?: string;
	arguments?: Record<string, any>;
	webSearchArguments?: WebSearchToolChoiceItem;
}

export enum InputKind {
	InputMessage = 'inputMessage',
	OutputMessage = 'outputMessage',
	ReasoningMessage = 'reasoningMessage',
	FunctionToolCall = 'functionToolCall',
	FunctionToolOutput = 'functionToolOutput',
	CustomToolCall = 'customToolCall',
	CustomToolOutput = 'customToolOutput',
	WebSearchToolCall = 'webSearchToolCall',
	WebSearchToolOutput = 'webSearchToolOutput',
}

export enum OutputKind {
	OutputMessage = 'outputMessage',
	ReasoningMessage = 'reasoningMessage',
	FunctionToolCall = 'functionToolCall',
	CustomToolCall = 'customToolCall',
	WebSearchToolCall = 'webSearchToolCall',
	WebSearchToolOutput = 'webSearchToolOutput',
}

export interface InputUnion {
	kind: InputKind;
	inputMessage?: InputOutputContent;
	outputMessage?: InputOutputContent;
	reasoningMessage?: ReasoningContent;
	functionToolCall?: ToolCall;
	functionToolOutput?: ToolOutput;
	customToolCall?: ToolCall;
	customToolOutput?: ToolOutput;
	webSearchToolCall?: ToolCall;
	webSearchToolOutput?: ToolOutput;
}
export interface OutputUnion {
	kind: OutputKind;
	outputMessage?: InputOutputContent;
	reasoningMessage?: ReasoningContent;
	functionToolCall?: ToolCall;
	customToolCall?: ToolCall;
	webSearchToolCall?: ToolCall;
	webSearchToolOutput?: ToolOutput;
}

export interface FetchCompletionResponse {
	outputs?: OutputUnion[];
	usage?: InferenceUsage;
	error?: InferenceError;
	debugDetails?: any;
}

export interface ToolCallBinding {
	choiceID: string;
	// ToolStoreChoice is the original tool-store handle the UI supplied.
	toolStoreChoice: ToolStoreChoice;
}

export interface CompletionResponseBody {
	inferenceResponse?: FetchCompletionResponse;
	hydratedCurrentInputs?: InputUnion[];
	toolCallBindings?: ToolCallBinding[];
}
