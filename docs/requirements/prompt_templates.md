# Prompt Templates

This document distils all product notes above into an actionable engineering reference for the _Prompt Template_ feature set.

## Objectives

- Allow end-users to save, manage and invoke reusable, variable-driven prompts inside the chat surface.
- Group templates into **Prompt Bundles** that can be enabled/disabled in one switch.
- Expose the feature across:
  - Chat input (`/` palette + auto-complete).
  - PromptTemplates page (CRUD & admin).
- Keep the system offline-friendly and “loose-coupled”: deleting a tool, data-source or model _never_ modifies stored templates.

## Lifecycle Requirements

- Create, read, update, delete Prompt Bundle

- Create, read, delete Prompt Template

  - Version templates
  - Validation:
    - JSON schema + business rules
    - Slug uniqueness within its bundle is mandatory at save-time
    - Keep versioning, but treat slug + highest version as the “active” template.
    - On duplicate‐slug PUT: accept only if version is strictly higher (semver monotonic, else Precondition Failed).
  - Allow only duplicate and create new higher version for a prompt template
  - If a slug has multiple versions in same bundle, only pick latest version
  - API has a provision for fetching all versions of a template
  - Hard-delete only removes local copy; existing chat messages keep a snapshot string of the rendered prompt

- Enable / disable a single template or an entire bundle

  - Have immediate effect on search & invocation

- Search & Ranking
  - Maintain client-side full-text index (async refresh on write)
  - Apply weighted ranking rules (prefix > whole word > fuzzy > recency/popularity)
  - Disabled rows are hidden from default results

## Workflows

### Chat Invocation

- Typing `/` in chat opens palette

  - Palette is populated via full-text search

- User can invoke by `/slug` or `/bundleAlias.slug`

  - Parser must support real-time suggestions with fuzzy matching

- If multiple matches, show ranked list; **Enter** activates highlighted row

- Variable Substitution Pipeline

  - Collect defaults
  - `${varName}` marker supports simple replacement only (no conditionals, loops in v1).
  - Undefined, non-required variables fallback to `""`.
  - Run Pre-Processors (may mutate variables)
  - Prompt blocks interpolation
  - Send to LLM

- Pre-Processor error-handling modes: `empty` (insert empty string) or `fail` (abort invocation and notify user)

- If a referenced tool/data-source is missing at runtime, treat as Pre-Processor error

### Deleted Template in Chat

- If old chat links to a disabled template chat renders warning and shows text with disabled modifications; user may request re-enable to continue chat

## Example API Surface

```shell
# Create bundle, no templates
PUT    /prompts/bundles/{bundleID}
# Delete bundle, only if empty. detect via scan. do a soft and deferred delete to avoid membership issues.
DELETE /prompts/bundles/{bundleID}
# Path a bundle: Only Enable/Disable support
PATCH  /prompts/bundles/{bundleID}     {isEnabled:bool}
# List bundles. Paginated
GET /prompts/bundles?bundleIDs=i1,i2&includeDisabled=true&pageSize=<n>&pageToken=abc

# Create/Replace a template
PUT    /prompts/bundles/{bundleID}/templates/{templateID}
# Delete a template
DELETE /prompts/bundles/{bundleID}/templates/{templateID}?version=<version>
# Path a template: Enable/Disable support only
PATCH  /prompts/bundles/{bundleID}/templates/{templateID}   {version:vX.y.z, isEnabled:bool}
# Get a template
GET    /prompts/bundles/{bundleID}/templates/{templateID}?version=<version>
# List templates. Paginated
GET    /prompts/templates?tags=x,y,z&includeDisabled=<bool>&bundleIDs=a,b,c&allVersions=<bool>&pageSize=<n>&pageToken=<str>
# Search templates
GET    /prompts/templates/search?q=<str>&includeDisabled=<bool>&pageToken=<str>&pageSize=<n>
```

## Internal Notes

- Storage

  - Store path: `$APP_DATA/prompttemplates/<bundleid>/<template>.json`.

- Ranked FTS support using reproducible sqlite. Sample weight factors (bigger = higher row):

  | Weight | Condition                                        |
  | -----: | ------------------------------------------------ |
  |   +100 | slug **prefix** match                            |
  |    +60 | full-word slug match                             |
  |    +40 | displayName prefix                               |
  |    +20 | fuzzy hit in name/description                    |
  |    +15 | _recently used by me_                            |
  |    +10 | _popular_ in workspace                           |
  |     +5 | belongs to **enabled bundle**                    |
  |    −20 | template `isEnabled==false` (but bundle enabled) |
  |     −∞ | `effectiveEnabled==false` (drop)                 |

- Invocation Flow (pseudocode)

  ```go
  tpl := ResolveTemplate(userInput)
  vars := CollectVars(tpl, userContext)
  vars = RunPreProcessors(tpl.PreProcessors, vars)   // handle OnError
  finalPrompt := Render(tpl.Blocks, vars)
  SendToLLM(finalPrompt, selectedModel)
  ```

## Open Questions / Future Work

- Template import/export (JSON, Bundle marketplace, etc).
- Variable UI wizard for multi-step filling.
- Pre-Processor graph execution (dependencies between calls).
- Internationalization.
