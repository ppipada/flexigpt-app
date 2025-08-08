# Project Done Items

- Design and implement AppLayout

  - [x] Manages overall layout structure

  - [x] Sidebar on the left

    - [x] Handles click events
    - [x] Settings icons at the bottom
    - [x] Should be hidden behind a hamburger on small screen. Always visible on large.
    - [x] Clicking icons should redirect to corresponding pages

- Theming

  - [x] Define and apply dark and light themes
  - [x] Applies consistent theming across all components

- Chat page

  - [x] a navbar with new chat button on left, export button on right, search in center

    - [x] Search should list conversations that match the search input. Also should behave like a autocomplete

  - [x] send message input field with a "send" button

    - [x] It should have "type message..." as hint
    - [x] send button should be enabled only when something is typed
    - [x] styled as a box with top as circular
    - [x] fixed at bottom of the parent container
    - [ ] Input box styling issues. should be centered on medium screens and small screens too.

  - [x] chat message list

    - [x] These are in the body, below navbar and above input
    - [x] is a card layout scrollable vertically and can have user and then assistant

  - [x] Each chat bubble should be a card

    - [x] It should have avatar/icon on left/right depending on the user
    - [x] Card body should support code with highlighting.
      - [x] Code block should have a header that has language on right, copy button on left.
      - [x] It should have scroll horizontal support.
    - [x] Footer:
      - [x] It should support copy, edit, button (depends on type of user) on left side.
      - [x] after it, there should be a "feedback" button. that should open a input text field with a send button.
        - [x] It should have a like dislike button on left side
        - [x] people can add free text feedback in input field
      - [x] Footer also should have a dropdown on right, which opens a code block on click, that is json sent of the request/response.

  - [ ] Advanced:
    - [ ] ~~Support add file, add image~~
    - [ ] ~~Support streaming response~~

- Electron support

  - [x] Support electron based bundling for desktop
  - [x] Test with prod build and run appimage.
  - [ ] ~~Add icon in title bar and ass appIcon for tray~~
  - [x] Add a desktop entry so that appimage can be registered with the system

- Settings screen

  - [x] A settings page is a multi card page.
  - [x] The settings page has a multiple cards and a "export settings" functionality
  - [x] Support all settings from vscode extension
  - [ ] ~~"import settings" functionality too~~
  - Card widget:
    - [x] Each card widget has a header i.e the settings name
    - [x] The body is a table of key value pairs
    - [x] The first row is a fixed frozen row with colum names i.e column1: key, column 2: value
    - [ ] ~~There are then rows of "predefined keys", these rows are not deleteable. Also the keys are not editable, only the values are.~~
    - [ ] ~~There should be capability to add and remove other rows.~~
    - [x] There is should be a capability to display a value as "hidden" i.e "\*\*\*", based on some input type (e.g: password or secret or apikeys)
    - [ ] ~~There should be a explicit "save settings" button too. That can have a dummy implementation for now. Ideally this should be in card footer.~~

- Settings State DB

  - [x] Create a state db that can store and retrieve settings
  - [x] Connect all above widgets to a background settings state lib (ideally json file based)
  - [ ] ~~The lib should provide CRUD functionality with some keys being immutable (for above "predefined keys")~~
  - [ ] ~~It should also provide and import export functionality. The export should be as a json file, same with import.~~

- LLM Provider

  - [x] Implement llm providers
  - [x] connect this with chat page

- Conversation State DB:

  - [x] Add a "conversations state db".
  - [x] It should provide a functionality of add a conversation, remove a conversation, list conversation titles, remove all conversations
  - [x] A conversation has a title, id (uuidv7), created time, modified time, list of messages
  - [x] For each conversation it should have a add message, list all messages interface.
  - [x] Each conversation should be saved as a "json file" in some predefined folder.
  - [x] Connect UI to this state db

- Conversation list

  - [x] Fetch conversation titles and add it to the drawer
  - [x] provide a new chat, delete one conversation, delete all conversations actions

- [x] code inline line overflow handle
- [x] feedback integration removal
- [x] sort order of drop down, during load
- [x] sort order of drop down, after new chat
- [x] title population
- [x] Full copy for assistant
- [x] Streaming
  - [x] aiprovider support
  - [x] page support
  - [x] add other providers
    - [x] Openai
    - [x] anthropic
    - [x] google
  - [x] someway to show things are loading
- [x] Content:
  - [x] Emoji
  - [ ] ~~Subscript~~
  - [ ] ~~custom headings~~
  - [ ] ~~navigate in same doc~~
  - [ ] ~~Math handling~~
- [x] packaged app debug
- [x] Move ai provider to electron main
- [x] screen blanks if lot of data is put in inputfield
- [x] selection doesnt reset on new conversation in search bar
- [x] today vs yesterday in group
- [x] logger is imported in securejsondb before it is set as ipc logger
- [x] error handling review and fix
- [x] Blanket error handle in api fetch chat helper
- [x] Add edit functionality
- [x] ability to discard previous message history in next call
- [x] ability to choose provider and model and temperature in the chat itself.
- [x] populate initial choices in input options
- [x] Take default model and temp from settings without caching
- [x] focus on edit box if someone presses edit. as of now focus is at end of page
- [x] In chat the last used model should be stored ideally so that next time you load the chat it loads that model by default
- [x] url in output and input should be clickable and copiable
- [x] Enable right click context menu
- [x] word wrap for single word if it is too long.
- [x] scroll handling for edit text area
- [x] Icon of user and assistant should be near bottom of card rather than at top
- [x] Mac install shows no title bar, better add a plain titlebar.
- [x] Mac: Add how to install non signed pkg in readme. click > ok > settings > privacy and security > blocked > open anyway
- [x] Mac icns file not present. Need proper icns file added
- [x] chat load is not happening before a request fire
- [x] if the app is on for a long time, looks like the conversation list is removed
- [x] there is some jerkiness in chat load for today

- Wails migrate

  - [x] Translate securejsondb to go
  - [x] Translate settings store and conversation store to go
  - [x] Move next frontend to appropriate folder and change config and build system as seen in wails template

  - [x] Transform next app to work with multiple backend provider like electron or wails.

    - [x] Involves segregating models and creating a facade to redirect to window apis that are injected by backend
    - [x] Ideally interface declaration should be present in frontend and all calls should be made to interface by components, the exact implementation should decide which backend is present and redirect there.
    - [x] wails main.go app integrations
    - [x] Logger integration

  - [x] Translate aiprovider to go. Try and use langchaingo directly
  - [x] Implement streaming.
  - [x] Integrate other AI providers: anthropic, google, hf

- [x] Add support for first getting available ai providers and then configuring them accordingly. UI needs to be modified as collapsed cards and disabled cards
- [x] Only enabled chat options should be available
- [x] Get most consts through API
- [x] AI provider: get masked http details from langchaingo whenever available and then send in details
- [ ] ~~Error wrapper and unwrapper on backend main and base apis~~
- [x] Whether to stream or not should be decided by model in conjunction with provider in frontend.
- [x] integrate file logger
- [x] Math formulas in content needs to render properly
- [x] Mermaid diagrams and sequence diagrams should render properly
- [x] Adapt download button to wails
- [x] Need a resend button too
- [x] Cursor should move to text area after loading a chat or other chat actions
- [x] Scroll bug in input text area
- [x] load chats properly at new chat etc. handle caching as needed
  - [x] Wails app test doesnt show any delay. Need to test electron.
- [x] Flatpak installer for Linux
- [x] PKG installer for Mac
- [x] NSIS installer for Windows
- [x] Wails save file dialog opens and saves but file not seen on system
- [x] Remove electron
- [x] Remove nextjs and use react router v7

  - [x] migrate to react 19
  - [x] use vite
  - [x] migrate to eslint 9
  - [x] use spa mode
  - [x] remove postcss
  - [x] remove eslint, prettier, knip from top

- [x] hydrationfallback addition
- [x] use loaders to do client loading of pre dom loading content like settings
- [x] Buttons in nav bar changed styling during migration, rectify
- [x] font google local source
- [x] Need an about section someplace with version of software, dev name etc
  - [x] Added simple version in title bar for now.
- [x] For user message, do not render as markdown, render as is.
- [x] Support developer message for openai. Add Formatting re-enabled on first line of developer message in >o1 models for md output.

- [x] Add models as a configuration in providers where each model can be enabled disabled as required. Add a custom model entry support too.

  - [x] UI with card table and actions
  - [x] Modal for edit/add
  - [x] Backend integration
  - [x] Do not delete inbuilt models, just disable them.
  - [x] Input options temperature should come from either modelparams or modelinfo
  - [x] Link add model
  - [x] Add a custom provider and model name setting too so that an OAI compatible provider and model can be added
  - [ ] Add a new provider flow in providerSet
    - [x] engine delete end to end
    - [x] chat completion prefix end to end
    - [x] need promptlength and output as max in setting and param need to be similar in names, info need to be about "global defaults"
    - [x] setting provider attr for custom providers
      - [x] At load time need to init default + settings specific custom providers and models
      - [x] Better do it inside go rather than from frontend. may solve for api key thing too
    - [x] Need a dynamic time key encoder decoder for secrets
    - [x] Dont allow disabling default model
    - [x] atleast one model should be enabled
    - [x] default should have only enabled models
    - [x] reasoning models dont have tick as of now
    - [x] add buttons should be disabled until add reqs are met.
  - [x] Chat options is showing the new default model as default + the actual default model as tick mark
  - [x] delete of custom new providers need to be there similar to models
  - [x] add a note in add provider that only OAI compatible api can be served

- [x] Overflow issue in parent i.e whole page sometimes moves up
- [x] Param input content should preserve tabs/spaces etc

- [x] allow dot and spaces in providername and model name
- [x] reasoning models should have a "reasoning" effort param too.

  - [x] Add reasoning support in modelparams
  - [x] reasoning effort comes in multiple flavors: medium high less etc. number from x to y. find a way to represent them.
  - [x] See if temperature is almost always optional. Anthropic supports reasoning + normal mode too. Need to see how to represent that too without real overwhelming of the user. Represent appropriately in UI

    - [x] For anthropic: temp and thinking are not compatible. streaming is present. thinking is controlled as "budget tokens "
    - [x] For openai: temp and thinking are not compatible. can pass temp as 1 to some models after dec 24. streaming is present in newer api. thinking is controlled as effort, low medium high
    - [x] google doesnt mention any control on thinking model

  - [x] created a fork and try to merge pr for anthropic and openai in until we move away from langchaingo.

- [x] Support additional params from UI to backend
- [x] Represent reasoning process in UI
- [x] Conversations sometime dont load at init from homepage
- [x] deprecate 4.5preview
- [x] Home page overflows some times
- [x] Chat input fields scroll height mismatches with textarea
- [x] move packaging to build dir and add apple plist etc. also use apple developerid
- [x] Inter font files need to be cached in repo
  - [x] Did this by using font source npm
- [x] Signing for Mac pkg and Win
- [x] Mac pkg install is add the .app in ~/build/bin. It should be in application. Check the settings, conversation and log location
- [ ] ~~Ideally should detect any code or math blocks even in input, annotate them and then display. i.e format a normal text into proper md.~~
- [x] Math support broke in between, do math rendering before gfm
- [x] see how to set version increment seamlessly at one place only

  - [x] Version string to be passed via ldconfig to code and use from there
  - [x] render templates for flatpak and mac

- [x] JSONRPC

  - [x] jsonrpc message protocol implementation
  - [x] Supprot batch for jsonrpc
  - [x] stdio conn
  - [x] humaadapter for jsonrpc
    - [x] openapi docs for jsonrpc
  - [x] http transport for jsonrpc
  - [x] stdio transport

- [x] New conversation is created again and again even if empty
- [x] Input param validation: If formatting reenabled is there in runtime input params prompt it is still sent as double
- [x] Icon resolutions: revalidate icon and its resolutions available vs used

- [x] Initial inbuilt models should be available by default without a entry in settings.
- [x] Settings should be overrides and new/custom models/providers

- [x] Performance: There is large cpu consumption sometime, debug

  - [x] It is mainly seen if both wails dev and current app is open and being used wails://wails consumes 100%
  - [x] Also check when one request was fired but not allowed to complete for some reason
  - [x] Also, if the message thread gets very very big there seems to be a issue.
  - [x] it is seen that post data streaming the cpu percentage shoots up. May be stream callback leaks or looped somewhere?
  - [x] check again after installing new version for cpu issues

- [x] Conversation collection month provider has hardcoded partitions that returns current month only. Need to derive month from filename
- [x] katex and math symbols are on top of the search bar
- [ ] ~~If user has scrolled up, dont keep scrolling at bottom. this seems to introduce some unwanted cpy~~
- [x] Use shiki instead of react syntax highlighter
- [x] Nvidia issue debug and Ubuntu mate env vars fix
- [x] Move syntax highlighting to a worker

- [x] Better title deduce

  - [x] Use compromise and stopword for heuristic based titles for now

- [x] support events in simplemapdb
- [x] add fts using sqlite for simplemapdb
- [x] test and integrate fts with conversations
- [x] mtime should be added to cols in conversations and it should be used to do a incremental walk along with path used as externalID
- [x] the engine somehow has path etc as inbuilt. It looks very tightly coupled to files or something called as path. Ideally it should have opaque and its own semantics that the consumer will adapt to
- [x] Title change in chat causes new file to be created. ideally we want to only one file.

  - [x] separate out put conversation and add messages to conversation so that things are idempotent and id based comparisons can be made in put
  - [x] need listfiles with filtering in dirstore

- Entry Point (Focus / Empty Query)

  - [x] On focus, call `listConversations(pageSize=20)`
    - [x] Render a "Recent conversations" list inside the dropdown.
  - [x] Each row: title + last-modified date.
  - [x] Up/Down / mouse-hover highlight; ⏎ or click navigates to the convo.

- Local Autocomplete (Query Length 1-2)

  - [x] Filter the already-loaded titles in memory; update instantly.
  - [x] Footer item: "Press Enter to search all messages" (inactive until length ≥ 3).
  - [x] No backend round-trip.

- Debounced Full-Text Search (Query Length ≥ 3)

  - [x] After 300 ms of inactivity fire `searchConversations(q, undefined, 20)`.
  - [x] If another keystroke occurs, cancel the in-flight promise.
  - [x] Present results in two sub-groups
    - [x] 1. Title matches
    - [x] 2. Message matches (with 1-line highlighted snippet).
  - [x] Show spinner overlay while waiting; keep previous list dimmed for stability.

- Scroll-to-Load "More Results"

  - [x] Dropdown body has its own scroll container (max-height ~60 vh).
  - [x] When the user scrolls to 80 % of the container and `nextToken` exists
    - [x] -> automatically call `searchConversations(q, nextToken, 20)` and append rows.
  - [x] Optional footer: "End of results" or "Loading more".

- Interaction Rules

  - [x] up/down moves through visible items; focus wraps at top/bottom.
  - [x] enter or click on an item -> open convo.
  - [ ] ~~If item originated from a message match, auto-scroll & briefly highlight that message. (not sure how?)~~
  - [x] escape clears search & closes dropdown.
  - [ ] ~~If the user presses enter with no item highlighted, do nothing (they must choose).~~

- Performance & Networking

  - [x] Debounce: 300 ms; per-query request count rarely > 4.
  - [x] Cache last 5 queries (query string -> result set) for instant recall.
  - [x] Re-use a single `AbortController` per search bar instance to cancel stale fetches.

- Edge & Empty States

  - [x] No recent conversations -> "No conversations yet" illustration.
  - [x] No matches -> "No results for ‘XYZ’".
  - [x] Network / SQLite error -> inline error row + toast "Retry".
  - [x] Very long titles are ellipsized; full title appears in tooltip.

- [x] Search extras

  - [x] after load of prev conversation the cursor is still in box so for next time no one clicks or i can type again, but nothing happens there. ideally focus should move out so that user can come there again either via keyboard or click OR type should result in query
  - [x] right now after one query I have to explicitly click to get things to show again, i.e click out and then click in again.
  - [x] Search fully when only few chars is broken

- [x] post last streamed msg the scroll is moving a bit up
- [x] spinner before first response ~~(details if possible)~~
- [x] When scroll kicks in the font appears to be bolder. This is due to gpu composting issue with webkit. Put antialiasing on body to make experience consistent.

- [x] Refactor provider in prep for presets and skills flow

  - [x] Types in aiprovider
  - [x] We should store all these preset files in settings or their own domain folders?
  - [x] Remove confusion of ModelParams vs ModelSettings vs ModelPreset. What is stored is a ModelPreset. What comes in is a ModelParam.
  - [x] Default model also should be with presets, with name being "default preset".
  - [x] Check if inbuilt models have tooltip mentioning internal defaults
  - [x] Presets should have id/slug as unique, may be add id too as forward compatible thing.
  - [x] font of settings page
  - [x] api key in settings page is not validated before send and then set
  - [x] no config in providers gives empty unhandled page rejection in console
  - [x] use preset in input field and completion
  - [x] model preset can be loaded in input bar rather than just model. may need to figure out how to show/allow edits to the persona vs model preset

- [x] Integrate presets end to end
- [x] details loading is very slow as of now
- [x] request details are not visible in ui as of now.
- [x] input text box seems to cause some cpu consumption and jerkiness.
  - [x] autoresize is the issue.
  - [ ] ~~may be memoize other toolbar components too~~
- [x] Click on zoom for mermaid diagram

- [x] in modify modal, the search bar is visible i.e it doesnt get in background
- [x] Input field height should go to 16 rows most probably
- [x] prompt page table is not outlined

- [x] Add builtin with overlay
- [x] Integrate builtin into prompt template store
  - [x] fts and search pending
- [x] Integrate the store with UI
- [x] Startup issue in linux mint if sqlite file is missing.
- [x] Details expansion was giving large cpu
- [x] Ideally conversation list should be sorted based on modified time and not created time.

- [x] implement prompts and tools using dirstore+fts
- [x] set default provider is pending
- [x] Inbuilt model handling is pending in modelpresets.
- [x] May be in settings too, check.
- [x] In hybrid mode, temperature setting is not coming for sonnet etc

- [x] Add other daisyui themes to settings and then use them as a dropdown
- [x] theming adjustments for mermaid and code
- [x] scroll non needed on assistents reply, let it be where it is.
- [x] the scroll down buttons possition needs adjusting a bit
- [x] Scroll button needs to become visible as soon as some streamed message arrives

- Settings/Presets/Prompt/tool UI

  - [x] Add Auth keys for provider should be allowed for current providers, else point to add provider page
  - [x] In Add mode, it should be allowed to copy data from one of current things for all, with some mods in conflict avoiding fields
  - [x] If we try to set a inbuilt preset as default for first time there is a error as of now.

  - [x] overlay needs to have non bool flag support too, need it for default model support, currently no default model change supported for inbuilt ones
  - [x] Add default model in non bool overlay
  - [x] backend should check if atleast one of reasoning or temp is set

- [x] For new month of conversations etc, if we set default provider preset it says read partition error
  - [x] This is ok as it is read error at start and then next dr is created
