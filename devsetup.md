# Dev Setup and Commands

## Setup

- Install `go > 1.23`. For Ubuntu you can do:
  - `sudo add-apt-repository ppa:longsleep/golang-backports`
  - `sudo apt update`
  - `sudo apt install golang-1.23`
  - You can see that go is installed in `/lib/go-1.23/bin/go`
  - Additional reference can be found at [Go Ubuntu Wiki](https://go.dev/wiki/Ubuntu)
- Check Go is installed correctly: `go version`
- Install `wails v2`. Documentation page is [here](https://wails.io/docs/gettingstarted/installation). Generally you need to do: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Install [pnpm](https://pnpm.io/installation) and [lerna](https://lerna.js.org/docs/getting-started)
- Checkout repo code
- Download all go modules: `go mod download`
- Download all pnpm modules: `pnpm i`
- All the available scripts are via pnpm scripts. You can check the `package.json` scripts section for all script available. A select few scripts are mentioned below.

## Backend only commands

- Build and run the backend: `pnpm run gobackend:run`
- Backend OpenAPI 3.1 docs should be visible at: `http://localhost:8080/docs`

## Wails commands

- Build wails app and run it: `pnpm run wails:run`
- Debug console in the app is available via `Ctrl + Shift + F12`

## Electron commands

- Dev build and run: `pnpm run dev`
- Prod build and run appimage: `pnpm run electron:run`
- Inspect built electron package: `npx asar extract agentts/build/linux-unpacked/resources/app.asar ./asarex`
