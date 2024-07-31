export enum ChatCompletionRoleEnum {
	system = 'system',
	user = 'user',
	assistant = 'assistant',
	function = 'function',
}

export interface ChatAPIMessage {
	id: string;
	createdAt?: Date;
	role: ChatCompletionRoleEnum;
	content: string;
	timestamp?: string;
	name?: string;
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
	messages?: Array<ChatCompletionRequestMessage> | null;
	prompt?: string | null;
	systemPrompt?: string | null;
	limitContextLength?: number | null;
	functions?: Array<ChatCompletionFunctions>;
	functionCall?: CreateChatCompletionRequestFunctionCall;
	suffix?: string | null;
	maxTokens?: number | null;
	temperature?: number | null;
	stream?: boolean | null;
	timeout?: number | null;
	/**
	 *  Map of additional parameters specific to the model.
	 *  @type {Record<string, any>}
	 *  Anything with non null/undefined value will be added to the request body
	 */
	additionalParameters?: Record<string, any> | null;
}

export interface CompletionProvider {
	completion(input: CompletionRequest): Promise<{ fullResponse: any; data: string | null }>;
	checkAndPopulateCompletionParams(
		prompt: string | null,
		messages: Array<ChatCompletionRequestMessage> | null,
		inputParams?: { [key: string]: any }
	): CompletionRequest;
	setAttribute(key: string, value: any): void;
}
