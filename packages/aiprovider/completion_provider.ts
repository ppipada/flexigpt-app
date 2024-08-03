import { APICaller } from './api_fetch';
import { ChatCompletionRequestMessage, CompletionRequest, CompletionResponse } from './chat_types';
import { ALL_MODEL_INFO } from './provider_consts';
import { ModelName, ProviderInfo } from './provider_types';
import { getCompletionRequest } from './provider_utils';

export interface CompletionProvider {
	getProviderInfo(): ProviderInfo;
	getCompletion(
		prompt: string,
		messages?: Array<ChatCompletionRequestMessage>,
		inputParams?: { [key: string]: any }
	): Promise<CompletionResponse | undefined>;
	completion(input: CompletionRequest): Promise<CompletionResponse | undefined>;
	setAttribute(apiKey?: string, defaultModel?: ModelName, defaultTemperature?: number, defaultOrigin?: string): void;
}

export class AIAPI implements CompletionProvider {
	protected providerInfo: ProviderInfo;
	protected apicaller: APICaller;

	constructor(providerInfo: ProviderInfo) {
		this.apicaller = new APICaller(
			providerInfo.defaultOrigin,
			providerInfo.apiKey,
			providerInfo.apiKeyHeaderKey,
			providerInfo.timeout,
			providerInfo.defaultHeaders
		);
		this.providerInfo = providerInfo;
	}

	getProviderInfo(): ProviderInfo {
		return this.providerInfo;
	}

	setAttribute(apiKey?: string, defaultModel?: ModelName, defaultTemperature?: number, defaultOrigin?: string) {
		if (apiKey) {
			this.apicaller.apiKey = apiKey;
			this.providerInfo.apiKey = apiKey;
		}
		if (defaultOrigin) {
			this.apicaller.origin = defaultOrigin;
			this.providerInfo.defaultOrigin = defaultOrigin;
		}
		if (defaultModel) {
			this.providerInfo.defaultModel = defaultModel;
		}
		if (defaultTemperature) {
			this.providerInfo.defaultTemperature = defaultTemperature;
		}
	}
	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	async completion(input: CompletionRequest): Promise<CompletionResponse | undefined> {
		return undefined;
	}

	async getCompletion(
		prompt: string,
		messages?: Array<ChatCompletionRequestMessage>,
		inputParams?: { [key: string]: any }
	): Promise<CompletionResponse | undefined> {
		const creq = getCompletionRequest(
			this.providerInfo.defaultModel,
			prompt,
			this.providerInfo.defaultTemperature,
			ALL_MODEL_INFO[this.providerInfo.defaultModel].maxPromptLength,
			messages,
			inputParams
		);
		return this.completion(creq);
	}
}
