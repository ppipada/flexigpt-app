# Project Current Milestones

## Ongoing Scope

- Milestone 0 completion
  - Top level feature: Chat, Conversation history, Settings
  - Local installer for Linux
- Milestone 1 partial
  - KB, Tools, Prompts

## Laundry list

- MCP handling

  - [ ] Implement MCP protocol
  - [ ] Add servers for: file, git, simplemapstore

- [ ] Need a settings data version and migration func

  - [ ] API key should never reach frontend after save

- [ ] There is large cpu consumption sometime, debug
  - [ ] It is mainly seen if both wails dev and current app is open and being used wails://wails consumes 100%
  - [ ] Also check when one request was fired but not allowed to complete for some reason
  - [ ] Also, if the message thread gets very very big there seems to be a issue.

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
