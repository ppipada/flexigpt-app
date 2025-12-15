# Project TODO

## Laundry list

- [ ] edit and save template doesn't work. says template slug already exists.
  - [ ] most probably a mismatch between backend and frontend.
- [ ] View mode needed for builtins
- [ ] Need to migrate to content block array to support citations in conversation turns.

## Features

- [ ] tool features

  - [ ] Implementation of openai/anthropic etc built-in tools that have been analyzed and filtered.
  - [ ] Analysis of openai/anthropic etc given tools so that they can be invoked in api directly.

    - [ ] Openai - Do

      - [x] Function calling
      - [ ] Web search: integrate sources and citations too, in ui as well.
      - [ ] Image generation: integrate image render in UI too.

    - [x] Openai - Deferred

      - [x] File search: it is about using files uplaoded to openai vector stores. may look at it when we do vector stores thing.
      - [x] Remote MCP: decide when to do mcp integrations.
      - [x] Remote Connectors: decide when to do mcp integrations. This is better as it would give concrete access to remote data sources.

    - [x] Openai - Deferred/Think through

      - [x] Shell tool: this is a fixed schema for a host driven shell tool implementation. Better to have a much more controlled and safe and tunable in app shell calls and execute tool and flow. arbitrary shell commands in loop can be quite dangerous overall.
      - [x] Apply path tool: this is a fixed schema for a host driven patch generation and apply tool. it allows models to generate patches (mostly will be some internal schema conformance thing) and we have to apply them locally.
      - [x] The utility of these tools is mostly that you dont have to give input schema for this in function calling. This may be useful over host driven similar tools only if there is some cost saving associated with it.

    - [x] Openai - Don't

      - [x] Computer use: Clicks and image captures of screenshots from computer. Not sure about the utility of it yet.
      - [x] Code interpreter: about executing python in openai's own sandbox. not sure about the utility of it yet.
      - [x] Local shell tool: outdated, and available only on codex mini.

    - [ ] Anthropic - Do

      - [x] Function calling
      - [ ] Web search: integrate sources and citations too, in ui as well.

    - [x] Anthropic - Deferred

      - [x] Remote MCP connector: decide when to do mcp integrations.

    - [x] Anthropic - Deferred/Think through

      - [x] Bash tool: Similar to OpenAIs shell tool. Both recommend handing things in a "session".
      - [x] Text editor: This is near to patch apply tool in behavior but has some text like commands string replace and view etc.
      - [x] Web fetch: it is page fetch for content or pdf. For normal usecase the current local url fetch and send will be better in terms of state management and better processing and errors. This may be useful when doing web search integration. think about it when doing that.

    - [x] Anthropic - Don't

      - [x] Computer use: Same as openai.
      - [x] Code execution: Same as openai code interpreter with some free hours of usage.
      - [x] Programmatic tool calling: this is a convoluted looping of code execution and local tool calls. No usecase known yet.
        - [x] Simple mental model: Let Claude write a little Python script once, and run that script in a loop while it calls your tools, instead of asking Claude to think and call tools over and over.
        - [x] May be useful only to save tokens when there is a looped tool call need.
      - [x] Memory tool: This seems interesting on first pass. it is similar to text editor in flow, but allows to accumulate context locally on client. Would be interesting to see how to do context management in local app ourselves and how this tool helps with that?
      - [x] Tool search tool: This is a server side tool that allows claude to hold a tool library on its server and then search through it to get appropriate tool and then invoke it. It is very efficient in terms of token usage and can be checked on how to implement it locally.

  - [ ] parallel tools can be created locally
  - [ ] there is a reference calculator tool in claude docs
  - [ ] When url cannot fetch content, there is no way of knowing what happened as of now. may want to see hwo to expose this or disable link only mode in this flow?

- [ ] valid input output modalities, valid levels, valid reasoning types, etc need to be added to modelpresetspec.

- [ ] Token count in build completion data may be wrong as it doesn't really account for attachments/tool calls etc.
  - [ ] Need to rectify the FilterMessagesByTokenCount function. Look at using gpt5 bpe encoder now rather than regex compilation. [go-tokenizer](https://github.com/tiktoken-go/tokenizer)
  - [ ] Claude has a free token counting api with some rate limits. may need to check it out too.

## Milestone thoughts

- [ ] M1 - API coverage - Pending items:

  - [ ] built-in tools from apis
  - [ ] local replacements for some builtin tools that are very vendor specific

  - [ ] Modalities coverage:

    - [x] Text
      - [x] content in/out
      - [x] reasoning in/out
      - [x] extracted web pages input
      - [x] extracted pdf input
      - [ ] extracted other docs input, sheets and docx mainly.
    - [x] Image input
    - [x] Document input
    - [ ] Image url input
    - [ ] Document url input
    - [ ] Image output
    - [ ] Deferred: audio in/out

  - [ ] i18n
  - [ ] Dont: New stateful APIs and its hooks from vendors
    - [ ] stored responses, stored conversations, on server memory context, on server prompt templates etc.
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
