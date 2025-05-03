#!/bin/sh
set -e

# Export all variables from buildvars.env
set -a
. build/buildvars.env
set +a

echo "==== ENVIRONMENT VARIABLES ===="
env | sort
echo "==============================="

chmod +x build/flatpak/render_templates.sh
./build/flatpak/render_templates.sh --version $VERSION_TAG

mkdir -p "$XDG_CONFIG_HOME"
mkdir -p "$XDG_DATA_HOME"

PATH="$PATH:$GOROOT/bin:$GOBIN:$NODEROOT/bin" pnpm install
PATH="$PATH:$GOROOT/bin:$GOBIN:$NODEROOT/bin" "$GOROOT/bin/go" mod download
PATH="$PATH:$GOROOT/bin:$GOBIN:$NODEROOT/bin" pnpm run build:linux

# Cleanup unneeded files
rm -rf node_modules
rm -rf frontend/node_modules
rm -rf ~/.cache/go-build
rm -rf "$GOROOT/pkg"

# Continue with installation
install -Dm00755 build/bin/FlexiGPT "$FLATPAK_DEST/bin/FlexiGPT"
install -Dm00644 frontend/public/icon_256x256.png "$FLATPAK_DEST/share/icons/hicolor/256x256/apps/$FLATPAK_ID.png"
install -Dm00644 build/flatpak/"$FLATPAK_ID".desktop "$FLATPAK_DEST/share/applications/$FLATPAK_ID.desktop"
install -Dm00644 build/flatpak/"$FLATPAK_ID".appdata.xml "$FLATPAK_DEST/share/appdata/$FLATPAK_ID.appdata.xml"
