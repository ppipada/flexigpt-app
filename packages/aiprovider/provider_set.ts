import { AnthropicAPI } from './anthropic';
import { GoogleAPI } from './google';
import { HuggingFaceAPI } from './huggingface';
import { LlamaCPPAPI } from './llamacpp';
import { OpenAIAPI } from './openai';
import { ALL_MODEL_INFO, huggingfaceProviderInfo } from './provider_consts';
import { CompletionProvider, ModelName, ProviderName } from './provider_types';

export class ProviderSet {
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

	getDefaultProvider(): ProviderName {
		return this.defaultProvider;
	}

	setDefaultProvider(provider: ProviderName) {
		this.defaultProvider = provider;
	}

	getProviderAPI(provider: ProviderName): CompletionProvider {
		return this.providers[provider];
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

export const providerSet = new ProviderSet(ProviderName.OPENAI);
