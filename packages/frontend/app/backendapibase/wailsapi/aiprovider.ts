/* eslint-disable @typescript-eslint/no-unused-vars */
import {
	ChatCompletionRequestMessage,
	CompletionRequest,
	CompletionResponse,
	IProviderSetAPI,
	ModelName,
	openaiProviderInfo,
	ProviderInfo,
	ProviderName,
} from '@/models/aiprovidermodel';

export class WailsProviderSetAPI implements IProviderSetAPI {
	// Implement the getDefaultProvider method
	async getDefaultProvider(): Promise<ProviderName> {
		// const provider = await window.ProviderSetAPI.getDefaultProvider();
		// return provider;
		return ProviderName.OPENAI;
	}

	// Implement the setDefaultProvider method
	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	async setDefaultProvider(provider: ProviderName): Promise<void> {
		// await window.ProviderSetAPI.setDefaultProvider(provider);
	}

	// Implement the getProviderInfo method
	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	async getProviderInfo(provider: ProviderName): Promise<ProviderInfo> {
		// return await window.ProviderSetAPI.getProviderInfo(provider);
		return openaiProviderInfo;
	}

	// Implement the getConfigurationInfo method
	async getConfigurationInfo(): Promise<Record<string, any>> {
		// return await window.ProviderSetAPI.getConfigurationInfo();
		return { x: 'y' };
	}

	// Implement the getCompletionRequest method
	async getCompletionRequest(
		provider: ProviderName,
		prompt: string,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		inputParams?: { [key: string]: any },
		stream?: boolean
	): Promise<CompletionRequest> {
		// return await window.ProviderSetAPI.getCompletionRequest(provider, prompt, prevMessages, inputParams, stream);
		const x: CompletionRequest = { model: 'x', temperature: 0, maxPromptLength: 2000, stream: false };
		return x;
	}

	// Implement the completion method
	async completion(
		provider: ProviderName,
		input: CompletionRequest,
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined> {
		// return await window.ProviderSetAPI.completion(provider, input, onStreamData);
		return undefined;
	}

	// Implement the setAttribute method
	async setAttribute(
		provider: ProviderName,
		apiKey?: string,
		defaultModel?: ModelName,
		defaultTemperature?: number,
		defaultOrigin?: string
	): Promise<void> {
		// await window.ProviderSetAPI.setAttribute(provider, apiKey, defaultModel, defaultTemperature, defaultOrigin);
	}
}
