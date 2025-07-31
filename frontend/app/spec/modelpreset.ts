// --- Type Aliases & Enums ---

type ModelName = string;
type ModelDisplayName = string;
type ModelSlug = string;
export type ModelPresetID = string;

export type ProviderName = string;
export const DefaultProviderName: ProviderName = 'openai';
type ProviderDisplayName = string;
type ProviderAPIType =
	| 'inbuiltAnthropicCompatible'
	| 'inbuiltHuggingFaceCompatible'
	| 'inbuiltOpenAICompatible'
	| 'customOpenAICompatible';

export enum ReasoningType {
	HybridWithTokens = 'hybridWithTokens',
	SingleWithLevels = 'singleWithLevels',
}

export enum ReasoningLevel {
	Low = 'low',
	Medium = 'medium',
	High = 'high',
}

// --- Core Shared Interfaces ---

export interface ReasoningParams {
	type: ReasoningType;
	level: ReasoningLevel;
	tokens: number;
}

export interface ModelParams {
	name: ModelName;
	stream: boolean;
	maxPromptLength: number;
	maxOutputLength: number;
	temperature?: number;
	reasoning?: ReasoningParams;
	systemPrompt: string;
	timeout: number;
	additionalParametersRawJSON?: string;
}

export interface ProviderInfo {
	name: ProviderName;
	apiKey: string;
	origin: string;
	chatCompletionPathPrefix: string;
	apiKeyHeaderKey: string;
	defaultHeaders: Record<string, string>;
}

export interface ChatOptions extends ModelParams {
	id: string;
	title: string;
	provider: ProviderName;
	disablePreviousMessages: boolean;
}

// --- Default Values ---

export const DefaultModelParams: ModelParams = {
	name: '',
	stream: false,
	maxPromptLength: 2048,
	maxOutputLength: 1024,
	temperature: 0.1,
	reasoning: {
		type: ReasoningType.SingleWithLevels,
		level: ReasoningLevel.Medium,
		tokens: 1024,
	},
	systemPrompt: '',
	timeout: 60,
	additionalParametersRawJSON: undefined,
};

export const DefaultChatOptions: ChatOptions = {
	...DefaultModelParams,
	id: 'NoModel',
	provider: 'No Provider',
	name: 'No Model',
	title: 'No Model configured',
	disablePreviousMessages: false,
};

export const ProviderInfoDescription = {
	apiKey: 'Your API key for the provider.',
	origin: 'Origin/URL to use for requests. This can be used to talk to any server that serves a compatible API',
	timeout: 'The timeout duration in milliseconds.',
	apiKeyHeaderKey: 'The header key for the API key.',
	defaultHeaders: 'The default headers to be included in requests.',
	chatCompletionPathPrefix: 'The path prefix for chat completions.',
	modelPrefixes: 'Optional prefixes for models.',
};

// --- API Payload Types ---

export interface PutProviderPresetPayload {
	name: ProviderName;
	displayName: ProviderDisplayName;
	apiType: ProviderAPIType;
	isEnabled: boolean;
	origin: string;
	chatCompletionPathPrefix: string;
	apiKeyHeaderKey: string;
	defaultHeaders: Record<string, string>;
}

export interface PutModelPresetPayload {
	name: ModelName;
	slug: ModelSlug;
	displayName: ModelDisplayName;
	isEnabled: boolean;
	stream?: boolean;
	maxPromptLength?: number;
	maxOutputLength?: number;
	temperature?: number;
	reasoning?: ReasoningParams;
	systemPrompt?: string;
	timeout?: number;
	additionalParametersRawJSON?: string;
}

// --- API Response Types ---

export interface ModelPreset extends PutModelPresetPayload {
	id: ModelPresetID;
}

export const DefaultModelPreset: ModelPreset = {
	id: '',
	name: '',
	displayName: '',
	slug: '',
	isEnabled: true,
	stream: DefaultModelParams.stream,
	maxPromptLength: DefaultModelParams.maxPromptLength,
	maxOutputLength: DefaultModelParams.maxOutputLength,
	temperature: DefaultModelParams.temperature ?? 0.1,
	reasoning: DefaultModelParams.reasoning,
	systemPrompt: DefaultModelParams.systemPrompt,
	timeout: DefaultModelParams.timeout,
	additionalParametersRawJSON: DefaultModelParams.additionalParametersRawJSON,
};

export interface ProviderPreset extends PutProviderPresetPayload {
	schemaVersion: string;
	createdAt: string; // ISO date string
	modifiedAt: string; // ISO date string
	isBuiltIn: boolean;
	defaultModelPresetID: ModelPresetID;
	modelPresets: Record<ModelPresetID, ModelPreset>;
}

export interface IModelPresetStoreAPI {
	putProviderPreset(providerName: ProviderName, payload: PutProviderPresetPayload): Promise<void>;

	patchProviderPreset(
		providerName: ProviderName,
		isEnabled?: boolean,
		defaultModelPresetID?: ModelPresetID
	): Promise<void>;

	deleteProviderPreset(providerName: ProviderName): Promise<void>;

	putModelPreset(
		providerName: ProviderName,
		modelPresetID: ModelPresetID,
		payload: PutModelPresetPayload
	): Promise<void>;

	patchModelPreset(providerName: ProviderName, modelPresetID: ModelPresetID, isEnabled: boolean): Promise<void>;

	deleteModelPreset(providerName: ProviderName, modelPresetID: ModelPresetID): Promise<void>;

	listProviderPresets(
		names?: ProviderName[],
		includeDisabled?: boolean,
		pageSize?: number,
		pageToken?: string
	): Promise<{ providers: ProviderPreset[]; nextPageToken?: string }>;
}
