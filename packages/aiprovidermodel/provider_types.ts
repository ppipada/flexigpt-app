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

	GPT_4O_MINI = 'gpt-4o-mini',
	GPT_4O = 'gpt-4o',
	GPT_4 = 'gpt-4',
	GPT_3_5_TURBO = 'gpt-3.5-turbo',
}

export interface ModelInfo {
	name: ModelName;
	provider: ProviderName;
	maxPromptLength: number;
	maxOutputLength: number;
	defaultTemperature: number;
}

export interface ProviderInfo {
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

	// Method to get description of a key
	getDescription(key: string): string | undefined;
}

// Implementing the ProviderInfo interface in a class
export class ProviderInfoImpl implements ProviderInfo {
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
	streamingSupport?: boolean;
	descriptions?: Partial<Record<keyof ProviderInfo, string>>;

	constructor(providerInfo: Omit<ProviderInfo, 'getDescription'>) {
		this.apiKey = providerInfo.apiKey;
		this.engine = providerInfo.engine;
		this.defaultOrigin = providerInfo.defaultOrigin;
		this.defaultModel = providerInfo.defaultModel;
		this.additionalSettings = providerInfo.additionalSettings;
		this.timeout = providerInfo.timeout;
		this.apiKeyHeaderKey = providerInfo.apiKeyHeaderKey;
		this.defaultHeaders = providerInfo.defaultHeaders;
		this.chatCompletionPathPrefix = providerInfo.chatCompletionPathPrefix;
		this.defaultTemperature = providerInfo.defaultTemperature;
		this.modelPrefixes = providerInfo.modelPrefixes;
		this.streamingSupport = providerInfo.streamingSupport;
		this.descriptions = providerInfo.descriptions;
	}

	getDescription(key: string): string | undefined {
		if (!(key as keyof ProviderInfo)) {
			return undefined;
		}
		return this.descriptions ? this.descriptions[key as keyof ProviderInfo] : undefined;
	}
}
