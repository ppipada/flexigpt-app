import type { ReasoningParam } from '@/spec/inference';

type ModelName = string;
export type ModelDisplayName = string;
type ModelSlug = string;
export type ModelPresetID = string;

export type ProviderName = string;
export type ProviderDisplayName = string;

export enum ProviderSDKType {
	ProviderSDKTypeAnthropic = 'providerSDKTypeAnthropic',
	ProviderSDKTypeOpenAIChatCompletions = 'providerSDKTypeOpenAI',
	ProviderSDKTypeOpenAIResponses = 'providerSDKTypeOpenAIResponses',
}

export const SDK_DISPLAY_NAME: Record<ProviderSDKType, string> = {
	[ProviderSDKType.ProviderSDKTypeAnthropic]: 'Anthropic Messages API',
	[ProviderSDKType.ProviderSDKTypeOpenAIChatCompletions]: 'OpenAI ChatCompletions API',
	[ProviderSDKType.ProviderSDKTypeOpenAIResponses]: 'OpenAI Responses API',
};

export const SDK_DEFAULTS: Record<
	ProviderSDKType,
	{ chatPath: string; apiKeyHeaderKey: string; defaultHeaders: Record<string, string> }
> = {
	[ProviderSDKType.ProviderSDKTypeAnthropic]: {
		chatPath: '/v1/messages',
		apiKeyHeaderKey: 'x-api-key',
		defaultHeaders: {
			'Content-Type': 'application/json',
			'anthropic-version': '2023-06-01',
		},
	},
	[ProviderSDKType.ProviderSDKTypeOpenAIChatCompletions]: {
		chatPath: '/v1/chat/completions',
		apiKeyHeaderKey: 'Authorization',
		defaultHeaders: {
			'Content-Type': 'application/json',
		},
	},
	[ProviderSDKType.ProviderSDKTypeOpenAIResponses]: {
		chatPath: '/v1/responses',
		apiKeyHeaderKey: 'Authorization',
		defaultHeaders: {
			'Content-Type': 'application/json',
		},
	},
};

export interface PutModelPresetPayload {
	name: ModelName;
	slug: ModelSlug;
	displayName: ModelDisplayName;
	isEnabled: boolean;
	stream?: boolean;
	maxPromptLength?: number;
	maxOutputLength?: number;
	temperature?: number;
	reasoning?: ReasoningParam;
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
	sdkType: ProviderSDKType;
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
