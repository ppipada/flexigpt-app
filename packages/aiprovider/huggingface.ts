import { AxiosRequestConfig } from 'axios';
import { APICaller } from './api_fetch';
import { ChatCompletionRequestMessage, ChatCompletionRoleEnum, CompletionRequest } from './chat_types';
import { filterMessagesByTokenCount, getCompletionRequest } from './chatapibase/chat_utils';
import { huggingfaceProviderInfo } from './provider_consts';
import { CompletionProvider, ProviderInfo } from './provider_types';

export class HuggingFaceAPI implements CompletionProvider {
	private providerInfo: ProviderInfo;
	private apicaller: APICaller;
	constructor() {
		this.apicaller = new APICaller(
			huggingfaceProviderInfo.defaultOrigin,
			huggingfaceProviderInfo.apiKey,
			huggingfaceProviderInfo.apiKeyHeaderKey,
			huggingfaceProviderInfo.timeout,
			huggingfaceProviderInfo.defaultHeaders
		);
		this.providerInfo = huggingfaceProviderInfo;
	}

	getProviderInfo(): ProviderInfo {
		return this.providerInfo;
	}

	async getModelType(model: string) {
		const requestConfig: AxiosRequestConfig = {
			url: `${this.providerInfo.chatCompletionPathPrefix}/${model}`,
			method: 'GET',
		};
		const data = await this.apicaller.request(requestConfig);
		if (typeof data !== 'object' || data === null) {
			throw new Error('Invalid data response. Expected an object.');
		}
		if ('tags' in data) {
			const tags = data.tags as string[];
			if ('conversational' in tags) {
				return 'chat';
			}
		}
		return 'completion';
	}

	// eslint-disable-next-line @typescript-eslint/naming-convention
	getInputs(messages: ChatCompletionRequestMessage[]): {
		text: string;
		// eslint-disable-next-line @typescript-eslint/naming-convention
		generated_responses: string[];
		// eslint-disable-next-line @typescript-eslint/naming-convention
		past_user_inputs: string[];
	} {
		// eslint-disable-next-line @typescript-eslint/naming-convention
		const generated_responses: string[] = [];
		// eslint-disable-next-line @typescript-eslint/naming-convention
		const past_user_inputs: string[] = [];
		let text = '';

		for (let i = 0; i < messages.length; i++) {
			const icontent: string = messages[i].content || '';
			if (messages[i].role === ChatCompletionRoleEnum.assistant) {
				generated_responses.push(icontent);
			} else if (
				messages[i].role === ChatCompletionRoleEnum.user ||
				messages[i].role === ChatCompletionRoleEnum.system
			) {
				past_user_inputs.push(icontent);

				// Assuming the last input from the user is at the end of the array
				if (i === messages.length - 1) {
					text = icontent;
				}
			}
		}
		// eslint-disable-next-line @typescript-eslint/naming-convention
		return { text, generated_responses, past_user_inputs };
	}

	async completion(input: CompletionRequest): Promise<any> {
		if (!input.messages) {
			throw Error('No input messages found');
		}
		const model = input.model;
		const modeltype = await this.getModelType(model);

		const parameters: Record<string, any> = {
			// eslint-disable-next-line @typescript-eslint/naming-convention
			max_length: input.maxTokens ? input.maxTokens : 4096,
			temperature: input.temperature ? input.temperature : 0.1,
			// eslint-disable-next-line @typescript-eslint/naming-convention
			max_time: input.timeout ? input.timeout : this.providerInfo.timeout,
		};

		if (input.additionalParameters) {
			for (const key in input.additionalParameters) {
				if (!Object.prototype.hasOwnProperty.call(parameters, key)) {
					parameters[key] = input.additionalParameters[key];
				}
			}
		}

		if (modeltype !== 'chat') {
			parameters.return_full_text = false;
			if (!input.maxTokens || input.maxTokens > 250) {
				parameters.max_length = 250;
			}
		}
		let filterTokens = 250;
		if (parameters.max_length) {
			filterTokens = parameters.max_length;
		}
		const messages = filterMessagesByTokenCount(input.messages, filterTokens);

		const inputmessages = this.getInputs(messages);

		const request: Record<string, any> = {
			parameters: parameters,
		};

		if (modeltype === 'chat') {
			request.inputs = inputmessages;
		} else {
			request.inputs = inputmessages.text;
		}

		const requestConfig: AxiosRequestConfig = {
			url: `${this.providerInfo.chatCompletionPathPrefix}/${input.model}`,
			method: 'POST',
			data: request,
		};
		const data = await this.apicaller.request(requestConfig);
		const fullResponse = data;
		if (typeof data !== 'object' || data === null) {
			throw new Error('Invalid data response. Expected an object.' + data);
		}
		let respText = '';
		if ('generated_text' in data) {
			respText = data.generated_text as string;
		} else if (Array.isArray(data) && data.length > 0) {
			// Get 'generated_text' from the first element of the array, if the array is not empty
			respText = data[0].generated_text as string;
		}
		return { fullResponse: fullResponse, data: respText };
	}

	public checkAndPopulateCompletionParams(
		prompt: string | null,
		messages: Array<ChatCompletionRequestMessage> | null,
		inputParams?: { [key: string]: any }
	): CompletionRequest {
		return getCompletionRequest(this.providerInfo.defaultModel, prompt, messages, inputParams);
	}
}
