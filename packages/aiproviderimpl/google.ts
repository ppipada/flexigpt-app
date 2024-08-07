import {
	ChatCompletionRequestMessage,
	ChatCompletionRoleEnum,
	CompletionRequest,
	CompletionResponse,
	googleProviderInfo,
} from 'aiprovidermodel';
import { AxiosRequestConfig } from 'axios';
import { AIAPI } from './completion_provider';

interface Content {
	role: string;
	parts: { text: string }[];
}

export class GoogleAPI extends AIAPI {
	constructor() {
		super(googleProviderInfo);
	}

	convertMessages(messages: ChatCompletionRequestMessage[]): Content[] {
		return messages.map(message => {
			let role: string = 'user';
			switch (message.role) {
				case ChatCompletionRoleEnum.user:
					role = 'user';
					break;
				case ChatCompletionRoleEnum.assistant:
					role = 'model';
					break;
			}
			return {
				role,
				parts: [
					{
						text: message.content || '', // If content is undefined, use an empty string
					},
				],
			};
		});
	}

	async completion(input: CompletionRequest): Promise<CompletionResponse | undefined> {
		// return tempCodeString;
		// let messages: ChatCompletionRequestMessage[] = [{"role": "user", "content": "Hello!"}];
		if (!input.messages) {
			throw Error('No input messages found');
		}

		const generateConfig: Record<string, any> = {
			maxOutputTokens: input.maxOutputLength,
			temperature: input.temperature,
		};

		if (input.additionalParameters) {
			for (const key in input.additionalParameters) {
				if (!Object.prototype.hasOwnProperty.call(generateConfig, key)) {
					generateConfig[key] = input.additionalParameters[key];
				}
			}
		}

		const content = this.convertMessages(input.messages);
		// eslint-disable-next-line @typescript-eslint/naming-convention
		const request: Record<string, any> = {
			contents: content,
			generationConfig: generateConfig,
		};

		const modelpath = `${this.providerInfo.chatCompletionPathPrefix}/${input.model}:generateContent?key=${this.providerInfo.apiKey}`;
		const requestConfig: AxiosRequestConfig = {
			url: modelpath,
			method: 'POST',
			data: request,
		};

		const data = await this.apicaller.request(requestConfig);
		if (typeof data !== 'object' || data === null) {
			throw new Error('Invalid data response. Expected an object.' + data);
		}
		let respText = '';
		if ('candidates' in data && Array.isArray(data.candidates) && data.candidates.length > 0) {
			if (
				'content' in data.candidates[0] &&
				'parts' in data.candidates[0].content &&
				Array.isArray(data.candidates[0].content.parts) &&
				data.candidates[0].content.parts.length > 0 &&
				'text' in data.candidates[0].content.parts[0]
			) {
				respText = data.candidates[0].content.parts[0].text as string;
			}
		}
		const completionResponse: CompletionResponse = { fullResponse: data, respContent: respText };
		return completionResponse;
	}
}
