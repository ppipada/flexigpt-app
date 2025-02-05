package docstore

// import (
// 	"errors"

// 	"github.com/flexigpt/flexiui/pkg/docstore/spec"
// )

// // HTTPDocStore implements the DocStoreAdapter interface for HTTP-based storage
// type HTTPDocStore struct {
// 	Name        spec.DocStoreName
// 	Metadata    map[string]interface{}
// 	Collections map[spec.CollectionName]*spec.Collection
// 	Endpoint    string
// 	Headers     map[string]string
// }

// // HTTPDocStore initializes a new DocStore
// func NewHTTPDocStore(
// 	name spec.DocStoreName,
// 	endpoint string,
// 	headers map[string]string,
// ) *HTTPDocStore {
// 	return &HTTPDocStore{
// 		Name:        name,
// 		Metadata:    make(map[string]interface{}),
// 		Collections: make(map[spec.CollectionName]*spec.Collection),
// 		Endpoint:    endpoint,
// 		Headers:     headers,
// 	}
// }

// // GetCollection retrieves a collection by name
// func (hds *HTTPDocStore) GetCollections() (map[spec.CollectionName]*spec.Collection, error) {
// 	return hds.Collections, nil
// }

// // AddCollection adds a new collection
// func (hds *HTTPDocStore) AddCollection(name spec.CollectionName) error {
// 	if _, exists := hds.Collections[name]; exists {
// 		return errors.New("collection already exists")
// 	}
// 	hds.Collections[name] = &spec.Collection{
// 		Name:      name,
// 		Documents: make(map[spec.DocumentID]*spec.Document),
// 	}
// 	return nil
// }

// func (hds *HTTPDocStore) DeleteDocument(collection string, docID spec.DocumentID) error {
// 	// Implementation for local file system
// 	return nil
// }

// func (hds *HTTPDocStore) AddDocument(collection string, doc *spec.Document) error {
// 	// Implementation for local file system
// 	return nil
// }

// func (hds *HTTPDocStore) ListDocuments(collection string) ([]*spec.Document, error) {
// 	// Implementation for local file system
// 	return nil, nil
// }
