package ftsengine

import (
	"context"
	"database/sql"
)

const (
	MemoryDBBaseDir   = ":memory:"
	ColNameExternalID = "externalid"
	ColNameRowID      = "rowid"
)

type SearchResult struct {
	// String id stored in the ColNameExternalID column.
	ID string
	// Bm25.
	Score float64
}

// ListResult is returned by BatchList().
type ListResult struct {
	// String id stored in the ColNameExternalID column.
	ID     string
	Values map[string]string
}

// Column declares one FTS5 column.
type Column struct {
	// SQL identifier.
	Name string `json:"name"`
	// Stored but not tokenised.
	Unindexed bool `json:"unindexed"`
	// Bm25 weight (0 is treated as 1).
	Weight float64 `json:"weight"`
}

type Config struct {
	BaseDir    string   `json:"baseDir"`
	DBFileName string   `json:"dbFileName"`
	Table      string   `json:"table"`
	Columns    []Column `json:"columns"`
}

type sqlExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}
