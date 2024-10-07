package spec

// CodeProps represents the properties of a code block
type CodeProps struct {
	Language string
	Value    string
}

// ProgrammingLanguages is a map of programming languages to their file extensions
var ProgrammingLanguages = map[string]string{
	"javascript":  ".js",
	"python":      ".py",
	"java":        ".java",
	"c":           ".c",
	"cpp":         ".cpp",
	"c++":         ".cpp",
	"c#":          ".cs",
	"ruby":        ".rb",
	"php":         ".php",
	"swift":       ".swift",
	"objective-c": ".m",
	"kotlin":      ".kt",
	"typescript":  ".ts",
	"go":          ".go",
	"perl":        ".pl",
	"rust":        ".rs",
	"scala":       ".scala",
	"haskell":     ".hs",
	"lua":         ".lua",
	"shell":       ".sh",
	"sql":         ".sql",
	"html":        ".html",
	"css":         ".css",
	"json":        ".json",
	"dart":        ".dart",
	// add more file extensions here, make sure the key is same as language prop in CodeBlock.tsx component
}
