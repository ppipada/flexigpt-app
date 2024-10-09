import {
	ChatCompletionRequestMessage,
	CompletionRequest,
	CompletionResponse,
	ModelName,
	ProviderInfo,
	ProviderName,
} from '@/models/aiprovidermodel';

export async function getDefaultProvider(): Promise<ProviderName> {
	const provider = await window.ProviderSetAPI.getDefaultProvider();
	// log.info('base api', JSON.stringify(provider));
	return provider;
}

export async function setDefaultProvider(provider: ProviderName): Promise<void> {
	return await window.ProviderSetAPI.setDefaultProvider(provider);
}

export async function getProviderInfo(provider: ProviderName): Promise<ProviderInfo> {
	return await window.ProviderSetAPI.getProviderInfo(provider);
}

export async function getConfigurationInfo(): Promise<Record<string, any>> {
	return await window.ProviderSetAPI.getConfigurationInfo();
}

export async function getCompletionRequest(
	provider: ProviderName,
	prompt: string,
	prevMessages?: Array<ChatCompletionRequestMessage>,
	inputParams?: { [key: string]: any },
	stream?: boolean
): Promise<CompletionRequest> {
	return await window.ProviderSetAPI.getCompletionRequest(provider, prompt, prevMessages, inputParams, stream);
}

export async function completion(
	provider: ProviderName,
	input: CompletionRequest,
	onStreamData?: (data: string) => Promise<void>
): Promise<CompletionResponse | undefined> {
	return await window.ProviderSetAPI.completion(provider, input, onStreamData);
}

export async function setAttribute(
	provider: ProviderName,
	apiKey?: string,
	defaultModel?: ModelName,
	defaultTemperature?: number,
	defaultOrigin?: string
): Promise<void> {
	return await window.ProviderSetAPI.setAttribute(provider, apiKey, defaultModel, defaultTemperature, defaultOrigin);
}
