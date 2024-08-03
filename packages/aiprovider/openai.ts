import { AxiosRequestConfig } from 'axios';
import { log } from 'logger';
import {
	ChatCompletionRequestMessage,
	ChatCompletionRoleEnum,
	CompletionRequest,
	CompletionResponse,
} from './chat_types';
import { AIAPI } from './completion_provider';
import { openaiProviderInfo } from './provider_consts';
import { unescapeChars } from './provider_utils';

export class OpenAIAPI extends AIAPI {
	constructor() {
		super(openaiProviderInfo);
	}

	async completion(input: CompletionRequest): Promise<CompletionResponse | undefined> {
		// return tempCodeString;
		// let messages: ChatCompletionRequestMessage[] = [{"role": "user", "content": "Hello!"}];
		if (!input.messages) {
			throw Error('No input messages found');
		}

		// eslint-disable-next-line @typescript-eslint/naming-convention
		const request: Record<string, any> = {
			model: input.model,
			// eslint-disable-next-line @typescript-eslint/naming-convention
			max_tokens: input.maxOutputLength,
			temperature: input.temperature,
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
			const completionResponse: CompletionResponse = {
				fullResponse: data,
				respContent: respText,
				functionName: functionName,
				functionArgs: functionArgs,
			};
			return completionResponse;
		} catch (error) {
			log.error('Error in completion request: ' + error);
			throw error;
		}
	}
}
