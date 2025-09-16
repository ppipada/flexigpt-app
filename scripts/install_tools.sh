#!/bin/bash

# Installs pinned Go tooling to $GOBIN (or $GOPATH/bin).
# Linux/macOS compatible

set -euo pipefail
IFS=$'\n\t'


GOLINES_VERSION="latest"
GOFUMPT_VERSION="latest"
GOPLS_VERSION="latest"
MODERNIZE_VERSION="latest" # Remember to bump build in .github/actions/lint.yml as well.
GOLANGCI_LINT_VERSION="v2.4.0"  # Remember to update .github/actions/lint.yml as well.
GO_MOD_UPGRADE_VERSION="v0.11.0"
WAILS_VERSION="v2.10.2"

# refdir:
# - Note: This tool can have false positives but can help sort Go functions in a file.
# - Actual command optiosn used are in package.json.
# - To fix a bug, we use patched version from: https://github.com/ppipada/refdir.
# - checkout the repo, compile locally as `go build` and then do `go install .`
REFDIR_VERSION="latest"

# Dependency inspector (graph):
# - Installs godepgraph (Go) and checks for Graphviz's 'dot'.
# - On macOS:   brew install graphviz
# - On Debian/Ubuntu: sudo apt-get update && sudo apt-get install -y graphviz
# -------------------------------
# - To generate a graph (example):
#   godepgraph -s -o github.com/ppipada/flexigpt-app,command-line-arguments cmd/agentgo/main.go | dot -Tpng -o godepgraph.png
GODEPGRAPH_VERSION="latest"

log() { printf '%s\n' "[$(date +'%H:%M:%S')] $*"; }
fail() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || fail "Missing required command: $1"; }

need go
GO_VERSION="$(go version)"
log "Using $GO_VERSION"

# Determine install bin directory
INSTALL_BIN="${GOBIN:-$(go env GOPATH)/bin}"
case ":$PATH:" in
  *":$INSTALL_BIN:"*) ;;
  *) export PATH="$INSTALL_BIN:$PATH"; log "Added $INSTALL_BIN to PATH for this session";;
esac

# Choose downloader for golangci-lint script

go_install() {
  # Usage: go_install <module-path> <version>
  local mod="$1"
  local ver="$2"
  log "go install ${mod}@${ver}"
  GOBIN="$INSTALL_BIN" go install "${mod}@${ver}"
}



# 1) golangci-lint (via official installer)
log "Installing golangci-lint ${GOLANGCI_LINT_VERSION} to ${INSTALL_BIN}"
# shellcheck disable=SC2086
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
  | sh -s -- -b "$INSTALL_BIN" "$GOLANGCI_LINT_VERSION"
command -v golangci-lint >/dev/null 2>&1 || fail "golangci-lint not found after install"
golangci-lint version

# 2) gopls
go_install "golang.org/x/tools/gopls" "$GOPLS_VERSION"
command -v gopls >/dev/null 2>&1 || fail "gopls not found after install"
gopls version || true  # Some versions may exit 0 without printing

# 3) modernize (from x/tools)
go_install "golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize" "$MODERNIZE_VERSION"
command -v modernize >/dev/null 2>&1 || fail "modernize not found after install"
modernize -V=full

# 4) go-mod-upgrade
go_install "github.com/oligot/go-mod-upgrade" "$GO_MOD_UPGRADE_VERSION"
command -v go-mod-upgrade >/dev/null 2>&1 || fail "go-mod-upgrade not found after install"

# 5) refdir
# go_install "github.com/devnev/refdir" "$REFDIR_VERSION"
# command -v refdir >/dev/null 2>&1 || fail "refdir not found after install"

# 6) golines
go_install "github.com/segmentio/golines" "$GOLINES_VERSION"
command -v golines >/dev/null 2>&1 || fail "golines not found after install"
# golines --version || true  # Uncomment if you want to print version (may vary by release)

# 7) gofumpt
go_install "mvdan.cc/gofumpt" "$GOFUMPT_VERSION"
command -v gofumpt >/dev/null 2>&1 || fail "gofumpt not found after install"
gofumpt -version || true

# 8) godepgraph (dependency graph helper)
go_install "github.com/kisielk/godepgraph" "$GODEPGRAPH_VERSION"
command -v godepgraph >/dev/null 2>&1 || fail "godepgraph not found after install"

# Graphviz check (optional)
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


# go mod install
go mod download
go mod tidy
go mod verify

log "All requested tools installed to: ${INSTALL_BIN}"
log "Reminder: Keep golangci-lint and (if applicable) modernize versions in sync with lint.yml"

log "running wails doctor"
wails doctor
