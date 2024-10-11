package spec

// ModelName is an enumeration of model names.
type ModelName string

const (
	CLAUDE_3_5_SONNET            ModelName = "claude-3-5-sonnet-20240620"
	CLAUDE_3_OPUS                ModelName = "claude-3-opus-20240229"
	CLAUDE_3_SONNET              ModelName = "claude-3-sonnet-20240229"
	CLAUDE_3_HAIKU               ModelName = "claude-3-haiku-20240307"
	GEMINI_1_5_FLASH             ModelName = "gemini-1.5-flash"
	GEMINI_1_5_PRO               ModelName = "gemini-1.5-pro"
	DEEPSEEK_CODER_1_3B_INSTRUCT ModelName = "deepseek-ai/deepseek-coder-1.3b-instruct"
	LLAMA_3                      ModelName = "llama3"
	LLAMA_3_1                    ModelName = "llama3.1"
	GPT_O1_PREVIEW               ModelName = "o1-preview"
	GPT_O1_MINI                  ModelName = "o1-mini"
	GPT_4O_MINI                  ModelName = "gpt-4o-mini"
	GPT_4O                       ModelName = "gpt-4o"
	GPT_4                        ModelName = "gpt-4"
	GPT_3_5_TURBO                ModelName = "gpt-3.5-turbo"
)

// ModelInfo represents information about a model.
type ModelInfo struct {
	Name               ModelName    `json:"name"`
	DisplayName        string       `json:"displayName"`
	Provider           ProviderName `json:"provider"`
	MaxPromptLength    int          `json:"maxPromptLength"`
	MaxOutputLength    int          `json:"maxOutputLength"`
	DefaultTemperature float64      `json:"defaultTemperature"`
}
