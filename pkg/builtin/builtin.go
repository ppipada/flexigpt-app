package builtin

import (
	"embed"
)

//go:embed promptbundles
var BuiltInPromptBundlesFS embed.FS

const (
	BuiltInPromptBundlesRootDir = "promptbundles"
	BuiltInPromptBundlesJSON    = "bundles.json"
)
