import { AxiosRequestConfig } from 'axios';
import {
	ChatCompletionRequestMessage,
	ChatCompletionRoleEnum,
	CompletionRequest,
	CompletionResponse,
} from './chat_types';
import { AIAPI } from './completion_provider';
import { huggingfaceProviderInfo } from './provider_consts';

export class HuggingFaceAPI extends AIAPI {
	constructor() {
		super(huggingfaceProviderInfo);
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

	async completion(input: CompletionRequest): Promise<CompletionResponse | undefined> {
		if (!input.messages) {
			throw Error('No input messages found');
		}
		const model = input.model;
		const modeltype = await this.getModelType(model);

		const parameters: Record<string, any> = {
			// eslint-disable-next-line @typescript-eslint/naming-convention
			max_length: input.maxOutputLength,
			temperature: input.temperature,
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
			if (!input.maxOutputLength || input.maxOutputLength > 250) {
				parameters.max_length = 250;
			}
		}
		let filterTokens = 250;
		if (parameters.max_length) {
			filterTokens = parameters.max_length;
		}
		const messages = this.filterMessagesByTokenCount(input.messages, filterTokens);

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
		const completionResponse: CompletionResponse = { fullResponse: data, respContent: respText };
		return completionResponse;
	}
}
