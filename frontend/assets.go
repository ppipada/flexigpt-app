//go:build !codeanalysis

package frontend

import "embed"

//go:embed all:build
var Assets embed.FS

//go:embed public/icon.png
var Icon []byte
