# Project TODO

## Laundry list

- [ ] implement prompts and tools using dirstore+fts
- [ ] in modify modal, the search bar is visible i.e it doesnt get in background
- [ ] details loading is very slow as of now
- [ ] responses api integration for o3-pro. should be done with moving away from langchaingo item

- [ ] Refactor provider in prep for presets and skills flow
  - [x] Types in aiprovider
  - [x] We should store all these preset files in settings or their own domain folders?
  - [x] Remove confusion of ModelParams vs ModelSettings vs ModelPreset. What is stored is a ModelPreset. What comes in is a ModelParam.
  - [ ] Default model also should be with presets, with name being "default preset".
  - [ ] Check if inbuilt models have tooltip mentioning internal defaults
  - [ ] Presets should have shortcommand as unique, may be add id too as forward compatible thing.
  - [x] font of settings page
  - [ ] api key in settings page is not validated before send and then set
  - [ ] no config in providers gives empty unhandled page rejection in console

## Random notes from UX discussions

- [ ] chat window should add a "persona" in left top that can be used as a loader of a "agent preset"
- [ ] model preset can be laoded in input bar rather than just model. may need to figure out how to show/allwo edits to the persona vs model preset
- [ ] The details of request response can be added as "info" button that allows for dig down in a modal rather than append in footer.
- [ ] need better pins in home
- [ ] Side bar:

  ```text
  ──────────────────── (Top static)
  🏠 Home                → Landing page, Recent activity, Dashboards
  💬 Chat                → Chat UI, Conversation lists

  ──────────────────── (Mid dynamic, Min 8px spacer above)

  🟦 Apps                → Grid & marketplace of installable apps
  🗒️ AI-Notepad          → Example pinned app
  🖼️ Image-Gen           → Example pinned app
  // max 5 pinned-app icons, drag to rearrange list

  ──────────────────── (Mid dynamic, Min 8px spacer below)

  ──────────────────── (Bottom static)
  🧩 Skills             → Build & edit: (Below tabs in a expanded drawer)
                          1. Prompts
                          2. Tools
                          3. Model presets
                          4. Data/Doc Sources
  📊 Insights           → Usage, cost, performance dashboards
  ❓ Help               → Docs, tutorials, support
  ⚙️👤 Account           → Manage: (Below tabs in a expanded drawer)
                          1. Profile/Workspace
                          2. Billing
                          // May combine 3 and 4 if required, depends on density of info in each
                          3. App preferences: Themes, shortcuts, etc.
                          4. Security & Keys.
  ```

## Tasks: Tools Implementation with CodeMirror

- Tools Page

  - Header

    - [x] Design the header with the title "Tools."
    - [ ] ~~Implement a search bar for tool searching.~~

  - Main Content Area:

    - [ ] ~~Design the tool list in a card format.~~
    - [ ] ~~Display tool name, short description, and last modified date on each card.~~
    - [x] Add quick action icons/buttons for edit, delete, ~~and duplicate~~.
    - [x] Implement a "Create New Tool" button using a floating action button (FAB) or a clearly marked button.

- Create/Edit Tool Modal:

  - [x] Design the modal layout for creating/editing tools.
  - [x] Add a tool name field (required).
  - [ ] ~~Add a description field (optional).~~
  - [ ] Implement a schema area using CodeMirror for defining the tool's schema.
    - [ ] Configure CodeMirror to support JSON syntax highlighting and validation.
    - [ ] Allow users to define input and output parameters using JSON schema.
  - [ ] Implement a function area using CodeMirror for accepting/implementing the schema.
    - [ ] Configure CodeMirror to support TypeScript syntax highlighting and validation.
    - [ ] Allow users to write an asynchronous TypeScript function that utilizes the defined schema.
  - [x] Usability
    - [x] Provide real-time validation and feedback for required fields.
    - [x] Use tooltips or inline messages for guidance on schema and function implementation.
    - [x] Ensure keyboard navigability.
    - [x] Implement ARIA labels and roles for screen readers.
  - [x] Action area:
    - [x] Implement a "Save & Close" option.
    - [x] Add a "Discard" button to exit without saving changes.

- Backend

  - [ ] Integrate with json store for tool storage.
  - [ ] Implement API endpoints for creating, retrieving, updating, and deleting tools.
  - [ ] Ensure schema validation and function execution are supported on the backend.

## Tasks: Prompt templates

- Prompt list page

  - Header

    - [x] Design the header with the title "Prompt Templates."
    - [ ] ~~Implement a search bar for prompt searching.~~

  - Main Content Area:

    - [ ] ~~Design the prompt list in a card format.~~
    - [ ] ~~Display prompt name, short description, short prompt start string and last modified date on each card.~~
    - [x] Add quick action icons/buttons for edit, delete, and ~~duplicate~~.
    - [x] Implement a "Create New Prompt" button using a floating action button (FAB) or a clearly marked button.

- Create/Edit Prompt Modal:

  - [x] Design the modal layout for creating/editing prompts.
  - [x] Add a prompt name field (required).
  - [ ] ~~Add a description field (optional).~~
  - [ ] Implement a large text area for entering the prompt template. This should be large enough and scrollable
  - [ ] May have preferred llm provider config (??)
  - [ ] Would need:
    - [ ] tools
    - [ ] KB
  - [x] Usability
    - [x] Provide real-time validation and feedback for required fields.
    - [x] Use tooltips or inline messages for guidance on template strings.
    - [x] Ensure keyboard navigability.
    - [x] Implement ARIA labels and roles for screen readers.
  - [ ] Detect template strings and dynamically add them to a "Variables" section.
    - [ ] This should be below the scrollable area
    - [ ] Automatically populate variables section with detected template strings.
    - [ ] Implement fields for variable name, type dropdown, and default value.
  - [x] Action area: should be below variables
    - [x] Implement a "Save & Close" option.
    - [x] Add a "Discard" button to exit without saving changes.

- Backend

  - [ ] Integrate with JSON file store per prompt
  - [ ] get list of prompts using the partitioned file store

- Integration of prompts in chat
  - [ ] Once defined all prompts should be available to use in chat via some keystroke (??).
  - [ ] Would also need some way to indicate if someone wants to add a prompt as a system prompt or just prompt
  - [ ] ~~Ideally if it is just prompt it should just expand the prompt in the input box~~
  - [ ] For any prompt with vars, the vars should be injected with defaults properly so that a user can edit them and frontend can parse it to create a proper string too.
  - [ ] ~~Can vars be few system functions like open file (??)~~
