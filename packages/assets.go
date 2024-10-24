//go:build !codeanalysis

package assets

import "embed"

//go:embed all:frontend/build
var Assets embed.FS

//go:embed frontend/public/icon.png
var Icon []byte
