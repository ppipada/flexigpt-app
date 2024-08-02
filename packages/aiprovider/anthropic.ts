import { AxiosRequestConfig } from 'axios';
import { APICaller } from './api_fetch';
import { ChatCompletionRequestMessage, CompletionRequest } from './chat_types';
import { ALL_MODEL_INFO, anthropicProviderInfo } from './provider_consts';
import { CompletionProvider, ProviderInfo } from './provider_types';
import { getCompletionRequest } from './provider_utils';

export class AnthropicAPI implements CompletionProvider {
	private providerInfo: ProviderInfo;
	private apicaller: APICaller;
	constructor() {
		this.apicaller = new APICaller(
			anthropicProviderInfo.defaultOrigin,
			anthropicProviderInfo.apiKey,
			anthropicProviderInfo.apiKeyHeaderKey,
			anthropicProviderInfo.timeout,
			anthropicProviderInfo.defaultHeaders
		);
		this.providerInfo = anthropicProviderInfo;
	}

	getProviderInfo(): ProviderInfo {
		return this.providerInfo;
	}

	async completion(input: CompletionRequest): Promise<any> {
		// return tempCodeString;
		// let messages: ChatCompletionRequestMessage[] = [{"role": "user", "content": "Hello!"}];
		if (!input.messages) {
			throw Error('No input messages found');
		}

		const request: Record<string, any> = {
			model: input.model,
			messages: input.messages,
			// eslint-disable-next-line @typescript-eslint/naming-convention
			max_tokens: input.maxOutputLength,
			temperature: input.temperature,
			stream: false,
		};

		if (input.additionalParameters) {
			for (const key in input.additionalParameters) {
				// eslint-disable-next-line no-prototype-builtins
				if (key === 'systemPrompt' && typeof key === 'string') {
					request['system'] = input.additionalParameters[key];
					continue;
				}
				if (!Object.prototype.hasOwnProperty.call(request, key)) {
					request[key] = input.additionalParameters[key];
				}
			}
		}

		// eslint-disable-next-line prefer-const
		let requestConfig: AxiosRequestConfig = {
			url: this.providerInfo.chatCompletionPathPrefix,
			method: 'POST',
			data: request,
		};
		const data = await this.apicaller.request(requestConfig);
		if (typeof data !== 'object' || data === null) {
			throw new Error('Invalid data response. Expected an object.');
		}
		let respText = '';
		if ('content' in data && Array.isArray(data.content) && data.content.length > 0) {
			for (const resp of data.content) {
				respText += resp.text + '\n';
			}
		}
		return { fullResponse: data, data: respText };
	}

	public checkAndPopulateCompletionParams(
		prompt: string,
		messages?: Array<ChatCompletionRequestMessage>,
		inputParams?: { [key: string]: any }
	): CompletionRequest {
		return getCompletionRequest(
			this.providerInfo.defaultModel,
			prompt,
			this.providerInfo.defaultTemperature,
			ALL_MODEL_INFO[this.providerInfo.defaultModel].maxPromptLength,
			messages,
			inputParams
		);
	}
}
