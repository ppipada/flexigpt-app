#!/bin/bash

export SERVICE_HOST="localhost"
export SERVICE_PORT=8080
export SERVICE_SETTINGS_DIR_PATH="./out/settings"
export SERVICE_CONVERSATIONS_DIR_PATH="./out/conversations"
export SERVICE_LOGS_DIR_PATH="./out/logs"
export SERVICE_DEFAULT_PROVIDER="openai"
export SERVICE_DEBUG="true"

go run ./backendgo
