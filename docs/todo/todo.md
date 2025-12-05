# Project TODO

## Laundry list

- Usage

  - [ ] With attachments, token usage display seems more important.
  - [ ] Token count in build completion data may be wrong as it doesnt really account for attachments/tool calls etc. Need to rectify the FilterMessagesByTokenCount function.
    - [ ] Look at using gpt5 bpe encoder now rather than regex compilation. [go-tokenizer](https://github.com/tiktoken-go/tokenizer)

- check if extraction pipeline for PDF urls is handled properly.
- [ ] tool enhancements

  - [ ] need to list openai/anthropic etc given tools so that they can be invoked in api directly
  - [ ] web search tools better integration needs to be done e2e
  - [ ] some parallel tools can be created locally
  - [ ] attached tools to be given back as schema to consumers for exec
