import { log } from 'logger';
import { ChatCompletionRequestMessage, ChatCompletionRoleEnum, CompletionRequest } from './chat_types';

export function unescapeChars(text: string) {
	return text.replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&amp;/g, '&');
}

function countTokensInContent(content: string): number {
	// Regular expression to split the content into tokens based on common delimiters.
	// This includes whitespaces, brackets, arithmetic operators, and punctuation.
	// eslint-disable-next-line no-useless-escape
	const tokenRegex = /[\s{}\[\]()+-=*/<>,;:.!&|\\]+/;

	// Split the content into tokens based on the regex and filter out empty strings.
	const tokens = content.split(tokenRegex).filter(token => token !== '');

	// Return the count of tokens.
	return tokens.length;
}

export function filterMessagesByTokenCount(
	messages: ChatCompletionRequestMessage[],
	maxTokenCount: number
): ChatCompletionRequestMessage[] {
	let totalTokens = 0;
	const filteredMessages: ChatCompletionRequestMessage[] = [];

	// Loop through the messages in reverse order (prioritizing the last element)
	for (let i = messages.length - 1; i >= 0; i--) {
		const message = messages[i];
		const c = message.content || '';
		const tokensInMessage = countTokensInContent(c);

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
		log.info(
			`Filtered messages count (${filteredMessages.length}) is less than input messages count (${messages.length})`
		);
	}
	return filteredMessages.reverse();
}

export function getCompletionRequest(
	defaultModel: string,
	prompt: string | null,
	messages: Array<ChatCompletionRequestMessage> | null,
	inputParams?: { [key: string]: any },
	defaultTemprature: number = 0.1,
	defaultMaxTokens: number = 2048,
	defaultLimitContextLength: number = 2048
): CompletionRequest {
	if (!inputParams) {
		inputParams = {};
	}
	const completionRequest: CompletionRequest = {
		model: defaultModel,
		prompt: prompt,
		messages: messages,
		temperature: defaultTemprature,
		maxTokens: defaultMaxTokens,
		limitContextLength: defaultLimitContextLength,
		stream: false,
	};

	for (const key in inputParams) {
		if (key === 'model' && typeof key === 'string') {
			completionRequest.model = inputParams.model;
			continue;
		}
		if (key === 'maxTokens' && typeof key === 'number') {
			completionRequest.maxTokens = inputParams.maxTokens;
			continue;
		}
		if (key === 'temperature' && typeof key === 'number') {
			completionRequest.temperature = inputParams.temperature;
			continue;
		}
		if (key === 'limitContextLength' && typeof key === 'number') {
			completionRequest.limitContextLength = inputParams.limitContextLength;
			continue;
		}
		if (key === 'systemPrompt' && typeof key === 'string') {
			completionRequest.systemPrompt = inputParams.systemPrompt;
			continue;
		}

		completionRequest.additionalParameters = completionRequest.additionalParameters || {};
		// eslint-disable-next-line no-prototype-builtins
		if (!completionRequest.hasOwnProperty(key) && key !== 'provider') {
			completionRequest.additionalParameters[key] = inputParams[key];
		}
	}

	if (completionRequest.prompt) {
		const message: ChatCompletionRequestMessage = {
			role: ChatCompletionRoleEnum.user,
			content: completionRequest.prompt,
		};
		if (!completionRequest.messages) {
			completionRequest.messages = [message];
		} else {
			completionRequest.messages.push(message);
		}
		completionRequest.prompt = null;
	}
	if (completionRequest.messages) {
		completionRequest.messages = filterMessagesByTokenCount(
			completionRequest.messages,
			completionRequest.limitContextLength || 2048
		);
	}
	return completionRequest;
}
