import { log } from 'logger';
import { CompletionProvider } from './chatapibase/chat_types';
import { ALL_AI_PROVIDERS } from './provider_consts';
import { ProviderInfo, ProviderName } from './provider_types';

export class ProviderSet {
	public defaultProvider: string;
	public providers: { [key: string]: CompletionProvider } = {};

	constructor(defaultProvider: string) {
		this.defaultProvider = defaultProvider;
	}

	public addProvider(name: string, provider: CompletionProvider | null) {
		if (!provider) {
			throw new Error('Provider cannot be null');
		}
		this.providers[name] = provider;
		if (!this.defaultProvider || this.defaultProvider === '') {
			this.defaultProvider = name;
		}
	}

	public setAttribute(name: string, attrName: string, attrValue: any) {
		if (!this.providers[name]) {
			log.error('No provider found for name:', name);
			return;
		}
		this.providers[name].setAttribute(attrName, attrValue);
	}

	public getProvider(model: string, providerName = ''): CompletionProvider {
		if (providerName && providerName !== '' && this.providers[providerName]) {
			return this.providers[providerName];
		}
		if (!model || model === '') {
			if (this.defaultProvider && this.defaultProvider !== '') {
				return this.providers[this.defaultProvider];
			}
			throw new Error('No default provider and No model as input');
		}

		const openAIModels = ['text-davinci-003, text-davinci-002, davinci, curie, babbage, ada', 'gpt-4', 'gpt-3.5-turbo'];
		if (openAIModels.some(search => model.startsWith(search)) && this.providers.openai) {
			return this.providers.openai;
		}
		const anthropicModels = ['claude'];
		if (anthropicModels.some(search => model.startsWith(search)) && this.providers.anthropic) {
			return this.providers.anthropic;
		}
		const googleglModels = ['bison', 'gecko'];
		if (googleglModels.some(search => model.includes(search)) && this.providers.googlegl) {
			return this.providers.googlegl;
		}
		const huggingfaceModels = ['microsoft/', 'replit/', 'Salesforce/', 'bigcode/', 'deepseek-ai/'];
		if (huggingfaceModels.some(search => model.startsWith(search)) && this.providers.huggingface) {
			return this.providers.huggingface;
		}
		// No provider was given and input model didnt match any known models, but has slash in it, so assume its a huggingface model
		if (model.includes('/')) {
			return this.providers.huggingface;
		}

		if (this.defaultProvider && this.defaultProvider !== '') {
			return this.providers[this.defaultProvider];
		}
		throw new Error('No default provider and No provider found for model ' + model);
	}
}

export function getAIProviderInfo(providerName: string): ProviderInfo | undefined {
	// Ensure the provided name is of type ProviderName
	if (Object.values(ProviderName).includes(providerName as ProviderName)) {
		return ALL_AI_PROVIDERS[providerName as ProviderName];
	}
	return undefined;
}
