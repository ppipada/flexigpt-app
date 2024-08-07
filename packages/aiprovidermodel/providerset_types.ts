import { ChatCompletionRequestMessage, CompletionRequest, CompletionResponse } from './chat_types';
import { ModelName, ProviderInfo, ProviderName } from './provider_types';

export interface IProviderSetAPI {
	getDefaultProvider(): Promise<ProviderName>;
	setDefaultProvider(provider: ProviderName): Promise<void>;
	getProviderInfo(provider: ProviderName): Promise<ProviderInfo>;
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
	setAttribute(
		provider: ProviderName,
		apiKey?: string,
		defaultModel?: ModelName,
		defaultTemperature?: number,
		defaultOrigin?: string
	): Promise<void>;
}
