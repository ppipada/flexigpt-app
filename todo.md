# Project Implementation Plan

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
    - [ ] currently supported only via appimagelauncher. Better do snap I suppose, but explore later.
  - [ ] see about package distribution and updates

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
  - [ ] connect this with chat page

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

- Agent screen

  - [ ] See if a jupyter notebook UI can be presented
