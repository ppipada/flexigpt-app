# Tools – Functional Specification

## Terminology

| Term          | Definition                                                                                 |
| ------------- | ------------------------------------------------------------------------------------------ |
| Tool          | A callable action. The call may execute local Go code or perform an outbound HTTP request. |
| Tool Type     | `"go"` or `"http"`.                                                                        |
| Tool Bundle   | Named container that groups related tools under a single on/off switch.                    |
| Slug          | Short human-friendly identifier (`+weather`).                                              |
| Version       | Opaque label that distinguishes revisions of the same slug.                                |
| Built-in Tool | Tool compiled into the binary. Content is read-only; only the `isEnabled` flag may change. |

## Objectives

- Let users list, enable/disable and invoke tools from:
  - The "+" invoke option/palette in chat
  - A dedicated "Tools" admin screen (CRUD + live tester)
- Support two execution back-ends:
  a. `go` - zero-latency local call (`func(ctx, jsonArgs) (jsonResp, error)`)
  b. `http` - one or many outbound REST calls with templated URLs
- Organize tools into bundles that can be toggled as a group.
- Store all user-defined data offline (flat-files); online connectivity only required at invocation time for HTTP tools.
- Never mutate stored data automatically if an LLM model or Go function disappears; instead mark as "disabled".

## Data Model and Constraints

- `tool.json`

  ```jsonc
  {
    "bundleID"     : "utilities",
    "slug"         : "weather",
    "version"      : "v2",
    "displayName"  : "Weather report",
    "description"  : "Fetch current weather for a city",
    "type"         : "http",          // "go" or "http"
    "isEnabled"    : true,
    "isBuiltIn"    : false,
    "argSchema"    : { ... },         // JSON-Schema
    "outputSchema" : { ... },         // JSON-Schema
    "impl"         : { ... },
    "createdAt"    : "2024-05-14T13:44:00Z",
    "modifiedAt"   : "2024-06-01T09:17:10Z"
  }
  ```

- Implementation block - Go (`type:"go"`)

  ```jsonc
  { "goFunc": "github.com/acme/agent/tools.Weather" }
  ```

- Implementation block - HTTP (`type:"http"`)

```jsonc
{
  "method": "GET",
  "urlTemplate": "https://api.weatherapi.com/v1/current.json?q=${city}&key=${WEATHER_API_KEY}",
  "headers": { "User-Agent": "MyApp/${version}" },
  "bodyTemplate": "",
  "successCodes": [200],
  "timeoutMs": 10000,
  "responseEncoding": "json", // "json" | "text"
  "extractExpr": "$.current.condition.text", // JSONPath or Regex
  "errorMode": "fail" // "fail" | "empty"
}
```

- Validation rules

  - `urlTemplate` must start with `http://` or `https://`.
  - Placeholders `${var}` come from call arguments, environment variables `${ENV}` or app-scoped secrets.
  - Destination host must match `config.allowedHosts`.

- Field limits

| Field      | Rule                                      |
| ---------- | ----------------------------------------- |
| `slug`     | `^[\p{L}\p{Nd}-]{1,64}$` (case-sensitive) |
| `version`  | same as slug                              |
| Uniqueness | (`bundleID`, `slug`, `version`) tuple     |

## API Surface

All routes are relative to `/tools`.

### Bundles

| Verb   | Path                  | Body                                          | Purpose                  |
| ------ | --------------------- | --------------------------------------------- | ------------------------ |
| PUT    | `/bundles/{bundleID}` | `{slug, displayName, description, isEnabled}` | Create or replace bundle |
| PATCH  | `/bundles/{bundleID}` | `{isEnabled}`                                 | Toggle bundle            |
| DELETE | `/bundles/{bundleID}` | —                                             | Soft-delete bundle       |
| GET    | `/bundles`            | — + filters (`includeDisabled`, `page*`)      | List bundles             |

### 4.2 Tools

| Verb   | Path                                             | Body / Query                                        | Purpose                                |
| ------ | ------------------------------------------------ | --------------------------------------------------- | -------------------------------------- |
| PUT    | `/bundles/{bundleID}/tools/{toolSlug}`           | full tool spec                                      | Create new version                     |
| PATCH  | `/bundles/{bundleID}/tools/{toolSlug}`           | `{version, isEnabled}`                              | Enable / disable a version             |
| DELETE | `/bundles/{bundleID}/tools/{toolSlug}?version=x` | —                                                   | Hard-delete a version                  |
| GET    | `/bundles/{bundleID}/tools/{toolSlug}?version=x` | —                                                   | Fetch specific or active version       |
| GET    | `/tools`                                         | filters (`type, bundleIDs, includeDisabled, page*`) | Global list                            |
| GET    | `/tools/search`                                  | `q, includeDisabled, pageSize, pageToken`           | Fuzzy search                           |
| POST   | `/invoke`                                        | `{bundleID?, slug, args}`                           | Execute active version, returns result |

## Runtime Behavior

- Invocation flow

  - Resolve **active version** for given `(bundle?, slug)`.
  - Validate `args` against `argSchema` (reject with `400` on failure).
  - Dispatch:
    - **Go** → `ctx, argsJSON` passed to registered func.
    - **HTTP** → build request from templates, run with retry/back-off.
  - Parse and validate response against `outputSchema`.
  - Return `{ok:true, value}` or `{ok:false, error}`.
  - Usage metrics (`lastCalledAt`, `callCount`) updated asynchronously.

- Enable/Disable rules

  - Disabling a bundle implicitly disables all contained tools.
  - Built-in tools can only flip `isEnabled`; any other mutation ⇒ `403`.

- Deletion

  - Soft-deleted bundles are hidden immediately; a janitor task purges them if the directory remains empty for >60 min.

## Non-Functional Requirements

- Storage - one JSON file per bundle in `~/.myapp/tools/{bundle}.json`.
- Search - optional SQLite FTS; falls back to in-memory prefix scan when unavailable.
- Concurrency - file-level lock while writing to enforce unique keys across processes.
- Timeouts - default 10 s for HTTP, 5 s for Go; overridable downward but not above 60 s.
- Observability - each call logs `(bundle, slug, duration, status)`; request/response bodies are only logged at DEBUG and redact secrets.

## Pending / Future

- Auth helpers (HMAC, OAuth2 token refresh) for HTTP tools.
- Streaming outputs (Server-Sent Events → chat tokens).
- Import / export bundles as ZIP.
- Rate limiting per tool (`maxCallsPerMinute`).
- UI recipe builder for non-technical users.
- Multi-step tools (sequence or graph of sub-tools).
- Internationalised names & descriptions.
