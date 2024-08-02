export enum ChatCompletionRoleEnum {
	system = 'system',
	user = 'user',
	assistant = 'assistant',
	function = 'function',
}

export interface ChatCompletionFunctions {
	name: string;
	description?: string;
	parameters?: { [key: string]: any };
}

export interface ChatCompletionRequestMessageFunctionCall {
	name?: string;
	arguments?: string;
}

export interface ChatCompletionRequestMessage {
	role: ChatCompletionRoleEnum;
	content?: string;
	name?: string;
	functionCall?: ChatCompletionRequestMessageFunctionCall;
}

export interface ChatCompletionResponseMessage {
	role: ChatCompletionRoleEnum;
	content?: string;
	functionCall?: ChatCompletionRequestMessageFunctionCall;
}

export type CreateChatCompletionRequestFunctionCall = CreateChatCompletionRequestFunctionCallOneOf | string;

export interface CreateChatCompletionRequestFunctionCallOneOf {
	name: string;
}

export interface CompletionRequest {
	model: string;
	messages?: Array<ChatCompletionRequestMessage>;
	temperature: number;
	maxPromptLength: number;
	stream: boolean;
	systemPrompt?: string;
	maxOutputLength?: number;
	functions?: Array<ChatCompletionFunctions>;
	functionCall?: CreateChatCompletionRequestFunctionCall;
	suffix?: string;
	timeout?: number;
	/**
	 *  Map of additional parameters specific to the model.
	 *  @type {Record<string, any>}
	 *  Anything with non null/undefined value will be added to the request body
	 */
	additionalParameters?: Record<string, any>;
}
