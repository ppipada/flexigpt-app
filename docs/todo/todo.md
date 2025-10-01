# Project TODO

## Laundry list

- [x] input area spacing and size when streaming back from assistant needs re-look
- [x] Tool invocation in backend

  - [x] base support
  - [x] Don't add go tools at runtime i.e only compiled tools are there
  - [x] invocation tests for go
  - [x] invocation tests for http, inbuilt and externally added

- [ ] Tool invoke when doing prompt template preprocess run
- [ ] prompt templates to fill in full text including preprocessors and then give back message blocks to caller for exec.
- [ ] attached tools to be given back as schema to consumers for exec
- [ ] prompt and tool processing for ai completion call

- [ ] Think through and enhance tools create/edit and prompts create/edit modals

  - [ ] Implement a schema area using CodeMirror for defining the tool's/prompts schema. (deferred)
    - [ ] Configure CodeMirror to support JSON syntax highlighting and validation.
    - [ ] Allow users to define input and output parameters using JSON schema.
  - [ ] Implement a function area using CodeMirror for accepting/implementing the schema.
    - [ ] Configure CodeMirror to support TypeScript syntax highlighting and validation.
    - [ ] Allow users to write an asynchronous TypeScript function that utilizes the defined schema.

- Completions API backend

  - [ ] Integrate stores with llm. decide on if you want to populate the prompt in input so that var expansion etc can be done properly.

- [x] CPU is high when "thinking". Need to debug
- [x] After system restart if app is open it should start again, it is not happening in mac.
- [x] Implement responses api for openai new models and integrate gpt 5 codex too.
