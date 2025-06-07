package ftsengine

type SearchResult struct {
	// string id stored in the "externalID" column
	ID string
	// bm25
	Score float64
}

// Column declares one FTS5 column.
type Column struct {
	// SQL identifier
	Name string `json:"name"`
	// stored but not tokenised
	Unindexed bool `json:"unindexed"`
	// bm25 weight (0 → treated as 1)
	Weight float64 `json:"weight"`
}

type Config struct {
	DBPath  string   `json:"dbPath"`
	Table   string   `json:"table"`
	Columns []Column `json:"columns"`
}
