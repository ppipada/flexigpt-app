import type {
	AddProviderRequest,
	ChatCompletionRequestMessage,
	CompletionResponse,
	ConfigurationResponse,
	IProviderSetAPI,
} from '@/models/aiprovidermodel';
import type { ModelParams, ProviderInfo, ProviderName, ProviderPreset } from '@/models/modelpresetsmodel';

import {
	AddProvider,
	DeleteProvider,
	FetchCompletion,
	GetConfigurationInfo,
	SetDefaultProvider,
	SetProviderAPIKey,
	SetProviderAttribute,
} from '@/apis/wailsjs/go/main/ProviderSetWrapper';
import type { inference as wailsInference, spec as wailsModelSpec } from '@/apis/wailsjs/go/models';
import { EventsOn } from '@/apis/wailsjs/runtime/runtime';

/**
 * @public
 */
export class WailsProviderSetAPI implements IProviderSetAPI {
	async setDefaultProvider(provider: ProviderName): Promise<void> {
		const req = { Body: { provider: provider } };
		await SetDefaultProvider(req as wailsInference.SetDefaultProviderRequest);
	}

	async getConfigurationInfo(): Promise<ConfigurationResponse> {
		const resp = await GetConfigurationInfo({} as wailsInference.GetConfigurationInfoRequest);
		const configInfo = resp.Body || {};
		if (
			!('configuredProviders' in configInfo) ||
			!('defaultProvider' in configInfo) ||
			!('inbuiltProviderModels' in configInfo)
		) {
			return {
				defaultProvider: '',
				configuredProviders: {},
				inbuiltProviderModels: {},
			};
		}

		const providerInfoDict: Record<ProviderName, ProviderInfo> = {};
		for (const providerInfo of configInfo['configuredProviders'] as ProviderInfo[]) {
			providerInfoDict[providerInfo.name] = providerInfo;
		}
		return {
			defaultProvider: configInfo['defaultProvider'] as ProviderName,
			configuredProviders: providerInfoDict,
			inbuiltProviderModels: configInfo['inbuiltProviderModels'] as Record<ProviderName, ProviderPreset>,
		};
	}

	async addProvider(providerInfo: AddProviderRequest): Promise<void> {
		const req = {
			Provider: providerInfo.provider,
			Body: {
				apiKey: providerInfo.apiKey,
				origin: providerInfo.origin,
				chatCompletionPathPrefix: providerInfo.chatCompletionPathPrefix,
			},
		};
		await AddProvider(req as wailsInference.AddProviderRequest);
	}

	async deleteProvider(provider: ProviderName): Promise<void> {
		const req = {
			Provider: provider,
		};
		await DeleteProvider(req as wailsInference.DeleteProviderRequest);
	}

	async setProviderAPIKey(provider: ProviderName, apiKey: string): Promise<void> {
		const req = {
			Provider: provider,
			Body: {
				apiKey: apiKey,
			},
		};
		await SetProviderAPIKey(req as wailsInference.SetProviderAPIKeyRequest);
	}

	async setProviderAttribute(
		provider: ProviderName,
		origin?: string,
		chatCompletionPathPrefix?: string
	): Promise<void> {
		const req = {
			Provider: provider,
			Body: {
				origin: origin,
				chatCompletionPathPrefix: chatCompletionPathPrefix,
			},
		};
		await SetProviderAttribute(req as wailsInference.SetProviderAttributeRequest);
	}
	// Need an eventflow for getting completion.
	// Implemented that in main App Wrapper than aiprovider go package.
	// Wrapper redirects to providerSet after doing event handling
	async completion(
		provider: ProviderName,
		prompt: string,
		modelParams: ModelParams,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		onStreamData?: (data: string) => void
	): Promise<CompletionResponse | undefined> {
		const callbackId = `stream-data-callback-${Date.now().toString()}-${Math.random().toString(36).substring(2, 9)}`;
		let prevData: string = '';
		if (onStreamData) {
			const cb = (data: string) => {
				if (data !== prevData) {
					prevData = data;
					onStreamData(data);
				}
			};
			EventsOn(callbackId, cb);
		}
		const response = await FetchCompletion(
			provider,
			prompt,
			modelParams as wailsModelSpec.ModelParams,
			prevMessages ? ([...prevMessages] as wailsInference.ChatCompletionRequestMessage[]) : [],
			callbackId
		);
		return response.Body as CompletionResponse;
	}
}
