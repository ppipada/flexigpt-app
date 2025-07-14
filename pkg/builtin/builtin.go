package builtin

import (
	"embed"
)

//go:embed promptbundles
var BuiltinPromptBundlesFS embed.FS

const (
	BuiltinPromptBundlesRootDir = "promptbundles"
	BuiltinPromptBundlesJSON    = "bundles.json"
)
