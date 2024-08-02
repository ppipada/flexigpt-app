import { AxiosRequestConfig } from 'axios';
import { log } from 'logger';
import { APICaller } from './api_fetch';
import { ChatCompletionRequestMessage, ChatCompletionRoleEnum, CompletionRequest } from './chat_types';
import { openaiProviderInfo } from './provider_consts';
import { CompletionProvider, ProviderInfo } from './provider_types';
import { getCompletionRequest, unescapeChars } from './provider_utils';

export class OpenAIAPI implements CompletionProvider {
	private providerInfo: ProviderInfo;
	private apicaller: APICaller;
	constructor() {
		this.apicaller = new APICaller(
			openaiProviderInfo.defaultOrigin,
			openaiProviderInfo.apiKey,
			openaiProviderInfo.apiKeyHeaderKey,
			openaiProviderInfo.timeout,
			openaiProviderInfo.defaultHeaders
		);
		this.providerInfo = openaiProviderInfo;
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

		// eslint-disable-next-line @typescript-eslint/naming-convention
		const request: Record<string, any> = {
			model: input.model,
			// eslint-disable-next-line @typescript-eslint/naming-convention
			max_tokens: input.maxTokens ? input.maxTokens : 4096,
			temperature: input.temperature ? input.temperature : 0.1,
			stream: false,
		};

		if (input.additionalParameters) {
			for (const key in input.additionalParameters) {
				if (key !== 'systemPrompt' && !Object.prototype.hasOwnProperty.call(request, key)) {
					request[key] = input.additionalParameters[key];
				}
			}
		}

		const modelpath = this.providerInfo.chatCompletionPathPrefix;
		request.messages = input.messages;
		if (input.additionalParameters) {
			for (const key in input.additionalParameters) {
				// eslint-disable-next-line no-prototype-builtins
				if (key === 'systemPrompt' && typeof key === 'string') {
					request['system'] = input.additionalParameters[key];
					const message: ChatCompletionRequestMessage = {
						role: ChatCompletionRoleEnum.system,
						content: input.additionalParameters[key],
					};
					request.messages = input.messages.unshift(message);
					break;
				}
			}
		}

		const requestConfig: AxiosRequestConfig = {
			url: modelpath,
			method: 'POST',
			data: request,
		};
		try {
			const data = await this.apicaller.request(requestConfig);
			const fullResponse = data;
			if (typeof data !== 'object' || data === null) {
				throw new Error('Invalid data response. Expected an object.' + data);
			}
			let respText = '';
			let functionName = '';
			let functionArgs: any;
			if ('choices' in data && Array.isArray(data.choices) && data.choices.length > 0) {
				const responseMessage = data.choices[0].message;
				respText = responseMessage?.content ? (responseMessage?.content as string) : '';
				if (
					'tool_calls' in responseMessage &&
					responseMessage['tool_calls'].length > 0 &&
					'function' in responseMessage['tool_calls'][0]
				) {
					functionName = responseMessage['tool_calls'][0]['function']['name'];
					respText += '\nFunction call:\nName:' + functionName;
					try {
						functionArgs = JSON.parse(unescapeChars(responseMessage['tool_calls'][0]['function']['arguments']));
					} catch (error) {
						log.error(
							'Error parsing function call arguments: ' +
								error +
								' ' +
								responseMessage['tool_calls'][0]['function']['arguments']
						);
						respText += '\nError in parsing returned args\n';
						functionArgs = responseMessage['tool_calls'][0]['function']['arguments'];
					}
					respText += '\nArgs: ' + JSON.stringify(functionArgs, null, 2);
				}
			}
			return {
				fullResponse: fullResponse,
				data: respText,
				functionName: functionName,
				functionArgs: functionArgs,
			};
		} catch (error) {
			log.error('Error in completion request: ' + error);
			throw error;
		}
	}

	public checkAndPopulateCompletionParams(
		prompt: string | null,
		messages: Array<ChatCompletionRequestMessage> | null,
		inputParams?: { [key: string]: any }
	): CompletionRequest {
		return getCompletionRequest(this.providerInfo.defaultModel, prompt, messages, inputParams);
	}
}
