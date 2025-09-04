# Project TODO

## Laundry list

- [ ] Chat input UI

  - [x] Add system prompt array to main bar.
  - [x] Move input to editor from text field

  - [ ] prompt/tool inline elements:

    - [x] slash command plumbing
    - [x] prompt templates via slash command

    - [x] prompt visual filling via toolbar

      - [x] prompt toolbar/editing implementation
      - [x] system prompt updates from prompts
      - [x] may be allow expand/decompose for prompts user message
      - [x] toolbar outline and stacking with shadows and round corners styling

    - [ ] prompt and tool processing for ai completion call

      - [ ] only prompt should be allowed to be sent to ai rather than some text compulsory
      - [ ] for tools atleast prompt/text may be needed

    - [ ] tools via slash command
    - [x] check if proper memo etc is there for usetemplates and use tools etc

    - [ ] tools slash command hooking, tools should be added to bottom of the editor rather than inline

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
