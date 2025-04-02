import {
	FetchCompletion,
	GetConfigurationInfo,
	GetDefaultProvider,
	MakeCompletion,
	SetDefaultProvider,
	SetProviderAttribute,
} from '@/backendapibase/wailsjs/go/main/ProviderSetWrapper';
import type { spec as wailsSpec } from '@/backendapibase/wailsjs/go/models';
import { EventsOn } from '@/backendapibase/wailsjs/runtime/runtime';
import type {
	ChatCompletionRequestMessage,
	CompletionRequest,
	CompletionResponse,
	IProviderSetAPI,
	ModelName,
	ProviderName,
} from '@/models/aiprovidermodel';

export class WailsProviderSetAPI implements IProviderSetAPI {
	async getDefaultProvider(): Promise<ProviderName> {
		const resp = await GetDefaultProvider({} as wailsSpec.GetDefaultProviderRequest);
		return resp.Body?.defaultProvider as ProviderName;
	}

	async setDefaultProvider(provider: ProviderName): Promise<void> {
		const req = { Body: { provider: provider } };
		await SetDefaultProvider(req as wailsSpec.SetDefaultProviderRequest);
	}

	async getConfigurationInfo(): Promise<Record<string, any>> {
		const resp = await GetConfigurationInfo({} as wailsSpec.GetConfigurationInfoRequest);
		return resp.Body as Record<string, any>;
	}

	async setAttribute(
		provider: ProviderName,
		apiKey?: string,
		defaultModel?: ModelName,
		defaultTemperature?: number,
		defaultOrigin?: string
	): Promise<void> {
		const req = {
			Provider: provider,
			Body: {
				apiKey: apiKey,
				defaultModel: defaultModel,
				defaultTemperature: defaultTemperature,
				defaultOrigin: defaultOrigin,
			},
		};
		await SetProviderAttribute(req as wailsSpec.SetProviderAttributeRequest);
	}
	async getCompletionRequest(
		provider: ProviderName,
		prompt: string,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		inputParams?: { [key: string]: any }
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
		const req = {
			Provider: provider,
			Body: {
				prompt: prompt,
				prevMessages: msgs,
				inputParams: inputParams,
			},
		};
		const resp = await MakeCompletion(req as wailsSpec.MakeCompletionRequest);
		return resp.Body as CompletionRequest;
		// const x: CompletionRequest = { model: 'x', temperature: 0, maxPromptLength: 2000, stream: false };
		// return x;
	}

	// Need an eventflow for getting completion.
	// Implemented that in main App Wrapper than aiprovider go package.
	// Wrapper redirects to providerSet after doing event handling
	async completion(
		provider: ProviderName,
		input: CompletionRequest,
		onStreamData?: (data: string) => void
	): Promise<CompletionResponse | undefined> {
		const callbackId = `stream-data-callback-${Date.now().toString()}-${Math.random().toString(36).substring(2, 9)}`;
		if (onStreamData) {
			EventsOn(callbackId, onStreamData);
		}
		const response = await FetchCompletion(provider, input as wailsSpec.CompletionRequest, callbackId);
		return response.Body as CompletionResponse;
	}
}
