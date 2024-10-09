import { log } from 'logger';
import { APICaller } from './api_fetch';
import {
	ALL_MODEL_INFO,
	ChatCompletionRequestMessage,
	ChatCompletionRoleEnum,
	CompletionRequest,
	CompletionResponse,
	ModelName,
	ProviderInfo,
} from './spec';

export interface CompletionProvider {
	getProviderInfo(): ProviderInfo;
	getCompletionRequest(
		prompt: string,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		inputParams?: { [key: string]: any },
		stream?: boolean
	): CompletionRequest;
	completion(
		input: CompletionRequest,
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined>;
	setAttribute(apiKey?: string, defaultModel?: ModelName, defaultTemperature?: number, defaultOrigin?: string): void;
	isConfigured(): boolean;
}

export class AIAPI implements CompletionProvider {
	protected providerInfo: ProviderInfo;
	protected apicaller: APICaller;

	constructor(providerInfo: ProviderInfo) {
		this.apicaller = new APICaller(
			providerInfo.defaultOrigin,
			providerInfo.apiKey,
			providerInfo.apiKeyHeaderKey,
			providerInfo.timeout,
			providerInfo.defaultHeaders
		);
		this.providerInfo = providerInfo;
	}

	getProviderInfo(): ProviderInfo {
		return this.providerInfo;
	}

	isConfigured(): boolean {
		if (this.apicaller.apiKey && this.apicaller.apiKey !== '') {
			return true;
		}
		return false;
	}

	setAttribute(apiKey?: string, defaultModel?: ModelName, defaultTemperature?: number, defaultOrigin?: string) {
		if (apiKey) {
			this.apicaller.apiKey = apiKey;
			this.providerInfo.apiKey = apiKey;
		}
		if (defaultOrigin) {
			this.apicaller.origin = defaultOrigin;
			this.providerInfo.defaultOrigin = defaultOrigin;
		}
		if (defaultModel) {
			this.providerInfo.defaultModel = defaultModel;
		}
		if (defaultTemperature) {
			this.providerInfo.defaultTemperature = defaultTemperature;
		}
	}
	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	async completion(
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		input: CompletionRequest,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined> {
		return undefined;
	}

	protected countTokensInContent(content: string): number {
		// Regular expression to split the content into tokens based on common delimiters.
		// This includes whitespaces, brackets, arithmetic operators, and punctuation.
		// eslint-disable-next-line no-useless-escape
		const tokenRegex = /[\s{}\[\]()+-=*/<>,;:.!&|\\]+/;

		// Split the content into tokens based on the regex and filter out empty strings.
		const tokens = content.split(tokenRegex).filter(token => token !== '');

		// Return the count of tokens.
		return tokens.length;
	}

	protected filterMessagesByTokenCount(
		messages: ChatCompletionRequestMessage[],
		maxTokenCount: number
	): ChatCompletionRequestMessage[] {
		let totalTokens = 0;
		const filteredMessages: ChatCompletionRequestMessage[] = [];

		// Loop through the messages in reverse order (prioritizing the last element)
		for (let i = messages.length - 1; i >= 0; i--) {
			const message = messages[i];
			const c = message.content || '';
			const tokensInMessage = this.countTokensInContent(c);

			// Check if adding this message will not exceed maxTokenCount
			// or if the filteredMessages array is empty, then at least add this message
			if (totalTokens + tokensInMessage <= maxTokenCount || filteredMessages.length === 0) {
				filteredMessages.push(message);
				totalTokens += tokensInMessage;

				// Always include at least one message, so if we've added one we can now enforce maxTokenCount
				if (totalTokens > maxTokenCount) {
					break;
				}
			} else {
				break;
			}
		}
		if (filteredMessages.length < messages.length) {
			log.debug(
				`Filtered messages count (${filteredMessages.length}) is less than input messages count (${messages.length})`
			);
		}
		return filteredMessages.reverse();
	}

	getCompletionRequest(
		prompt: string,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		inputParams?: { [key: string]: any },
		stream?: boolean
	): CompletionRequest {
		const defaultModel = this.providerInfo.defaultModel;
		const defaultTemperature = this.providerInfo.defaultTemperature;
		const maxPromptLength = ALL_MODEL_INFO[this.providerInfo.defaultModel].maxPromptLength;

		if (!inputParams) {
			inputParams = {};
		}

		const messages: Array<ChatCompletionRequestMessage> = prevMessages ? [...prevMessages] : [];
		if (prompt) {
			const message: ChatCompletionRequestMessage = {
				role: ChatCompletionRoleEnum.user,
				content: prompt,
			};
			messages.push(message);
		}

		const completionRequest: CompletionRequest = {
			model: defaultModel,
			messages: messages,
			temperature: defaultTemperature,
			stream: stream || false,
			maxPromptLength: maxPromptLength,
		};

		for (const key in inputParams) {
			if (!Object.prototype.hasOwnProperty.call(inputParams, key)) {
				continue;
			}
			if (key === 'model' && typeof inputParams[key] === 'string') {
				completionRequest.model = inputParams[key];
			} else if (key === 'maxOutputLength' && typeof inputParams[key] === 'number') {
				completionRequest.maxOutputLength = inputParams[key];
			} else if (key === 'temperature' && typeof inputParams[key] === 'number') {
				completionRequest.temperature = inputParams[key];
			} else if (key === 'maxPromptLength' && typeof inputParams[key] === 'number') {
				completionRequest.maxPromptLength = inputParams[key];
			} else if (key === 'systemPrompt' && typeof inputParams[key] === 'string') {
				completionRequest.systemPrompt = inputParams[key];
			} else {
				completionRequest.additionalParameters = completionRequest.additionalParameters || {};
				// No additional param support as of now
				if (key !== 'provider') {
					completionRequest.additionalParameters[key] = inputParams[key];
				}
			}
		}

		if (completionRequest.messages) {
			completionRequest.messages = this.filterMessagesByTokenCount(
				completionRequest.messages,
				completionRequest.maxPromptLength
			);
		}

		return completionRequest;
	}
}
