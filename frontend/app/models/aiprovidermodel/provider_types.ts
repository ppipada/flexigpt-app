import { ChatCompletionRequestMessage, CompletionRequest, CompletionResponse } from './chat_types';

export enum ProviderName {
	OPENAI = 'openai',
	ANTHROPIC = 'anthropic',
	GOOGLE = 'google',
	HUGGINGFACE = 'huggingface',
	LLAMACPP = 'llamacpp',
}

export enum ModelName {
	CLAUDE_3_5_SONNET = 'claude-3-5-sonnet-20241022',
	CLAUDE_3_5_HAIKU = 'claude-3-5-haiku-20241022',
	CLAUDE_3_OPUS = 'claude-3-opus-20240229',
	CLAUDE_3_SONNET = 'claude-3-sonnet-20240229',
	CLAUDE_3_HAIKU = 'claude-3-haiku-20240307',

	GEMINI_2_FLASH_EXP = 'gemini-2.0-flash-exp',
	GEMINI_1_5_FLASH = 'gemini-1.5-flash',
	GEMINI_1_5_PRO = 'gemini-1.5-pro',

	DEEPSEEK_CODER_1_3B_INSTRUCT = 'deepseek-ai/deepseek-coder-1.3b-instruct',

	LLAMA_3 = 'llama3',
	LLAMA_3_1 = 'llama3.1',

	GPT_O3_MINI = 'o3-mini',
	GPT_O1 = 'o1',
	GPT_O1_PREVIEW = 'o1-preview',
	GPT_O1_MINI = 'o1-mini',
	GPT_4O_MINI = 'gpt-4o-mini',
	GPT_4O = 'gpt-4o',
	GPT_4 = 'gpt-4',
	GPT_3_5_TURBO = 'gpt-3.5-turbo',
}

export interface ModelInfo {
	name: ModelName;
	displayName: string;
	provider: ProviderName;
	maxPromptLength: number;
	maxOutputLength: number;
	defaultTemperature: number;
}

export interface ProviderInfo {
	name: ProviderName;
	apiKey: string;
	engine: string;
	defaultOrigin: string;
	defaultModel: ModelName;
	additionalSettings: Record<string, any>;
	timeout: number;
	apiKeyHeaderKey: string;
	defaultHeaders: Record<string, string>;
	chatCompletionPathPrefix: string;
	defaultTemperature: number;
	modelPrefixes?: string[];
	descriptions?: Partial<Record<keyof ProviderInfo, string>>;
	streamingSupport?: boolean;
}

export interface IProviderSetAPI {
	getDefaultProvider(): Promise<ProviderName>;
	setDefaultProvider(provider: ProviderName): Promise<void>;
	getConfigurationInfo(): Promise<Record<string, any>>;
	setAttribute(
		provider: ProviderName,
		apiKey?: string,
		defaultModel?: ModelName,
		defaultTemperature?: number,
		defaultOrigin?: string
	): Promise<void>;

	getCompletionRequest(
		provider: ProviderName,
		prompt: string,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		inputParams?: { [key: string]: any }
	): Promise<CompletionRequest>;
	completion(
		provider: ProviderName,
		input: CompletionRequest,
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined>;
}
