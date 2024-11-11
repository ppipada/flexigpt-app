export enum ProviderName {
	OPENAI = 'openai',
	ANTHROPIC = 'anthropic',
	GOOGLE = 'google',
	HUGGINGFACE = 'huggingface',
	LLAMACPP = 'llamacpp',
}

export enum ModelName {
	CLAUDE_3_5_SONNET = 'claude-3-5-sonnet-20240620',
	CLAUDE_3_OPUS = 'claude-3-opus-20240229',
	CLAUDE_3_SONNET = 'claude-3-sonnet-20240229',
	CLAUDE_3_HAIKU = 'claude-3-haiku-20240307',

	GEMINI_1_5_FLASH = 'gemini-1.5-flash',
	GEMINI_1_5_PRO = 'gemini-1.5-pro',

	DEEPSEEK_CODER_1_3B_INSTRUCT = 'deepseek-ai/deepseek-coder-1.3b-instruct',

	LLAMA_3 = 'llama3',
	LLAMA_3_1 = 'llama3.1',

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
	defaultTemperature?: number;
	streamingSupport?: boolean;
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
	streamingSupport: boolean;
	modelPrefixes?: string[];
	descriptions?: Partial<Record<keyof ProviderInfo, string>>;
}
