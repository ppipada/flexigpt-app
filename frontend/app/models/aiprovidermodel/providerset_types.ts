import { ChatCompletionRequestMessage, CompletionRequest, CompletionResponse } from './chat_types';
import { ModelName, ProviderInfo, ProviderName } from './provider_types';

export interface IProviderSetAPI {
	getDefaultProvider(): Promise<ProviderName>;
	setDefaultProvider(provider: ProviderName): Promise<void>;
	getConfigurationInfo(): Promise<Record<string, any>>;
	getProviderInfo(provider: ProviderName): Promise<ProviderInfo>;
	setAttribute(
		provider: ProviderName,
		apiKey?: string,
		defaultModel?: ModelName,
		defaultTemperature?: number,
		defaultOrigin?: string
	): Promise<void>;

	getCompletionRequest(
		provider: ProviderName,
		prompt: string,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		inputParams?: { [key: string]: any },
		stream?: boolean
	): Promise<CompletionRequest>;
	completion(
		provider: ProviderName,
		input: CompletionRequest,
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined>;
}
