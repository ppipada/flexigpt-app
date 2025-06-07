package ftsengine

type SearchResult struct {
	// String id stored in the "externalID" column.
	ID string
	// Bm25.
	Score float64
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
	DBPath  string   `json:"dbPath"`
	Table   string   `json:"table"`
	Columns []Column `json:"columns"`
}
