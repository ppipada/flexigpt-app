package docstore

import (
	"errors"

	"github.com/flexigpt/flexigpt-app/internal/docstore/spec"
)

// DocumentDBSet manages multiple DocStore instances.
type DocumentDBSet struct {
	basePath string
	docDBs   map[spec.DocumentDBID]spec.IDocumentDB
}

func InitDocumentDBSet(dds *DocumentDBSet, basePath string) error {
	if dds == nil || basePath == "" {
		return errors.New("invalid input arguments to InitDocumentDBSet")
	}
	dds.basePath = basePath
	dds.docDBs = make(map[spec.DocumentDBID]spec.IDocumentDB)

	// Initialize default local filesystem docstore.
	cdb, err := NewChromemDocumentDB(
		WithName(spec.ChromemDocStoreName),
		WithMetadata(map[string]string{}),
		WithBasePath(spec.ChromemDocStorePath),
		WithCompression(true),
	)
	if err != nil {
		return err
	}

	dds.docDBs[spec.ChromemDocStoreName] = cdb
	return nil
}
