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
		log.debug(
			`Filtered messages count (${filteredMessages.length}) is less than input messages count (${messages.length})`
		);
	}
	return filteredMessages.reverse();
}

export function getCompletionRequest(
	defaultModel: string,
	prompt: string,
	defaultTemperature: number,
	maxPromptLength: number,
	prevMessages?: Array<ChatCompletionRequestMessage>,
	inputParams?: { [key: string]: any }
): CompletionRequest {
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
		stream: false,
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
			if (key !== 'provider') {
				completionRequest.additionalParameters[key] = inputParams[key];
			}
		}
	}

	if (completionRequest.messages) {
		completionRequest.messages = filterMessagesByTokenCount(
			completionRequest.messages,
			completionRequest.maxPromptLength
		);
	}

	return completionRequest;
}
