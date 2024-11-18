package spec

import "context"

type EmbeddingFuncID string

type DocumentID string

type Document struct {
	ID        DocumentID
	Metadata  map[string]string
	Embedding []float32
	Content   string
}

// DocumentQueryResult represents a single result from a query.
type DocumentQueryResult struct {
	ID        DocumentID
	Metadata  map[string]string
	Embedding []float32
	Content   string

	// The cosine similarity between the query and the document.
	// The higher the value, the more similar the document is to the query.
	// The value is in the range [-1, 1].
	Similarity float32
}

type DocumentCollectionID string

// DocumentCollection represents a collection of documents
type DocumentCollection struct {
	ID          DocumentCollectionID
	Name        string
	Documents   map[DocumentID]*Document
	Metadata    map[string]string
	EmbeddingID EmbeddingFuncID
	BasePath    string
	Compress    bool
}

type IDocumentCollection interface {
	AddDocuments(
		ctx context.Context,
		collectionID DocumentCollectionID,
		documents []Document,
		concurrency int,
	) error
	GetDocumentByID(
		ctx context.Context,
		collectionID DocumentCollectionID,
		id DocumentID,
	) (Document, error)
	DeleteDocumentByID(ctx context.Context, collectionID DocumentCollectionID, id DocumentID) error
	ListDocuments(ctx context.Context, collectionID DocumentCollectionID) ([]*Document, error)

	Query(
		ctx context.Context,
		collectionID DocumentCollectionID,
		queryText string,
		nResults int,
		where, whereDocument map[string]string,
	) ([]DocumentQueryResult, error)
	QueryEmbedding(
		ctx context.Context,
		collectionID DocumentCollectionID,
		queryEmbedding []float32,
		nResults int,
		where, whereDocument map[string]string,
	) ([]DocumentQueryResult, error)
}

type DocumentDBID string

// IDocumentDB interface for different document storage backends
type IDocumentDB interface {
	CreateCollection(
		ctx context.Context,
		name string,
		metadata map[string]string,
		embeddingFunc EmbeddingFuncID,
		apiKey string,
	) (*DocumentCollection, error)
	DeleteCollection(ctx context.Context, id DocumentCollectionID) error
	GetCollection(ctx context.Context, id DocumentCollectionID) (*DocumentCollection, error)
	ListCollections(ctx context.Context) (map[DocumentCollectionID]*DocumentCollection, error)
}
