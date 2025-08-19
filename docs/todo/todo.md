# Project TODO

## Laundry list

- [ ] Chat input UI

  - [x] Add system prompt array to main bar.
  - [x] Move input to editor from text field
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
