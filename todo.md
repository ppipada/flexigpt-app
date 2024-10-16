# Project Current Milestones

## MileStone 0

### Scope

- Local installer for Linux/Mac/Win
- Top level feature: Chat, Conversation history
- Available during Chat: Prompt templates, Tools, KB servers

### Tasks

- Post migration list

  - [x] AI provider: get masked http details from langchaingo whenever available and then send in details
  - [ ] ~~Error wrapper and unwrapper on backend main and base apis~~
  - [x] Whether to stream or not should be decided by model in conjunction with provider in frontend.
  - [x] integrate file logger
  - [ ] Add support for langchaingo supported ai providers. Would need testing with each.
  - [ ] load chats properly at new chat etc. handle caching as needed
  - [ ] Better title deduce
  - [ ] Create dmg and flatpack and distribute for experimentation

- [ ] Prompt templates
- [ ] Tool use
- [ ] KB stores
