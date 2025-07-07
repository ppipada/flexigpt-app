# Project TODO

## Laundry list

- [ ] implement prompts and tools using dirstore+fts
- [ ] in modify modal, the search bar is visible i.e it doesnt get in background
- [ ] responses api integration for o3-pro. should be done with moving away from langchaingo item
- [ ] CPU is high when "thinking". Need to debug
- [ ] If there was some amount of data that was streamed, but error occured at end somepalce, the whole message goes away and only error is visible. Ideally you want to concat the issue at end and leave text as is.

## Random notes from UX discussions

- [ ] chat window should add a "persona" in left top that can be used as a loader of a "agent preset"
- [ ] The details of request response can be added as "info" button that allows for dig down in a modal rather than append in footer.
- [ ] need better pins in home

- [ ] The concept of "agent presets"/"assistants", which are not really app, but defined prompt templates + model presets + tools need to be there.

  - [ ] ChatSession always boots from an AgentPreset/Assistant; if user just picks a PT, the system spins an implicit AgentPreset with defaults.
  - [ ] This is still inside chat UI but, with a defined persona.
  - [ ] Prompt templates are "templates" with placeholders. No linking anywhere. At runtime, or save time there can be validations if these placeholders are proper or not, but lets say later someone deletes a tool etc. the template should not be modified or touched in any way. If a tool exec or tool or anything else is not available anytime, the template place can be left empty.

```shell
  PromptTemplate
    ├─ MessageBlocks      (role-tagged text)
    ├─ Variables          (type, default, validation)
    └─ PreProcessors[]    (may call helper ToolIDs)

  Assistant
    ├─ templateId         (FK -> PromptTemplate)
    ├─ modelPresetId      (already handled)
    ├─ toolBundleIds[]    ← LLM-callable tools live here
    └─ dataSourceIds[]    (RAG etc.)
```

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
  // May be we can have "Assistants" in place of skills too and all the below are ways to create assistants
  🧩 Skills             → Build & edit: (Below tabs in a expanded drawer).
                          1. Prompts
                          2. Tools
                          3. Model presets
                          4. Data/Doc Sources
                          5. Assistants is a preset of things from above 4 things.

  📊 Insights           → Usage, cost, performance dashboards
  ❓ Help               → Docs, tutorials, support
  ⚙️👤 Account           → Manage: (Below tabs in a expanded drawer)
                          1. Profile/Workspace
                          2. Billing
                          // May combine 3 and 4 if required, depends on density of info in each
                          3. App preferences: Themes, shortcuts, etc.
                          4. Security & Keys.
  ```

  ```mermaid
  graph TD
  %% ───────────────────────────────
  %% 1. MAIN SIDEBAR NAVIGATION
  %% ───────────────────────────────
  home[🏠 Home]
  chat[💬 Chat]
  apps[🟦 Apps]
  insights[📊 Insights]
  help[❓ Help]
  account[⚙️👤 Account]

  %% sidebar order (dashed to show UI order, not data-flow)
  home -.-> chat
  chat -.-> apps
  apps -.-> skillsSection
  skillsSection -.-> insights
  insights -.-> help
  help -.-> account


  %% ───────────────────────────────
  %% 2. PINNED / MARKETPLACE APPS
  %% ───────────────────────────────
  aiNotepad["🗒️ AI-Notepad"]
  imageGen["🖼️ Image-Gen"]

  apps --> aiNotepad
  apps --> imageGen


  %% ───────────────────────────────
  %% 3. SKILLS / ASSISTANTS AREA
  %% ───────────────────────────────
  subgraph skillsSection["🧩 Skills / Assistants"]
    prompts["Prompt Templates"]
    tools["Tools"]
    modelPresets["Model Presets"]
    dataSources["Data / Doc Sources"]
    assistants["Assistants<br/>(Agent Presets)"]
  end

  prompts --> assistants
  tools --> assistants
  modelPresets --> assistants
  dataSources --> assistants


  %% ───────────────────────────────
  %% 4. CHAT SESSION RELATION
  %% ───────────────────────────────
  chatSession["ChatSession<br/>(loads Persona)"]
  assistants -->|persona loader| chatSession


  %% ───────────────────────────────
  %% 5. ACCOUNT DRAWER
  %% ───────────────────────────────
  subgraph accountDetails["Account Sections"]
    profile["Profile / Workspace"]
    billing["Billing"]
    prefs["App Preferences"]
    security["Security & Keys"]
  end

  account --> profile
  account --> billing
  account --> prefs
  account --> security
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
