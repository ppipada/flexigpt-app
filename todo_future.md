# Project Future notes

## Laundry list

- [ ] Signing for Win
- [ ] Better title deduce
- [ ] Test Windows build
- [ ] Details of request are not visible immediately. We use a trapper in golang ai provider package to get req/resp details, that means until the completion returns there is no way of getting the request details. Some refactoring of workflow is needed to make this happen.
- [ ] Need a stop streaming / cancel button on assistant

## Knowledge base

Implement a `DocStoreSet` that manages multiple `DocStore` instances within a Wails desktop application. Each `DocStore` can use different adapters (e.g., local file system, HTTP) and contains collections of documents. Documents are split into chunks and stored in a vector database as separate vector "documents".

### UI

- [x] Simple single page that lists servers + collections under them.
  - [x] KB server has some associated detail
  - [x] Collection row is still a visible row but placed so that it is visually known that it belongs to the server
- [x] Add/edit/remove buttons as needed
- [x] Modal to edit server details
- [x] Modal to edit collection details and Viewing documents within a collection and number of chunks info.
- [ ] Add/Edit docs inside the collections modal
- [ ] Integrate with backend

### Entity: `DocStoreSet`

- [ ] Top level entity

- [ ] **Create the `DocStoreSet` struct with the following fields:**

  ```go
  type DocStoreSet struct {
    MetadataPath string
    DocStores    map[DocStoreName]*DocStore
  }
  ```

- [ ] **Implement methods for `DocStoreSet`:**

  - Initialize the default local filesystem docstore and add it in the set.
  - Dont implement add for now. ~~`AddDocStore(name string, adapter Adapter)`: Adds a new docstore with the specified adapter.~~
  - `GetDocStore(name string) *DocStore`: Retrieves a docstore by name.
  - Dont implement add for now. ~~`RemoveDocStore(name string)`: Removes a docstore by name.~~
  - Dont implement list for now. ~~`ListDocStores() []string`: Lists all docstore names.~~

### Entity: `DocStore`

- [ ] Design the `DocStore` Interface

  - Initialize can be independent and store specific.

  ```go
  type DocStoreName string

  // DocStore interface for different storage backends
  type DocStore interface {
   GetCollections() (map[CollectionName]*Collection, error)
   AddCollection(name CollectionName) error
   AddDocument(collection string, doc *Document) error
   DeleteDocument(collection string, docID DocumentID) error
   ListDocuments(collection string) ([]*Document, error)
  }
  ```

- [ ] Implement the Default Local File Adapter

  - [ ] Create a `ChromemDocStore` struct implementing the `DocStore` interface.
  - [ ] Define the fields for `ChromemDocStore`
    - [ ] `BasePath` (string): Base file path for storing documents.
  - [ ] Implement the `DocStore` methods for `ChromemDocStore`.

- [ ] Support for HTTP-Based DocStores
  - [ ] Similarly implement a HTTPDocStore
  - [ ] Define the fields for `HTTPDocStore`:
    - `Endpoint` (string): Base URL of the HTTP docstore.
    - `Headers` (map[string]string): Optional headers for HTTP requests.

### Entity: `Collection`

- [ ] **Create the `Collection` struct with the following fields:**

  ```go
  type CollectionName string
  type Collection struct {
    Name      CollectionName
    Documents map[string]*Document
  }
  ```

- [ ] **Implement methods for `Collection`:**
  - `AddDocument(doc *Document) error`: Adds a new document.
  - `GetDocument(docID string) *Document`: Retrieves a document by ID.
  - `RemoveDocument(docID string) error`: Removes a document by ID.
  - `ListDocuments() ([]string, error)`: Lists all document IDs.

### Entity: `Document`

- [ ] Allow CRUD for docs inside a collection
- [ ] ~~May not need a separate page. docs list can be in collection page~~
- [ ] Create searchable chunked entities
- [ ] Need ways to integrate with doc type specific loader/chunker/semantic analyzer.
- [ ] A doc type is going to have specific loader and semantic parsers. E.g: PDF would need to extract text then do some semantic classification and then create chunks to store embeddings.
- [ ] Can provide basic already available loaders in app, but would need a defined hook for it.

- Document

  - Features

    - CRUD operations
    - Store and manage documents
    - Allow document upload and deletion
    - Provide search capabilities

- [ ] **Create the `Document` struct with the following fields:**

  - `ID` (string): Unique identifier for the document.
  - `Content` (string): Full content of the document.
  - `Metadata` (map[string]interface{}): Document metadata.

  ```go
  type Document struct {
    ID       string
    Content  string
    Chunks   []*Chunk
    Metadata map[string]interface{}
  }
  ```

- [ ] **Implement methods for `Document`:**

  - `SplitIntoChunks(chunkSize int) error`: Splits the document content into chunks.
  - `GetChunk(index int) (*Chunk, error)`: Retrieves a specific chunk by index.

- [ ] Define the `Chunk` Structure

  - `Index` (int): The order of the chunk in the document.
  - `Content` (string): The content of the chunk.
  - `VectorID` (string): Identifier in the vector database.

  ```go
  type Chunk struct {
    Index    int
    Content  string
    VectorID string
  }
  ```

- [ ] **Modify `Document.SplitIntoChunks` to:**

  - Split the document content into chunks of specified size.
  - For each chunk:
    - Create a `Chunk` instance.
    - Store the chunk in the vector database using `VectorStore.StoreChunk`.
    - Update the `VectorStore` in the `Chunk`.

### Entity: `VectorStore`

- Use VectoreStore interface from langchaingo and implement for chromamem for local files

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

## Pushed out list

- [ ] ~~logger is imported in securejsondb before it is set as ipc file logger in appimage~~
- [ ] ~~Electron: currently supported only via appimagelauncher. Better do snap I suppose, but explore later.~~
- [ ] ~~Electron: see about package distribution and updates~~
- [ ] Code interpreter for few languages baked in. This can be provided as a normal tool so that
- [ ] Container build and publish
