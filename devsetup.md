# Dev Setup and Commands

## Setup

- Install `go > 1.22` and `wails`
- Install `pnpm` and `lerna`
- Checkout code
- Download all go modules: `go mod download`
- Download all pnpm modules: `pnpm i`

## Backend only commands

- Build and run the backend: `./scripts/run_backend.sh`
- Backend OpenAPI 3.1 docs should be visible at: `http://localhost:8080/docs`

## Wails commands

- Build wails app and run it

```
pnpm run build:wails
./build/bin/flexigpt
```

## Electron commands

- run a electron dev server

```
pnpm run clean
pnpm run dev
```

- Build a electron prod appimage and run it

```
pnpm run build:electron
./agentts/build/flexigpt-linux-x86_64.AppImage
```

- Inspect built electron package

```
npx asar extract agentts/build/linux-unpacked/resources/app.asar ./asarex
```
