# Project Current Milestones

## Ongoing Scope

- Milestone 0 completion
  - Top level feature: Chat, Conversation history, Settings
  - Local installer for Linux
- Milestone 1 partial
  - KB, Tools, Prompts

## MCP handling

- [x] JSONRPC

  - [x] jsonrpc message protocol implementation
  - [x] Supprot batch for jsonrpc
  - [x] stdio conn
  - [x] humaadapter for jsonrpc
    - [x] openapi docs for jsonrpc
  - [x] http transport for jsonrpc
  - [x] stdio transport

- [ ] MCP transport

  - [ ] mcp httpsse
  - [ ] stdio for mcp

- [ ] Add servers for: file, git, simplemapstore
- [ ] New conversation is created again and again even if empty
- [ ] if formatting reenabled is there in runtime input params prompt it is still sent as double

## Settings

- [ ] Need a settings data version and migration func

  - [ ] API key should never reach frontend after save

## Performance

- [ ] There is large cpu consumption sometime, debug
  - [ ] It is mainly seen if both wails dev and current app is open and being used wails://wails consumes 100%
  - [ ] Also check when one request was fired but not allowed to complete for some reason
  - [ ] Also, if the message thread gets very very big there seems to be a issue.
  - [ ] it is seen that post data streaming the cpu percentage shoots up. May be stream callback leaks or looped somewhere?
