# Dev Setup and Commands

## Setup

- Checkout code.
- Init and checkout all submodules (currently only flatpak shared modules): `git submodule update --init --recursive`

- Install `go > 1.25`.

  - For Ubuntu you can do install via standard download and install OR via `asdf` or through the backports channel:
    - `sudo add-apt-repository ppa:longsleep/golang-backports`
    - `sudo apt update`
    - `sudo apt install golang-1.25`
    - You can see that go is installed in `/lib/go-1.25/bin/go`
    - Additional reference can be found at [Go Ubuntu Wiki](https://go.dev/wiki/Ubuntu)
  - Check Go is installed correctly: `go version`

- Install go tool prerequisites:

  - Run the script at `scripts/install_tools.sh`
  - Wails
    - If you have any issue in wails install, refer documentation page at [Wails Install Guide](https://wails.io/docs/gettingstarted/installation).
    - If you are using latest Linux version (example: Ubuntu 24.04) and it is not supporting `libwebkit2gtk-4.0-dev`, then you might encounter an issue in wails doctor: `libwebkit not found`.
    - To resolve this issue you can install `libwebkit2gtk-4.1-dev`. We do use the tag `-tags webkit2_41` during build.

- Install node and associated packages

  - Node versions can be managed using nvm.
  - Check the current required version in `package.json`. Verify using `node -v`.
  - Make sure corepack is enabled: `corepack enable`
  - Install pnpm: `corepack install`
  - Download all pnpm modules: `pnpm i`

- Run a build: `pnpm build:withbindings`

## Runner scripts

- All the available scripts are via pnpm scripts.
- You can check the `package.json` scripts section for all script available.
- A select few scripts are mentioned below.

- Build wails app in dev mode: `pnpm build:withbindings`

  - After this you will see the binary path at the end of build.
  - You can execute it directly. Dev console is available in this mode.
  - Debug console in the app is available via `Ctrl + Shift + F12`

- Build and run the backend: `pnpm run:gobackend`

  - Backend OpenAPI 3.1 docs should be visible at: `http://localhost:8080/docs`

- Build wails app and run it as a dev server: `pnpm run:watch`

  - Ideally this is to be used mainly for UI development via the local server running inside the browser.
  - Wails has some issues reloading both backend and frontend file changes together in some cases.

- Packaging commands

  - Ideally this should be done in github actions as secrets etc are setup there properly.
  - Create a flatpak on a local machine with some local defaults: `pnpm pack:flatpak`

## VSCode setup

Note: Similar can apply to other IDEs.

- Recommended plugins are at: `./.vscode/extensions.json`
  - You can check and install them at the notification time, or later by checking them via VSCode command pallette
- Settings config for VSCode is at: `./.vscode/settings.json`.
  - These get applied as workspace settings if you open this project in VSCode.
