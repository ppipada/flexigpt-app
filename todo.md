# Project TODO

## Laundry list

- [ ] implement prompts and tools using dirstore+fts
- [ ] spinner before first response (details if possible)
- [ ] post last streamed msg the scroll is moving a bit up
- [ ] Search extras
  - [ ] after load of prev conversation thecursor is still in box so for next time no one clicks or i can type again, but nothing happens there. ideally focus should move out so that user can come there again either via keyboard or click OR type should result in query
  - [ ] right now after one query I have to explicitly click to get things to show again

## UI integration for search fts of conversations

- Entry Point (Focus / Empty Query)

  - [ ] On focus, call `listConversations(pageSize=20)`
    - [ ] Render a “Recent conversations” list inside the dropdown.
  - [ ] Each row: title + last-modified date.
  - [ ] Up/Down / mouse-hover highlight; ⏎ or click navigates to the convo.

- Local Autocomplete (Query Length 1–2)

  - [ ] Filter the already-loaded titles in memory; update instantly.
  - [ ] Footer item: “Press Enter to search all messages” (inactive until length ≥ 3).
  - [ ] No backend round-trip.

- Debounced Full-Text Search (Query Length ≥ 3)

  - [ ] After 300 ms of inactivity fire `searchConversations(q, undefined, 20)`.
  - [ ] If another keystroke occurs, cancel the in-flight promise.
  - [ ] Present results in two sub-groups
    - [ ] 1. Title matches
    - [ ] 2. Message matches (with 1-line highlighted snippet).
  - [ ] Show spinner overlay while waiting; keep previous list dimmed for stability.

- Scroll-to-Load “More Results”

  - [ ] Dropdown body has its own scroll container (max-height ~60 vh).
  - [ ] When the user scrolls to 80 % of the container and `nextToken` exists
    - [ ] → automatically call `searchConversations(q, nextToken, 20)` and append rows.
  - [ ] Optional footer: “End of results” or “Loading more…”.

- Interaction Rules

  - [ ] up/down moves through visible items; focus wraps at top/bottom.
  - [ ] enter or click on an item → open convo.
  - [ ] If item originated from a message match, auto-scroll & briefly highlight that message. (not sure how?)
  - [ ] escape clears search & closes dropdown.
  - [ ] If the user presses enter with no item highlighted, do nothing (they must choose).

- Performance & Networking

  - [ ] Debounce: 300 ms; per-query request count rarely > 4.
  - [ ] Cache last 5 queries (query string → result set) for instant recall.
  - [ ] Re-use a single `AbortController` per search bar instance to cancel stale fetches.

- Edge & Empty States
  - [ ] No recent conversations → “No conversations yet” illustration.
  - [ ] No matches → “No results for ‘XYZ’”.
  - [ ] Network / SQLite error → inline error row + toast “Retry”.
  - [ ] Very long titles are ellipsized; full title appears in tooltip.

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
