import { AxiosRequestConfig } from 'axios';
import { log } from 'logger';
import { AIAPI } from './completion_provider';
import {
	APIFetchResponse,
	ChatCompletionRequestMessage,
	ChatCompletionRoleEnum,
	CompletionRequest,
	CompletionResponse,
	huggingfaceProviderInfo,
} from './spec';

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
		const request = await this.createRequestData(input);
		const requestConfig: AxiosRequestConfig = {
			url: `${this.providerInfo.chatCompletionPathPrefix}/${input.model}`,
			method: 'POST',
			data: request,
		};
		return this.handleDirectResponse(requestConfig);
	}

	private async createRequestData(input: CompletionRequest): Promise<Record<string, any>> {
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
		return request;
	}

	private async handleDirectResponse(requestConfig: AxiosRequestConfig): Promise<CompletionResponse | undefined> {
		try {
			const data = await this.apicaller.request(requestConfig);
			return this.parseFullResponse(data);
		} catch (error) {
			log.error('Error in completion request: ' + error);
			throw error;
		}
	}

	private parseFullResponse<T>(apiFetchData: APIFetchResponse<T>): CompletionResponse {
		if (!apiFetchData) {
			log.error('No API fetch data');
			return {};
		}
		const completionResponse: CompletionResponse = {
			requestDetails: apiFetchData.requestDetails,
			responseDetails: apiFetchData.responseDetails,
			errorDetails: apiFetchData.errorDetails,
		};

		const data = apiFetchData.data;
		if (typeof data !== 'object' || data === null) {
			const msg = 'Invalid data response. Expected an object.';
			// log.error(msg);
			if (completionResponse.errorDetails) {
				completionResponse.errorDetails.message += msg;
			} else {
				completionResponse.errorDetails = { message: msg };
			}
			return completionResponse;
		}

		let respText = '';
		if ('generated_text' in data) {
			respText = data.generated_text as string;
		} else if (Array.isArray(data) && data.length > 0) {
			// Get 'generated_text' from the first element of the array, if the array is not empty
			respText = data[0].generated_text as string;
		}

		completionResponse.respContent = respText;
		return completionResponse;
	}
}
