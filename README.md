# FlexiGPT

FlexiGPT is a _cross-platform desktop client_ that lets you chat with multiple Large Language Models (LLMs) from a single, streamlined interface.
Launch FlexiGPT, open Settings -> Providers, paste your API key, start chatting.

> ⚠ Early Access Notice:
>
> The project is under active development; expect breaking changes and incomplete features between releases.

- [Quick Start](#quick-start)
- [Key Features at a Glance](#key-features-at-a-glance)
- [Install](#install)
  - [MacOS](#macos)
  - [Windows](#windows)
  - [Linux](#linux)
- [Feature: Conversation History](#feature-conversation-history)
- [Feature: Model Presets](#feature-model-presets)
- [Feature: Prompt Templates](#feature-prompt-templates)
- [Built With](#built-with)
- [Contributing](#contributing)

## Quick Start

- Install FlexiGPT [for your OS](#install).
- Launch and Open **Settings**.
- In **Auth Keys** -> Paste the API key for your provider.
- Start chatting!

## Key Features at a Glance

- **Multi-provider Connectivity**

  - First-class connectors for OpenAI, Anthropic, Google, DeepSeek, xAI, Huggingface, Openrouter, Local LLamaCPP, **plus any local/remote OpenAI or Anthropic compatible endpoint**.
  - API keys are stored securely in your OS keyring - never in plain text.

- **Unified Chat Workspace**

  - Switch models mid-conversation, chain results from one model into the next, and fine-tune generation parameters on the fly.
  - Productivity tooling inside the chat
    - Code snippets with copy/export
    - Mermaid diagrams with copy/export/zoom
    - Full Math/LaTeX rendering
    - Readily available chat API requests and response details

- **Persistent Conversation History**

  - Auto-saved sessions with full-text search, resumable at any time.

- **Model Presets**

  - One-click loading of curated or user-defined parameter sets (model, temperature, max-tokens, system prompt, etc.).
  - Tweak values per conversation without altering the saved preset.

- **Slash-Command Prompt Templates**

  - Turn complex prompts into reusable `/<shortcuts>` with variables and pre-processing tools
  - If required, group into toggle-able bundles.

- **System/Light/Dark Themes** supported with auto-detect & manual toggle, with an option to choose any of the [**DaisyUI themes**](https://daisyui.com/docs/themes/?lang=en#list-of-themes).

## Install

### MacOS

- Download the `.pkg` release package.
- Click to install the `.pkg`. It will walk you through the installation process.
- Local data (settings, conversations, logs) is stored at:
  - `~/Library/Containers/io.github.flexigpt.client/Data/Library/Application\ Support/flexigpt/`

### Windows

- Download the `.exe` release package.
- Click to install the `.exe`. It will walk you through the installation process.
- Note: Windows builds have undergone very limited testing.

### Linux

- Download the `.flatpak` release package.
- If Flatpak is not installed, enable it for your distribution

  - Ubuntu/Debian/etc (APT based systems):

    ```shell
    sudo apt update # update packages
    sudo apt install -y flatpak # install flatpak
    sudo apt install -y gnome-software-plugin-flatpak # optional, enables flathub packages in gnome sofware center
    flatpak remote-add --if-not-exists flathub https://dl.flathub.org/repo/flathub.flatpakrepo
    ```

  - Some additional helper commands can be found in this [script](./scripts/initialize_flatpak.sh)

- Install the package

  - `flatpak install --user FlexiGPT-xyz.flatpak`
  - `flatpak info io.github.flexigpt.client`

- Running the app

  - Using launcher GUI: You can launch the app from your distributions's launcher. E.g: In Ubuntu: Press the window key, type flexigpt and click on icon.
  - Using terminal: `flatpak run io.github.flexigpt.client`
  - If you use Nvidia and its proprietary drivers, you _may_ see that the run command open a blank screen and close, the workaround for it is to run the app as:
    - `flatpak run --env=WEBKIT_DISABLE_COMPOSITING_MODE=1 io.github.flexigpt.client`
    - Open bugs in upstream projects that cause this: [Webkit issue 180739](https://bugs.webkit.org/show_bug.cgi?id=180739), [Webkit issue 262607](https://bugs.webkit.org/show_bug.cgi?id=262607), [Debian issue 1082139](https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=1082139), [Wails issue 2977](https://github.com/wailsapp/wails/issues/2977)

- Your local data (settings, conversations, logs) will be at:
  - `~/.var/app/io.github.flexigpt.client/data/flexigpt`

## Feature: Conversation History

- Overview

  - Every chat session is persisted as a _Conversation_ containing its title, timestamps and full message sequence.
  - Conversations are stored locally; you can reload a conversation and continue from last point at any time.
  - A full-text search bar provides instant retrieval across titles and message contents.

- Key Concepts

  | Term              | Description                                                                      |
  | ----------------- | -------------------------------------------------------------------------------- |
  | Conversation      | A saved chat session composed of messages and metadata.                          |
  | Conversation Item | Lightweight record (ID, title, creation time) shown in lists and search results. |
  | Message Roles     | `system`, `user`, `assistant`; each recorded with timestamp and details.         |

- Using Conversation History

  - Click the _Search Bar_ to open the history dropdown.
  - Recent conversations are listed chronologically; type in the search bar to filter by keywords appearing in either titles or message text.
  - Select a row to reload the full conversation and continue chatting.

- Automatic Saving

  - A new conversation record is started automatically when you open a fresh chat window.
  - The title is generated heuristically from the first user message but can be renamed at any time.
  - Messages are appended in real-time; timestamps are stored for precise ordering.

## Feature: Model Presets

- Overview

  - Model Presets bundle a model selection (e.g. `gpt-4.1`) together with recommended defaults such as temperature, token limits, system prompt and reasoning parameters.
  - Presets are organized per provider. A provider may expose many presets and may declare one of them as its default.
  - A preset can be enabled or disabled at any time; existing chat history is not modified.
  - Per-conversation overrides are possible: after selecting a preset the user may fine-tune individual parameters (temperature, reasoning, etc.) without changing the stored preset.
  - Supports models with multiple behaviors: non-reasoning, reasoning with levels (high/medium/low), hybrid reasoning (with reasoning tokens).

- Key Concepts

  | Term                 | Description                                                                                                    |
  | -------------------- | -------------------------------------------------------------------------------------------------------------- |
  | Model Preset         | Named set of defaults for a single model (temperature, max tokens, reasoning, etc.).                           |
  | Provider Preset      | Collection of model presets that belong to one provider and defines that provider’s default preset.            |
  | Default Model Preset | The preset automatically applied when the provider is selected and the user has not chosen a different preset. |

- Selecting a Preset in Chat

  - In the chat input bar, open the dropdown and choose any preset listed.
  - The preset’s parameters appear next to the dropdown; use the slider or number fields to make conversation-specific adjustments if desired.
  - The adjusted parameters affect only the current conversation and are discarded when the session ends.

- Managing Presets

  - Click the _Model Presets_ icon in the sidebar.
  - Select a provider or create a new one.
  - The app provides inbuilt presets. These will be listed.
  - Create, edit, delete, enable/disable presets as required.
  - Mark one preset as _Default_ for the provider; it loads automatically when a model from that provider is first chosen.

- Notes

  - Deleting a preset does not alter existing conversations; those sessions keep a snapshot of the parameters that were active when the message was sent.
  - Global defaults are applied for any parameter left blank in the preset.

## Feature: Prompt Templates

- Overview

  - Prompt Templates allow you to store complex prompts (including variables and pre-processing tools) and reuse them with a single slash-command.
  - Templates can be organized into Bundles, which act as configurable packs that can be enabled or disabled.
  - Templates and Bundles can be disabled or removed at any time without affecting existing chat history.

- Key Concepts

  | Term                 | Description                                                                                    |
  | -------------------- | ---------------------------------------------------------------------------------------------- |
  | Prompt Template      | A reusable prompt that can contain variables and pre-processor tools.                          |
  | Slash Command (Slug) | The shortcut a user types in chat (e.g. `/release-notes`). A slug is unique within its bundle. |
  | Prompt Bundle        | A collection of templates that can be toggled on or off as a unit and shared with other users. |

- Invoking a Template

  - In the chat input, type `/` followed by at least one character.
    - Optionally type `/<bundleShortCode>.<slug>` (e.g. `/prod.price`) to target a specific bundle.
  - An auto-complete list appears, ranked by relevance.
  - Use `↑/↓` to navigate if necessary, then press **Enter**.
  - If two bundles provide the same slug, both options are listed.
  - The most relevant match is selected for processing.

- Managing Prompt Templates

  - Click the _Prompts_ icon in the sidebar.
  - Select a current prompt bundle or create a new one.
  - The app provides inbuilt bundles and templates. These will be listed.
  - Create, edit, delete, enable/disable bundles/templates as required.

- Notes

  - If a template references a tool that was later removed, the template remains intact. If the tool is unavailable at runtime the corresponding step returns an error or is skipped dependent on usage config.

## Built With

- [Go](https://go.dev/) backend.
- [Wails](https://wails.io/) as a desktop application building platform.
- [Vite](https://vite.dev/) + [React Router v7](https://reactrouter.com/) frontend.
- [DaisyUI](https://daisyui.com/) with [TailwindCSS](https://tailwindcss.com/) for styling.

- Tooling:

  - [Gopls](https://pkg.go.dev/golang.org/x/tools/gopls) and [Modernize](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize)
  - [Golangci-lint](https://golangci-lint.run/)
  - [Knip](https://knip.dev/)
  - [ESLint](https://eslint.org/)
  - [Prettier](https://prettier.io/)
  - [Github actions](https://github.com/features/actions)

- Data storage: `JSON` and `SQLite` files in local filesystem.

## Contributing

- The dev setup docs are located at [devsetup.md](./docs/contributing/devsetup.md)
