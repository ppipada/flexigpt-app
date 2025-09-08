module github.com/ppipada/flexigpt-app

go 1.25.0

// We use golangci-lint, gopls modernize, gopls cli and go-mod-upgrade as tools too,
// but have not added here as dependency as noted in golangci-lint website
// Updating in linux
// 	- go: asdf install golang 1.25.0 ; asdf global golang 1.25.0
//  - golangci-lint: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.4.0
//       - update lint.yml for version too
//  - gopls: go install golang.org/x/tools/gopls@latest
//  - modernize: go install golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest; modernize -V=full
//       - update lint.yml for version increment too
//  - helper to upgrade modules: go-mod-upgrade

// Dependency inspector. Install godepgraph and graphwiz and run
// godepgraph -s -o github.com/ppipada/flexigpt-app,command-line-arguments  cmd/agentgo/main.go | dot -Tpng -o godepgraph.png
require (
	github.com/adrg/xdg v0.5.3
	github.com/anthropics/anthropic-sdk-go v1.9.1
	github.com/danielgtaylor/huma/v2 v2.34.1
	github.com/glebarez/go-sqlite v1.22.0
	github.com/google/uuid v1.6.0
	github.com/openai/openai-go/v2 v2.1.1
	github.com/philippgille/chromem-go v0.7.0
	github.com/wailsapp/wails/v2 v2.10.2
	github.com/zalando/go-keyring v0.2.6
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/spf13/cobra v1.9.2-0.20250831231508-51d675196729 // indirect
	github.com/spf13/pflag v1.0.8 // indirect
	modernc.org/libc v1.66.8 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/sqlite v1.38.2 // indirect
)

require (
	al.essio.dev/pkg/shellescape v1.6.0 // indirect
	github.com/bep/debounce v1.2.1 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jchv/go-winloader v0.0.0-20250406163304-c1995be93bd1 // indirect
	github.com/labstack/echo/v4 v4.13.4 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/leaanthony/go-ansi-parser v1.6.1 // indirect
	github.com/leaanthony/gosod v1.0.4 // indirect
	github.com/leaanthony/slicer v1.6.0 // indirect
	github.com/leaanthony/u v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/samber/lo v1.51.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/tkrajina/go-reflector v0.5.8 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/wailsapp/go-webview2 v1.0.21 // indirect
	github.com/wailsapp/mimetype v1.4.1 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)
