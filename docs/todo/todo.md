# Project TODO

## Laundry list

- [ ] edit and save template doesn't work. says template slug already exists.
  - [ ] most probably a mismatch between backend and frontend.

## Features

- [ ] tool features

  - [ ] need to list openai/anthropic etc given tools so that they can be invoked in api directly
  - [ ] web search tools better integration needs to be done e2e
  - [ ] some parallel tools can be created locally

- [ ] valid input output modalities, valid levels, valid reasoning types, etc need to be added to modelpresetspec.

- [ ] Token count in build completion data may be wrong as it doesn't really account for attachments/tool calls etc. Need to rectify the FilterMessagesByTokenCount function. Look at using gpt5 bpe encoder now rather than regex compilation. [go-tokenizer](https://github.com/tiktoken-go/tokenizer)
