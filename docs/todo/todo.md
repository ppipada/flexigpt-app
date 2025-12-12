# Project TODO

## Laundry list

- [ ] tool enhancements

  - [x] attached tools to be given back as schema to consumers for exec
  - [x] backend of tool calls and outputs from build to fetch to back
  - [x] testing of chip bars
  - [x] tool exec bars review
  - [x] tool choices also need to be added as chips
  - [x] anthropic needs thinking blocks too in response.
  - [x] message attachments bar needs to have similar dropdown info chips rather than one long bar.
  - [x] url attach is still leading to a submit.

  - [x] check for empty msg with only toolresult in fe. seems some issue
    - [x] abort and then edit again is also some issue.
  - [x] the mode menu looks cutoff after pushed inside the composer.
  - [x] tool choices need to be added as "previous tool choices" and an option to reattach needs to be given.

- [ ] edit and save template doesn't work. says template slug already exists.
  - [ ] most probably a mismatch between backend and frontend.

## Features

- [ ] Token count in build completion data may be wrong as it doesnt really account for attachments/tool calls etc. Need to rectify the FilterMessagesByTokenCount function. Look at using gpt5 bpe encoder now rather than regex compilation. [go-tokenizer](https://github.com/tiktoken-go/tokenizer)

- [ ] tool features

  - [ ] need to list openai/anthropic etc given tools so that they can be invoked in api directly
  - [ ] web search tools better integration needs to be done e2e
  - [ ] some parallel tools can be created locally

- [x] 5.1 codex max, deepseek 3.2, 5.2
- [x] Can now introduce a independent per message thinking block and manage it as separate input block in the api
- [ ] valid input output modalities, valid levels, valid reasoning types, etc need to be added to modelpresetspec.
