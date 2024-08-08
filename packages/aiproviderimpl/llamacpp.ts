import {
	APIFetchResponse,
	ChatCompletionRequestMessage,
	ChatCompletionRoleEnum,
	CompletionRequest,
	CompletionResponse,
	llamacppProviderInfo,
} from 'aiprovidermodel';
import { AxiosRequestConfig } from 'axios';
import { log } from 'logger';
import { AIAPI } from './completion_provider';

export class LlamaCPPAPI extends AIAPI {
	constructor() {
		super(llamacppProviderInfo);
	}

	convertChat(
		messages: ChatCompletionRequestMessage[],
		userName = '\\nUSER: ',
		aiName = '\\nASSISTANT: ',
		stop = '</s>'
	): string {
		let prompt = '';

		const userN = userName.replace('\\n', '\n');
		const aiN = aiName.replace('\\n', '\n');
		const stopSymbol = stop.replace('\\n', '\n');

		for (const line of messages) {
			if (line.role === ChatCompletionRoleEnum.user) {
				prompt += `${userN}${line.content}`;
			}
			if (line.role === ChatCompletionRoleEnum.assistant) {
				prompt += `${aiN}${line.content}${stopSymbol}`;
			}
		}
		prompt += aiN.trimEnd();

		return prompt;
	}

	async completion(input: CompletionRequest): Promise<CompletionResponse | undefined> {
		const request = this.createRequestData(input);
		const requestConfig: AxiosRequestConfig = {
			url: this.providerInfo.chatCompletionPathPrefix,
			method: 'POST',
			data: request,
		};
		return this.handleDirectResponse(requestConfig);
	}

	private createRequestData(input: CompletionRequest): Record<string, any> {
		// return tempCodeString;
		// let messages: ChatCompletionRequestMessage[] = [{"role": "user", "content": "Hello!"}];
		if (!input.messages) {
			throw Error('No input messages found');
		}

		// eslint-disable-next-line @typescript-eslint/naming-convention
		const request: Record<string, any> = {
			prompt: this.convertChat(input.messages),
			// eslint-disable-next-line @typescript-eslint/naming-convention
			n_predict: input.maxOutputLength,
			temperature: input.temperature,
			stream: false,
		};

		if (input.additionalParameters) {
			for (const key in input.additionalParameters) {
				if (key === 'systemPrompt' && typeof key === 'string') {
					request['system_prompt'] = input.additionalParameters[key];
					continue;
				}
				// eslint-disable-next-line no-prototype-builtins
				if (!request.hasOwnProperty(key)) {
					request[key] = input.additionalParameters[key];
				}
			}
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
		const respText = 'content' in data ? (data?.content as string) : '';
		completionResponse.respContent = respText;
		return completionResponse;
	}
}
