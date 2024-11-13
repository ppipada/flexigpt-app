# Project Current Milestones

## Ongoing Scope

- Milestone 0 completion
  - Top level feature: Chat, Conversation history, Settings
  - Local installer for Linux
- Milestone 1 partial
  - Tools, Prompts

## Laundry list

- [x] Math formulas in content needs to render properly
- [x] Mermaid diagrams and sequence diagrams should render properly
- [x] Adapt download button to wails
- [ ] Need a resend button too
- [ ] Cursor should move to text area after loading a chat or other chat actions
- [x] Scroll bug in input text area
- [ ] load chats properly at new chat etc. handle caching as needed
  - [ ] Wails app test doesnt show any delay. Need to test electron.
- [ ] Better title deduce
- [x] Flatpak installer for Linux
- [x] DMG installer for Mac
- [x] NSIS installer for Windows
- [ ] Mac install shows no title bar, better add a plain titlebar.
- [ ] Mac: Add how to install non signed pkg in readme. click > ok > settings > privacy and security > blocked > open anyway
- [ ] Mac icns file not present. Need proper icns file added
- [ ] Mac pkg install is add the .app in ~/build/bin. It should be in application. Check the settings, conversation and log location
- [ ] Test Windows build

## Tasks: Tools Implementation with CodeMirror

- Tools Page

  - Header

    - [ ] Design the header with the title "Tools."
    - [ ] Implement a search bar for tool searching.

  - Main Content Area:

    - [ ] Design the tool list in a card format.
    - [ ] Display tool name, short description, and last modified date on each card.
    - [ ] Add quick action icons/buttons for edit, delete, and duplicate.
    - [ ] Implement a "Create New Tool" button using a floating action button (FAB) or a clearly marked button.

- Create/Edit Tool Modal:

  - [ ] Design the modal layout for creating/editing tools.
  - [ ] Add a tool name field (required).
  - [ ] Add a description field (optional).
  - [ ] Implement a schema area using CodeMirror for defining the tool's schema.
    - [ ] Configure CodeMirror to support JSON syntax highlighting and validation.
    - [ ] Allow users to define input and output parameters using JSON schema.
  - [ ] Implement a function area using CodeMirror for accepting/implementing the schema.
    - [ ] Configure CodeMirror to support TypeScript syntax highlighting and validation.
    - [ ] Allow users to write an asynchronous TypeScript function that utilizes the defined schema.
  - [ ] Usability
    - [ ] Provide real-time validation and feedback for required fields.
    - [ ] Use tooltips or inline messages for guidance on schema and function implementation.
    - [ ] Ensure keyboard navigability.
    - [ ] Implement ARIA labels and roles for screen readers.
  - [ ] Action area:
    - [ ] Implement a "Save & Close" option.
    - [ ] Add a "Discard" button to exit without saving changes.

- Backend

  - [ ] Integrate with json store for tool storage.
  - [ ] Implement API endpoints for creating, retrieving, updating, and deleting tools.
  - [ ] Ensure schema validation and function execution are supported on the backend.

## Tasks: Prompt templates

- Prompt list page

  - Header

    - [ ] Design the header with the title "Prompt Templates."
    - [ ] Implement a search bar for prompt searching.

  - Main Content Area:

    - [ ] Design the prompt list in a card format.
    - [ ] Display prompt name, short description, short prompt start string and last modified date on each card.
    - [ ] Add quick action icons/buttons for edit, delete, and duplicate.
    - [ ] Implement a "Create New Prompt" button using a floating action button (FAB) or a clearly marked button.

- Create/Edit Prompt Modal:

  - [ ] Design the modal layout for creating/editing prompts.
  - [ ] Add a prompt name field (required).
  - [ ] Add a description field (optional).
  - [ ] Implement a large text area for entering the prompt template. This should be large enough and scrollable
  - [ ] May have preferred llm provider config (??)
  - [ ] Would need:
    - [ ] tools
    - [ ] KB
  - [ ] Usability
    - [ ] Provide real-time validation and feedback for required fields.
    - [ ] Use tooltips or inline messages for guidance on template strings.
    - [ ] Ensure keyboard navigability.
    - [ ] Implement ARIA labels and roles for screen readers.
  - [ ] Detect template strings and dynamically add them to a "Variables" section.
    - [ ] This should be below the scrollable area
    - [ ] Automatically populate variables section with detected template strings.
    - [ ] Implement fields for variable name, type dropdown, and default value.
  - [ ] Action area: should be below variables
    - [ ] Implement a "Save & Close" option.
    - [ ] Add a "Discard" button to exit without saving changes.

- Backend

  - [ ] Integrate with JSON file store per prompt
  - [ ] get list of prompts using the partitioned file store

- Integration of prompts in chat
  - [ ] Once defined all prompts should be available to use in chat via some keystroke (??).
  - [ ] Would also need some way to indicate if someone wants to add a prompt as a system prompt or just prompt
  - [ ] ~~Ideally if it is just prompt it should just expand the prompt in the input box~~
  - [ ] For any prompt with vars, the vars should be injected with defaults properly so that a user can edit them and frontend can parse it to create a proper string too.
  - [ ] ~~Can vars be few system functions like open file (??)~~
