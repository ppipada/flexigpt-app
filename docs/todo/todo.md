# Project TODO

## Laundry list

- [ ] edit and save template doesn't work. says template slug already exists.
  - [ ] most probably a mismatch between backend and frontend.
- [ ] View mode needed for builtins
- [ ] with more and more usage, atleast some tabs seem to be worthwhile to do. basically start a command in one and then switch to another. we would need to ensure that we dont really mix anything anywhere and streaming is proper but req wise it seems worthwhile to do.
- [ ] frameless app
- [ ] editor becomes heavy when lots of input content is there. verify and rectify.

## Features

- [ ] tool features

  - [ ] Implementation of openai/anthropic etc built-in tools that have been analyzed and filtered.
  - [ ] Analysis of openai/anthropic etc given tools so that they can be invoked in api directly.
  - [ ] ~~parallel tools can be created locally~~
  - [ ] there is a reference calculator tool in claude docs
  - [ ] When url cannot fetch content, there is no way of knowing what happened as of now. may want to see hwo to expose this or disable link only mode in this flow?

- [ ] valid input output modalities, valid levels, valid reasoning types, etc need to be added to modelpresetspec.

- [ ] Token count in build completion data may be wrong as it doesn't really account for attachments/tool calls etc.
  - [ ] Need to rectify the FilterMessagesByTokenCount function. Look at using gpt5 bpe encoder now rather than regex compilation. [go-tokenizer](https://github.com/tiktoken-go/tokenizer)
  - [ ] Claude has a free token counting api with some rate limits. may need to check it out too.

## Milestone thoughts

- [ ] M1 - API coverage - Pending items:

  - [ ] Modalities coverage:

    - [x] Text
      - [x] content in/out
      - [x] reasoning in/out
      - [x] extracted web pages input
      - [x] extracted pdf input
      - [ ] extracted other docs input, sheets and docx mainly.
    - [x] Image input
    - [x] Document input
    - [x] Image url input
    - [x] Document url input

  - [ ] Tools

    - [x] built-in tools from apis

      - [x] web search

    - [ ] local replacements for some builtin tools that are very vendor specific

      - [ ] bash
      - [ ] apply patch
      - [ ] text editor

    - [x] Dont: New stateful APIs and its hooks from vendors
      - [x] stored responses, stored conversations, on server memory context, on server prompt templates etc.

  - [ ] i18n
  - [ ] provider and model level allow disallow list of model params/capabilities etc.
  - [ ] Some more additional params in presets and advanced params modal.
    - [ ] tool choice tuning
    - [ ] verbosity tuning
    - [ ] top k
    - [ ] Not sure: Safety parameter, that identifies a user if they violate safety policies.
    - [ ] Not sure: stop strings
    - [ ] Not sure: cache control in claude

- [ ] M2 - Better context

  - [ ] MCP local connections and hooks
  - [ ] MCP options in apis connections and hooks
  - [ ] Doc stores/vector stores connections
    - [ ] Only if MCP cannot serve this.

- [ ] Agent Skills but via local "explorer" or "skills" flow???

- [ ] Deferred.
  - [ ] Image output: See inference-go notes.
  - [ ] audio in/out
