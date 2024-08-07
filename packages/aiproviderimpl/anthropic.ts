import {
	ALL_MODEL_INFO,
	anthropicProviderInfo,
	CompletionRequest,
	CompletionResponse,
	ModelName,
} from 'aiprovidermodel';
import { AxiosRequestConfig } from 'axios';
import { AIAPI } from './completion_provider';

export class AnthropicAPI extends AIAPI {
	constructor() {
		super(anthropicProviderInfo);
	}

	async completion(input: CompletionRequest): Promise<CompletionResponse | undefined> {
		// return tempCodeString;
		// let messages: ChatCompletionRequestMessage[] = [{"role": "user", "content": "Hello!"}];
		if (!input.messages) {
			throw Error('No input messages found');
		}

		let maxTokens = input.maxOutputLength;
		if (!maxTokens) {
			if (input.model in ALL_MODEL_INFO) {
				maxTokens = ALL_MODEL_INFO[input.model as ModelName].maxOutputLength;
			} else {
				maxTokens = 4096;
			}
		}
		const request: Record<string, any> = {
			model: input.model,
			messages: input.messages,
			// eslint-disable-next-line @typescript-eslint/naming-convention
			max_tokens: maxTokens,
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
		const completionResponse: CompletionResponse = { fullResponse: data, respContent: respText };
		return completionResponse;
	}
}
