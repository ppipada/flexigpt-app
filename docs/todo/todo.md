# Project TODO

## Laundry list

- [x] input area spacing and size when streaming back from assistant needs relook
- [ ] prompt and tool processing for ai completion call

  - [ ] only prompt should be allowed to be sent to ai rather than some text compulsory
  - [ ] for tools atleast prompt/text may be needed

- [ ] Think through and enhance tools create/edit and prompts create/edit modals

  - [ ] Implement a schema area using CodeMirror for defining the tool's/prompts schema. (deferred)
    - [ ] Configure CodeMirror to support JSON syntax highlighting and validation.
    - [ ] Allow users to define input and output parameters using JSON schema.
  - [ ] Implement a function area using CodeMirror for accepting/implementing the schema.
    - [ ] Configure CodeMirror to support TypeScript syntax highlighting and validation.
    - [ ] Allow users to write an asynchronous TypeScript function that utilizes the defined schema.

- Completions API backend

  - [ ] Integrate stores with llm. decide on if you want to populate the prompt in input so that var expansion etc can be done properly.

- [ ] CPU is high when "thinking". Need to debug
- [ ] After system restart if app is open it should start again, it is not happening in mac.
