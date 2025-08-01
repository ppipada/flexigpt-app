export type ModelName = string;
export type ModelDisplayName = string;
type ModelSlug = string;
export type ModelPresetID = string;

export type ProviderName = string;
export type ProviderDisplayName = string;
export enum ProviderAPIType {
	AnthropicCompatible = 'providerAPITypeAnthropicCompatible',
	HuggingFaceCompatible = 'providerAPITypeHuggingFaceCompatible',
	OpenAICompatible = 'providerAPITypeOpenAICompatible',
}

export enum ReasoningType {
	HybridWithTokens = 'hybridWithTokens',
	SingleWithLevels = 'singleWithLevels',
}

export enum ReasoningLevel {
	Low = 'low',
	Medium = 'medium',
	High = 'high',
}

export interface ReasoningParams {
	type: ReasoningType;
	level: ReasoningLevel;
	tokens: number;
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

export interface ModelPreset extends PutModelPresetPayload {
	id: ModelPresetID;
	isBuiltIn: boolean;
}

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
export interface ProviderPreset extends PutProviderPresetPayload {
	isBuiltIn: boolean;
	defaultModelPresetID: ModelPresetID;
	modelPresets: Record<ModelPresetID, ModelPreset>;
}

export interface IModelPresetStoreAPI {
	getDefaultProvider(): Promise<ProviderName>;

	patchDefaultProvider(providerName: ProviderName): Promise<void>;

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
