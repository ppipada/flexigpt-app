export interface ReasoningParam {
	type: string;
	level: string;
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

export interface CacheControlEphemeral {
	ttl: string;
}

export interface CacheControl {
	kind: string;
	cacheControlEphemeral?: CacheControlEphemeral;
}

export interface CitationConfig {
	enabled: boolean;
}

export interface URLCitation {
	url: string;
	title: string;
	citedText: string;
	startIndex: number;
	endIndex: number;
	encryptedIndex: string;
}

export interface Citation {
	kind: string;
	// Go type: URLCitationOpenAIResponses
	urlCitationOpenAIResponses?: any;
	// Go type: URLCitationAnthropicMessages
	urlCitationAnthropicMessages?: any;
}

export interface ContentItemRefusal {
	refusal: string;
}

export interface ContentItemFile {
	id: string;
	fileName: string;
	fileMIME: string;
	fileURL: string;
	fileData: string;
	additionalContext: string;
	citationConfig?: CitationConfig;
}

export interface ContentItemImage {
	id: string;
	detail: string;
	imageName: string;
	imageMIME: string;
	imageURL: string;
	imageData: string;
}

export interface ContentItemText {
	text: string;
	citations?: Citation[];
}
export interface InferenceError {
	code: string;
	message: string;
}

export interface InputOutputContentItemUnion {
	kind: string;
	textItem?: ContentItemText;
	refusalItem?: ContentItemRefusal;
	imageItem?: ContentItemImage;
	fileItem?: ContentItemFile;
}

export interface InputOutputContent {
	id: string;
	role: string;
	status: string;
	cacheControl?: CacheControl;
	contents?: InputOutputContentItemUnion[];
}

export interface Usage {
	inputTokensTotal: number;
	inputTokensCached: number;
	inputTokensUncached: number;
	outputTokens: number;
	reasoningTokens: number;
}

export interface ToolOutputItemUnion {
	kind: string;
	textItem?: ContentItemText;
	imageItem?: ContentItemImage;
	fileItem?: ContentItemFile;
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
	kind: string;
	searchItem?: WebSearchToolCallSearch;
	openPageItem?: WebSearchToolCallOpenPage;
	findItem?: WebSearchToolCallFind;
}
export interface WebSearchToolChoiceItemUserLocation {
	city: string;
	country: string;
	region: string;
	timezone: string;
}
export interface WebSearchToolChoiceItem {
	max_uses: number;
	searchContextSize: string;
	allowed_domains: string[];
	blocked_domains: string[];
	user_location?: WebSearchToolChoiceItemUserLocation;
}
export interface WebSearchToolOutputError {
	code: string;
}
export interface WebSearchToolOutputSearch {
	url: string;
	title: string;
	encryptedContent: string;
	renderedContent: string;
	page_age: string;
}
export interface WebSearchToolOutputItemUnion {
	kind: string;
	searchItem?: WebSearchToolOutputSearch;
	errorItem?: WebSearchToolOutputError;
}
