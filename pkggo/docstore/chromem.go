package docstore

import (
	"context"
	"errors"

	"github.com/flexigpt/flexiui/pkggo/docstore/spec"
	"github.com/google/uuid"
	"github.com/philippgille/chromem-go"
)

var embeddingModelMapOpenAI = map[spec.EmbeddingFuncID]chromem.EmbeddingModelOpenAI{
	spec.EmbeddingModelOpenAI3Large: chromem.EmbeddingModelOpenAI3Large,
	spec.EmbeddingModelOpenAI3Small: chromem.EmbeddingModelOpenAI3Small,
}

var embeddingModelMapCohere = map[spec.EmbeddingFuncID]chromem.EmbeddingModelCohere{
	spec.EmbeddingModelCohereMultilingualV2:      chromem.EmbeddingModelCohereMultilingualV2,
	spec.EmbeddingModelCohereEnglishLightV2:      chromem.EmbeddingModelCohereEnglishLightV2,
	spec.EmbeddingModelCohereEnglishV2:           chromem.EmbeddingModelCohereEnglishV2,
	spec.EmbeddingModelCohereMultilingualLightV3: chromem.EmbeddingModelCohereMultilingualLightV3,
	spec.EmbeddingModelCohereEnglishLightV3:      chromem.EmbeddingModelCohereEnglishLightV3,
	spec.EmbeddingModelCohereMultilingualV3:      chromem.EmbeddingModelCohereMultilingualV3,
	spec.EmbeddingModelCohereEnglishV3:           chromem.EmbeddingModelCohereEnglishV3,
}

func getEmbeddingFunc(
	_ context.Context,
	funcID spec.EmbeddingFuncID,
	apiKey string,
) (chromem.EmbeddingFunc, error) {
	// These are from OpenAI platform
	// If other providers like Azure or OLlama are to be used which provide OpenAI compatible API, declareNewIDs and use
	omodel, exists := embeddingModelMapOpenAI[funcID]
	if exists {
		return chromem.NewEmbeddingFuncOpenAI(apiKey, omodel), nil
	}

	cmodel, exists := embeddingModelMapCohere[funcID]
	if exists {
		return chromem.NewEmbeddingFuncCohere(apiKey, cmodel), nil
	}

	return nil, errors.New("unsupported embedding function ID")
}

type ChromemDocumentDB struct {
	id          spec.DocumentDBID
	name        string
	metadata    map[string]string
	collections map[spec.DocumentCollectionID]*spec.DocumentCollection
	basePath    string
	compress    bool
	chromemDB   *chromem.DB
}

// Option is a function that configures a ChromemDocumentDB
type Option func(*ChromemDocumentDB) error

// WithName sets the name of the ChromemDocumentDB
func WithName(name string) Option {
	return func(db *ChromemDocumentDB) error {
		db.name = name
		return nil
	}
}

// WithMetadata sets the metadata of the ChromemDocumentDB
func WithMetadata(metadata map[string]string) Option {
	return func(db *ChromemDocumentDB) error {
		db.metadata = metadata
		return nil
	}
}

// WithBasePath sets the base path of the ChromemDocumentDB
func WithBasePath(basePath string) Option {
	return func(db *ChromemDocumentDB) error {
		db.basePath = basePath
		return nil
	}
}

// WithCompression sets the compression flag of the ChromemDocumentDB
func WithCompression(compress bool) Option {
	return func(db *ChromemDocumentDB) error {
		db.compress = compress
		return nil
	}
}

// NewChromemDocumentDB initializes a new ChromemDocumentDB with the given options
func NewChromemDocumentDB(options ...Option) (*ChromemDocumentDB, error) {
	db := &ChromemDocumentDB{
		metadata:    make(map[string]string),
		collections: make(map[spec.DocumentCollectionID]*spec.DocumentCollection),
	}

	for _, option := range options {
		if err := option(db); err != nil {
			return nil, err
		}
	}

	// Initialize the underlying chromem DB
	chromemDB, err := chromem.NewPersistentDB(db.basePath, db.compress)
	if err != nil {
		return nil, err
	}
	db.chromemDB = chromemDB

	// Generate a unique ID for the document DB
	u, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	db.id = spec.DocumentDBID(u.String())

	return db, nil
}

// CreateCollection creates a new document collection
func (db *ChromemDocumentDB) CreateCollection(
	ctx context.Context,
	name string,
	metadata map[string]string,
	embeddingFuncID spec.EmbeddingFuncID,
	apiKey string,
) (*spec.DocumentCollection, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	// Add the name to the metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["name"] = name

	efunc, err := getEmbeddingFunc(ctx, embeddingFuncID, apiKey)
	if err != nil {
		return nil, err
	}

	// Use the generated ID as the name in chromem
	collectionID := spec.DocumentCollectionID(u.String())
	_, err = db.chromemDB.CreateCollection(string(collectionID), metadata, efunc)
	if err != nil {
		return nil, err
	}

	collection := &spec.DocumentCollection{
		ID:          collectionID,
		Name:        name,
		Documents:   make(map[spec.DocumentID]*spec.Document),
		Metadata:    metadata,
		EmbeddingID: embeddingFuncID,
		BasePath:    db.basePath,
		Compress:    db.compress,
	}
	db.collections[collectionID] = collection
	return collection, nil
}

// DeleteCollection deletes a document collection
func (db *ChromemDocumentDB) DeleteCollection(
	ctx context.Context,
	id spec.DocumentCollectionID,
) error {
	_, exists := db.collections[id]
	if !exists {
		return errors.New("collection not found")
	}

	// Use the collection ID as the name in chromem
	err := db.chromemDB.DeleteCollection(string(id))
	if err != nil {
		return err
	}

	delete(db.collections, id)
	return nil
}

// GetCollection retrieves a document collection by ID
func (db *ChromemDocumentDB) GetCollection(
	ctx context.Context,
	id spec.DocumentCollectionID,
) (*spec.DocumentCollection, error) {
	collection, exists := db.collections[id]
	if !exists {
		return nil, errors.New("collection not found")
	}
	return collection, nil
}

// ListCollections lists all document collections
func (db *ChromemDocumentDB) ListCollections(
	ctx context.Context,
) (map[spec.DocumentCollectionID]*spec.DocumentCollection, error) {
	return db.collections, nil
}
