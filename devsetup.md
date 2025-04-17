# Dev Setup and Commands

## Setup

- Install `go > 1.24`. For Ubuntu you can do:
  - `sudo add-apt-repository ppa:longsleep/golang-backports`
  - `sudo apt update`
  - `sudo apt install golang-1.24`
  - You can see that go is installed in `/lib/go-1.23/bin/go`
  - Additional reference can be found at [Go Ubuntu Wiki](https://go.dev/wiki/Ubuntu)
- Check Go is installed correctly: `go version`
- Install `wails v2`. Documentation page is [here](https://wails.io/docs/gettingstarted/installation). Generally you need to do: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Install [pnpm](https://pnpm.io/installation)
- Checkout repo code
- Download all go modules: `go mod download`
- Download all pnpm modules: `pnpm i`
- All the available scripts are via pnpm scripts. You can check the `package.json` scripts section for all script available. A select few scripts are mentioned below.

## VSCode setup

Note: Similar can apply to other IDEs.

- Plugins:
  - Recommended plugins are at: `./vscode/extensions.json`
  - You can check and install them at the notification time, or later by checking them via VSCode command pallette
- Settings config for VSCode is at: `./vscode/settings.json`. These get applied as workspace settings if you open this project in VSCode.
- External CI tools: `golangci-lint`, `golines`

## Backend only commands

- Build and run the backend: `pnpm run:gobackend`
- Backend OpenAPI 3.1 docs should be visible at: `http://localhost:8080/docs`

## Wails commands

- Build wails app and run it as a dev server: `pnpm run:watch`
- Debug console in the app is available via `Ctrl + Shift + F12`

## Build commands

- Create a flatpak on a local machine: `pnpm pack:flatpak`
