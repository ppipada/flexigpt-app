# Project TODO

## Laundry list

- [ ] Lots of testing needed for new attachments.

  - [ ] Check debug details.
  - [ ] Attachment follow through across conversations.
  - [ ] cache token usage.
  - [ ] url attachments.

- Usage

  - [ ] With attachments, token usage display seems more important.
  - [ ] Token count in build completion data may be wrong as it doesnt really account for attachments/tool calls etc. Need to rectify the FilterMessagesByTokenCount function.
    - [ ] Look at using gpt5 bpe encoder now rather than regex compilation. https://github.com/tiktoken-go/tokenizer

- [ ] tool enhancements

  - [ ] attached tools to be given back as schema to consumers for exec
  - [ ] need to list openai/anthropic etc given tools so that they can be invoked in api directly
  - [ ] some parallell tools cna be created locally
  - [ ] web search tools better integration needs to be done e2e
