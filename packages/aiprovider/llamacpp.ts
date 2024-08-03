import { AxiosRequestConfig } from 'axios';
import { log } from 'logger';
import {
	ChatCompletionRequestMessage,
	ChatCompletionRoleEnum,
	CompletionRequest,
	CompletionResponse,
} from './chat_types';
import { AIAPI } from './completion_provider';
import { llamacppProviderInfo } from './provider_consts';

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

		const requestConfig: AxiosRequestConfig = {
			url: this.providerInfo.chatCompletionPathPrefix,
			method: 'POST',
			data: request,
		};
		try {
			const data = await this.apicaller.request(requestConfig);
			if (typeof data !== 'object' || data === null) {
				throw new Error('Invalid data response. Expected an object.' + data);
			}
			const respText = 'content' in data ? (data?.content as string) : '';
			const completionResponse: CompletionResponse = { fullResponse: data, respContent: respText };
			return completionResponse;
		} catch (error) {
			log.error('Error in completion request: ' + error);
			throw error;
		}
	}
}
