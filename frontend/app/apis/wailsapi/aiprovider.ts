import {
	FetchCompletion,
	GetConfigurationInfo,
	MakeCompletion,
	SetDefaultProvider,
	SetProviderAttribute,
} from '@/apis/wailsjs/go/main/ProviderSetWrapper';
import type { spec as wailsSpec } from '@/apis/wailsjs/go/models';
import { EventsOn } from '@/apis/wailsjs/runtime/runtime';
import type {
	ChatCompletionRequestMessage,
	CompletionRequest,
	CompletionResponse,
	ConfigurationResponse,
	IProviderSetAPI,
	ModelName,
	ModelParams,
	ProviderInfo,
	ProviderName,
} from '@/models/aiprovidermodel';

/**
 * @public
 */
export class WailsProviderSetAPI implements IProviderSetAPI {
	async setDefaultProvider(provider: ProviderName): Promise<void> {
		const req = { Body: { provider: provider } };
		await SetDefaultProvider(req as wailsSpec.SetDefaultProviderRequest);
	}

	async getConfigurationInfo(): Promise<ConfigurationResponse> {
		const resp = await GetConfigurationInfo({} as wailsSpec.GetConfigurationInfoRequest);
		const configInfo = resp.Body || {};
		if (!('configuredProviders' in configInfo) || !('defaultProvider' in configInfo)) {
			return { defaultProvider: '', configuredProviders: {} };
		}

		const providerInfoDict: Record<ProviderName, ProviderInfo> = {};
		for (const providerInfo of configInfo['configuredProviders'] as ProviderInfo[]) {
			providerInfoDict[providerInfo.name] = providerInfo;
		}
		return { defaultProvider: configInfo['defaultProvider'] as ProviderName, configuredProviders: providerInfoDict };
	}

	async setAttribute(
		provider: ProviderName,
		apiKey?: string,
		defaultModel?: ModelName,
		origin?: string
	): Promise<void> {
		const req = {
			Provider: provider,
			Body: {
				apiKey: apiKey,
				defaultModel: defaultModel,
				origin: origin,
			},
		};
		await SetProviderAttribute(req as wailsSpec.SetProviderAttributeRequest);
	}
	async getCompletionRequest(
		provider: ProviderName,
		prompt: string,
		modelParams: ModelParams,
		prevMessages?: Array<ChatCompletionRequestMessage>
	): Promise<CompletionRequest> {
		const msgs: wailsSpec.ChatCompletionRequestMessage[] = [];
		if (prevMessages) {
			for (const m of prevMessages) {
				msgs.push(m as wailsSpec.ChatCompletionRequestMessage);
			}
		}

		const req = {
			Provider: provider,
			Body: {
				prompt: prompt,
				modelParams: modelParams,
				prevMessages: msgs,
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
