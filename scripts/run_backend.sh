#!/bin/sh

export SERVICE_HOST="localhost"
export SERVICE_PORT=8080
export SERVICE_SETTINGS_DIR_PATH="./out/settings"
export SERVICE_CONVERSATIONS_DIR_PATH="./out/conversations"
export SERVICE_MODEL_PRESETS_DIR_PATH="./out/modelpresets"
export SERVICE_PROMPT_TEMPLATES_DIR_PATH="./out/prompttemplates"
export SERVICE_TOOLS_DIR_PATH="./out/tools"
export SERVICE_LOGS_DIR_PATH="./out/logs"
export SERVICE_DEBUG="true"

# Run the Go application in the background
go run ./cmd/httpbackend &

# Capture the PID of the Go process
GO_PID=$!

if [ -z "$GO_PID" ] || ! kill -0 $GO_PID 2>/dev/null; then
	echo "No valid PID found. Exiting without opening the browser."
	exit 1
fi

echo "PID: ${GO_PID}"
# Function to clean up and exit
cleanup() {
	echo "Cleaning up..."
	kill $GO_PID 2>/dev/null
	exit
}

# Trap Ctrl+C (SIGINT) and call cleanup
trap cleanup INT

sleep 2

# Open the default web browser
if [ "$(uname)" = "Darwin" ]; then
	open "http://$SERVICE_HOST:$SERVICE_PORT/docs"
else
	xdg-open "http://$SERVICE_HOST:$SERVICE_PORT/docs"
fi

# Wait for the Go process to finish
wait $GO_PID

# Call cleanup when the Go process exits
cleanup
