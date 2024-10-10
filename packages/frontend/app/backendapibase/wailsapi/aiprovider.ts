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
	// Implement the getDefaultProvider method
	async getDefaultProvider(): Promise<ProviderName> {
		const provider = await window.ProviderSetAPI.getDefaultProvider();
		return provider;
	}

	// Implement the setDefaultProvider method
	async setDefaultProvider(provider: ProviderName): Promise<void> {
		await window.ProviderSetAPI.setDefaultProvider(provider);
	}

	// Implement the getProviderInfo method
	async getProviderInfo(provider: ProviderName): Promise<ProviderInfo> {
		return await window.ProviderSetAPI.getProviderInfo(provider);
	}

	// Implement the getConfigurationInfo method
	async getConfigurationInfo(): Promise<Record<string, any>> {
		return await window.ProviderSetAPI.getConfigurationInfo();
	}

	// Implement the getCompletionRequest method
	async getCompletionRequest(
		provider: ProviderName,
		prompt: string,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		inputParams?: { [key: string]: any },
		stream?: boolean
	): Promise<CompletionRequest> {
		return await window.ProviderSetAPI.getCompletionRequest(provider, prompt, prevMessages, inputParams, stream);
	}

	// Implement the completion method
	async completion(
		provider: ProviderName,
		input: CompletionRequest,
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined> {
		return await window.ProviderSetAPI.completion(provider, input, onStreamData);
	}

	// Implement the setAttribute method
	async setAttribute(
		provider: ProviderName,
		apiKey?: string,
		defaultModel?: ModelName,
		defaultTemperature?: number,
		defaultOrigin?: string
	): Promise<void> {
		await window.ProviderSetAPI.setAttribute(provider, apiKey, defaultModel, defaultTemperature, defaultOrigin);
	}
}
