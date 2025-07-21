package builtin

import (
	"embed"
)

//go:embed prompts
var BuiltInPromptBundlesFS embed.FS

//go:embed tools
var BuiltInToolBundlesFS embed.FS

const (
	BuiltInPromptBundlesRootDir = "prompts"
	BuiltInPromptBundlesJSON    = "prompts.bundles.json"
	BuiltInToolBundlesRootDir   = "tools"
	BuiltInToolBundlesJSON      = "tools.bundles.json"
)
