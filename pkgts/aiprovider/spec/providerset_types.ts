import { ChatCompletionRequestMessage, CompletionRequest, CompletionResponse } from './chat_types';
import { ModelName, ProviderName } from './provider_types';

export interface IProviderSetAPI {
	getDefaultProvider(): Promise<ProviderName>;
	setDefaultProvider(provider: ProviderName): Promise<void>;
	getConfigurationInfo(): Promise<Record<string, any>>;
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
		inputParams?: { [key: string]: any }
	): Promise<CompletionRequest>;
	completion(
		provider: ProviderName,
		input: CompletionRequest,
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined>;
}
