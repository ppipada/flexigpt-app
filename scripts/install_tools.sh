#!/bin/bash
#
# Installs pinned Go tooling to $GOBIN (or $GOPATH/bin).
# Linux/macOS compatible (Bash 3.2+ safe).
#
# Notes:
# - Tool dependencies are maintained below as (module path, version) tuples.
# - This script does NOT install the prerequisites: go, golangci-lint.
#   They must be installed first (use the tool versions file at repo root).
#   Remember to bump version in .github/actions/lint.yml when upgrading golangci-lint.

# - refdir:
#   - Note: This tool can have false positives but can help sort Go functions in a file.
#   - Actual command options used are in package.json.
#   - To fix a bug, we use patched version from: https://github.com/ppipada/refdir.
#   - checkout the repo, compile locally as `go build` and then do `go install .`

set -euo pipefail
IFS=$'\n\t'

log()  { printf '[%s] %s\n' "$(date +'%H:%M:%S')" "$*"; }
warn() { printf 'WARN: %s\n' "$*" >&2; }

# Check mandatory prerequisites explicitly requested:
# - go, golangci-lint (expected to be managed externally via the tool versions file)
MISSING_PREREQS=()
for t in go golangci-lint; do
  if ! command -v "$t" >/dev/null 2>&1; then
    MISSING_PREREQS+=("$t")
  fi
done

if ((${#MISSING_PREREQS[@]})); then
  warn "Missing prerequisites: ${MISSING_PREREQS[*]}"
	cat >&2 <<EOF
- Please install the prerequisites before running this script.
- A tool versions file can be found at repo root as .tool-versions
- Easiest way to install prerequisites is via asdf.
- Install asdf and relevant plugins and then run this script again.
- Alternatively, install the tools in any way that provides the pinned versions from .tool-versions
EOF
  exit 1
fi

TOOLS=(
"golang.org/x/tools/gopls@v0.20.0"
"github.com/wailsapp/wails/v2/cmd/wails@v2.11.0"
"github.com/oligot/go-mod-upgrade@v0.12.0"
"github.com/kisielk/godepgraph@v1.0.0"
# refdir (disabled) We rely on a patched fork; keeping the context here but do not install automatically.
# "github.com/devnev/refdir@latest"
)


# If we get here, prerequisites exist on PATH.
GO_VERSION="$(go version)"
log "Using $GO_VERSION"



go_install() {
  # Usage: go_install <module-path> <version>
  log "go install $1"
  go install $1
}

for entry in "${TOOLS[@]}"; do
  go_install $entry
done

MISSING_TOOLS=()
for t in gopls wails go-mod-upgrade godepgraph; do
  if ! command -v "$t" >/dev/null 2>&1; then
    MISSING_TOOLS+=("$t")
  fi
done

if ((${#MISSING_TOOLS[@]})); then
  warn "Missing tools found: ${MISSING_TOOLS[*]}"
  exit 1
fi


log "Reminder: Keep golangci-lint version in sync with lint.yml"
# -------------- go mod maintenance --------------
log "Running 'go mod download', 'go mod tidy', and 'go mod verify'"
go mod download
go mod tidy
go mod verify

log "running wails doctor"
wails doctor || true

# -------------- Optional: Graphviz check (for godepgraph) --------------
# Dependency inspector (graph):
# - Installs godepgraph (Go) and checks for Graphviz's 'dot'.
# - On macOS:   brew install graphviz
# - On Debian/Ubuntu: sudo apt-get update && sudo apt-get install -y graphviz
# -------------------------------
# - To generate a graph (example):
#   godepgraph -s -o github.com/ppipada/flexigpt-app,command-line-arguments cmd/agentgo/main.go | dot -Tpng -o godepgraph.png

if command -v dot >/dev/null 2>&1; then
  log "Graphviz 'dot' found: $(dot -V 2>&1)"
else
  log "Graphviz 'dot' not found. Install it for graphs:"
  if [[ "$(uname -s)" == "Darwin" ]]; then
    log "  brew install graphviz"
  else
    log "  sudo apt-get update && sudo apt-get install -y graphviz   (Debian/Ubuntu)"
  fi
fi


