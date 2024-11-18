package spec

const ChromemDocStoreName = "chromemlocal"
const ChromemDocStorePath = "./chromemgo"

const HTTPDocStoreName = DocumentDBID("http")
const HTTPDocStoreEndpoint = "http://127.0.0.1:8080"

var HTTPDocStoreHeaders = map[string]string{}

const (
	EmbeddingModelOpenAI3Small              EmbeddingFuncID = "text-embedding-3-small"
	EmbeddingModelOpenAI3Large              EmbeddingFuncID = "text-embedding-3-large"
	EmbeddingModelCohereMultilingualV2      EmbeddingFuncID = "embed-multilingual-v2.0"
	EmbeddingModelCohereEnglishLightV2      EmbeddingFuncID = "embed-english-light-v2.0"
	EmbeddingModelCohereEnglishV2           EmbeddingFuncID = "embed-english-v2.0"
	EmbeddingModelCohereMultilingualLightV3 EmbeddingFuncID = "embed-multilingual-light-v3.0"
	EmbeddingModelCohereEnglishLightV3      EmbeddingFuncID = "embed-english-light-v3.0"
	EmbeddingModelCohereMultilingualV3      EmbeddingFuncID = "embed-multilingual-v3.0"
	EmbeddingModelCohereEnglishV3           EmbeddingFuncID = "embed-english-v3.0"
)
