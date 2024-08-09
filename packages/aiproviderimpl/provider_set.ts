import {
	ALL_MODEL_INFO,
	ChatCompletionRequestMessage,
	CompletionRequest,
	CompletionResponse,
	huggingfaceProviderInfo,
	IProviderSetAPI,
	ModelName,
	ProviderInfo,
	ProviderName,
} from 'aiprovidermodel';

import { AnthropicAPI } from './anthropic';
import { CompletionProvider } from './completion_provider';
import { GoogleAPI } from './google';
import { HuggingFaceAPI } from './huggingface';
import { LlamaCPPAPI } from './llamacpp';
import { OpenAIAPI } from './openai';

export class ProviderSet implements IProviderSetAPI {
	private defaultProvider: ProviderName;
	private providers: { [key in ProviderName]: CompletionProvider };

	constructor(defaultProvider: ProviderName) {
		this.defaultProvider = defaultProvider;
		this.providers = {
			[ProviderName.ANTHROPIC]: new AnthropicAPI(),
			[ProviderName.GOOGLE]: new GoogleAPI(),
			[ProviderName.HUGGINGFACE]: new HuggingFaceAPI(),
			[ProviderName.LLAMACPP]: new LlamaCPPAPI(),
			[ProviderName.OPENAI]: new OpenAIAPI(),
		};
	}

	async getDefaultProvider(): Promise<ProviderName> {
		return this.defaultProvider;
	}

	async setDefaultProvider(provider: ProviderName): Promise<void> {
		this.defaultProvider = provider;
	}

	async getProviderInfo(provider: ProviderName): Promise<ProviderInfo> {
		return this.providers[provider].getProviderInfo();
	}

	async setAttribute(
		provider: ProviderName,
		apiKey?: string,
		defaultModel?: ModelName,
		defaultTemperature?: number,
		defaultOrigin?: string
	): Promise<void> {
		return this.providers[provider].setAttribute(apiKey, defaultModel, defaultTemperature, defaultOrigin);
	}

	async getCompletionRequest(
		provider: ProviderName,
		prompt: string,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		inputParams?: { [key: string]: any },
		stream?: boolean
	): Promise<CompletionRequest> {
		let completionProvider = provider;
		if (inputParams && 'provider' in inputParams) {
			const providerStr = inputParams['provider'];
			if (providerStr in ProviderName) {
				completionProvider = inputParams['provider'] as ProviderName;
			}
			delete inputParams['provider'];
		}

		return this.providers[completionProvider].getCompletionRequest(prompt, prevMessages, inputParams, stream);
	}

	async completion(
		provider: ProviderName,
		input: CompletionRequest,
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined> {
		return this.providers[provider].completion(input, onStreamData);
	}

	getProviderAPIUsingModelStr(model: string): CompletionProvider | undefined {
		if (model in ALL_MODEL_INFO) {
			return this.providers[ALL_MODEL_INFO[model as ModelName].provider];
		}

		// didnt find the model name, lets try prefixes
		const huggingfaceModelPrefixes = huggingfaceProviderInfo.modelPrefixes;
		if (
			huggingfaceModelPrefixes &&
			huggingfaceModelPrefixes.some(search => model.startsWith(search)) &&
			this.providers.huggingface
		) {
			return this.providers.huggingface;
		}
		// No provider was given and input model didnt match any known models, but has slash in it, so assume its a huggingface model
		if (model.includes('/')) {
			return this.providers.huggingface;
		}

		return undefined;
	}
}
