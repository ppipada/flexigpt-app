import { AxiosRequestConfig } from 'axios';
import { log } from 'logger';
import { APICaller } from './api_fetch';
import { ChatCompletionRequestMessage, ChatCompletionRoleEnum, CompletionRequest } from './chat_types';
import { getCompletionRequest } from './chatapibase/chat_utils';
import { llamacppProviderInfo } from './provider_consts';
import { CompletionProvider, ProviderInfo } from './provider_types';

export class LlamaCPPAPI implements CompletionProvider {
	private providerInfo: ProviderInfo;
	private apicaller: APICaller;
	constructor() {
		this.apicaller = new APICaller(
			llamacppProviderInfo.defaultOrigin,
			llamacppProviderInfo.apiKey,
			llamacppProviderInfo.apiKeyHeaderKey,
			llamacppProviderInfo.timeout,
			llamacppProviderInfo.defaultHeaders
		);
		this.providerInfo = llamacppProviderInfo;
	}

	getProviderInfo(): ProviderInfo {
		return this.providerInfo;
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

	async completion(input: CompletionRequest): Promise<any> {
		// return tempCodeString;
		// let messages: ChatCompletionRequestMessage[] = [{"role": "user", "content": "Hello!"}];
		if (!input.messages) {
			throw Error('No input messages found');
		}

		// eslint-disable-next-line @typescript-eslint/naming-convention
		const request: Record<string, any> = {
			prompt: this.convertChat(input.messages),
			// eslint-disable-next-line @typescript-eslint/naming-convention
			n_predict: input.maxTokens ? input.maxTokens : 1024,
			temperature: input.temperature ? input.temperature : 0.1,
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
			const fullResponse = data;
			if (typeof data !== 'object' || data === null) {
				throw new Error('Invalid data response. Expected an object.' + data);
			}
			const respText = 'content' in data ? (data?.content as string) : '';
			return {
				fullResponse: fullResponse,
				data: respText,
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
