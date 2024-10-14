/* eslint-disable @typescript-eslint/no-unused-vars */
import {
	FetchCompletion,
	GetCompletionRequest,
	GetConfigurationInfo,
	GetDefaultProvider,
	GetProviderInfo,
	SetDefaultProvider,
	SetProviderAttribute,
} from '@/backendapibase/wailsjs/go/main/WrappedProviderSetAPI';
import { spec as wailsSpec } from '@/backendapibase/wailsjs/go/models';
import { EventsOn } from '@/backendapibase/wailsjs/runtime/runtime';
import {
	ChatCompletionRequestMessage,
	CompletionRequest,
	CompletionResponse,
	IProviderSetAPI,
	ModelName,
	ProviderInfo,
	ProviderName,
} from '@/models/aiprovidermodel';

export class WailsProviderSetAPI implements IProviderSetAPI {
	async getDefaultProvider(): Promise<ProviderName> {
		const provider = await GetDefaultProvider();
		return provider;
	}

	async setDefaultProvider(provider: ProviderName): Promise<void> {
		await SetDefaultProvider(provider);
	}

	async getProviderInfo(provider: ProviderName): Promise<ProviderInfo> {
		const resp = await GetProviderInfo(provider);
		return resp as ProviderInfo;
	}

	async getConfigurationInfo(): Promise<Record<string, any>> {
		const resp = await GetConfigurationInfo();
		return resp;
	}

	async getCompletionRequest(
		provider: ProviderName,
		prompt: string,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		inputParams?: { [key: string]: any },
		stream?: boolean
	): Promise<CompletionRequest> {
		const msgs: wailsSpec.ChatCompletionRequestMessage[] = [];
		if (prevMessages) {
			for (const m of prevMessages) {
				msgs.push(m as wailsSpec.ChatCompletionRequestMessage);
			}
		}
		if (!inputParams) {
			inputParams = {};
		}
		const resp = await GetCompletionRequest(provider, prompt, msgs, inputParams, stream || false);
		return resp as CompletionRequest;
		// const x: CompletionRequest = { model: 'x', temperature: 0, maxPromptLength: 2000, stream: false };
		// return x;
	}

	// Need an eventflow for getting completion.
	// Implemented that in main App Wrapper than aiprovider go package.
	// Wrapper redirects to providerSet after doing event handling
	async completion(
		provider: ProviderName,
		input: CompletionRequest,
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined> {
		const callbackId = `stream-data-callback-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
		if (onStreamData) {
			EventsOn(callbackId, onStreamData);
		}
		const response = await FetchCompletion(provider, input as wailsSpec.CompletionRequest, callbackId);
		return response;
	}

	async setAttribute(
		provider: ProviderName,
		apiKey?: string,
		defaultModel?: ModelName,
		defaultTemperature?: number,
		defaultOrigin?: string
	): Promise<void> {
		await SetProviderAttribute(provider, apiKey, defaultModel, defaultTemperature, defaultOrigin);
	}
}
