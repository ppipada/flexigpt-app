# Project TODO

## Laundry list

- [ ] Token count in build completion data may be wrong as it doesnt really account for attachments/tool calls etc. Need to rectify the FilterMessagesByTokenCount function. Look at using gpt5 bpe encoder now rather than regex compilation. [go-tokenizer](https://github.com/tiktoken-go/tokenizer)

- [ ] tool enhancements

  - [ ] need to list openai/anthropic etc given tools so that they can be invoked in api directly
  - [ ] web search tools better integration needs to be done e2e
  - [ ] some parallel tools can be created locally
  - [ ] attached tools to be given back as schema to consumers for exec

- [x] 5.1 codex max, deepseek 3.2
