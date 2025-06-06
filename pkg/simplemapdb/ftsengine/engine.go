package ftsengine

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	_ "github.com/glebarez/go-sqlite" // pure-Go SQLite/FTS5 driver
)

/* ------------------------------------------------------------------ */
/*  public configuration                                              */
/* ------------------------------------------------------------------ */

// Column declares one FTS5 column.
type Column struct {
	Name      string  `json:"name"`      // SQL identifier
	Unindexed bool    `json:"unindexed"` // stored but not tokenised
	Weight    float64 `json:"weight"`    // bm25 weight (0 → treated as 1)
}

type Config struct {
	DBPath  string   `json:"dbPath"`
	Table   string   `json:"table"`
	Columns []Column `json:"columns"`
}

// Result is returned by Search().
type Result struct {
	ID    string  // rowid (=docID)
	Score float64 // bm25
}

/* ------------------------------------------------------------------ */
/*  Engine                                                            */
/* ------------------------------------------------------------------ */

type Engine struct {
	db  *sql.DB
	cfg Config
	hsh string     // schema checksum
	mu  sync.Mutex // serialises write-queries
}

/* ------------------------------------------------------------------ */
/*  construction                                                      */
/* ------------------------------------------------------------------ */

func New(cfg Config) (*Engine, error) {
	if len(cfg.Columns) == 0 {
		return nil, errors.New("ftsengine: need ≥1 column")
	}
	if cfg.DBPath == "" {
		return nil, errors.New("ftsengine: DBPath missing")
	}
	if dir := filepath.Dir(cfg.DBPath); dir != "" && dir != ":memory:" {
		if err := os.MkdirAll(dir, 0o770); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", cfg.DBPath+
		"?busy_timeout=5000&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}

	e := &Engine{db: db, cfg: cfg}
	e.hsh = e.schemaChecksum()

	if err := e.bootstrap(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return e, nil
}

/* ------------------------------------------------------------------ */
/*  internal helpers                                                  */
/* ------------------------------------------------------------------ */

func (e *Engine) schemaChecksum() string {
	h := sha256.New()
	cfg := e.cfg
	_ = json.NewEncoder(h).Encode(cfg)
	return hex.EncodeToString(h.Sum(nil))
}

func (e *Engine) bootstrap(ctx context.Context) error {
	const sqlCreateMetaTable = `CREATE TABLE IF NOT EXISTS meta(k TEXT PRIMARY KEY,v TEXT);`
	const sqlSelectMetaHash = `SELECT v FROM meta WHERE k='h'`
	const sqlInsertMetaHash = `INSERT OR REPLACE INTO meta(k,v) VALUES('h',?)`
	const sqlCreateVirtualTable = `CREATE VIRTUAL TABLE IF NOT EXISTS %s
		USING fts5 (%s,
			tokenize='unicode61 remove_diacritics 1');`
	const sqlDeleteAllRows = `DELETE FROM %s`

	// meta for schema hash
	if _, err := e.db.ExecContext(ctx, sqlCreateMetaTable); err != nil {
		return err
	}

	// existing hash?
	var stored string
	_ = e.db.QueryRowContext(ctx, sqlSelectMetaHash).Scan(&stored)

	// create / replace FTS virtual table
	if stored != e.hsh {
		var cols []string
		cols = append(cols, `path UNINDEXED`)
		for _, c := range e.cfg.Columns {
			col := quote(c.Name)
			if c.Unindexed {
				col += " UNINDEXED"
			}
			cols = append(cols, col)
		}
		ddl := fmt.Sprintf(sqlCreateVirtualTable,
			quote(e.cfg.Table), strings.Join(cols, ","))

		if _, err := e.db.ExecContext(ctx, ddl); err != nil {
			return err
		}
		_, _ = e.db.ExecContext(ctx, sqlInsertMetaHash, e.hsh)

		// schema changed → clear previous rows
		if stored != "" {
			_, _ = e.db.ExecContext(ctx, fmt.Sprintf(sqlDeleteAllRows, quote(e.cfg.Table)))
		}
	}
	return nil
}

func (e *Engine) IsEmpty() bool {
	const sqlIsEmpty = `SELECT count(*) FROM %s` // Table name must be injected safely

	var n int
	_ = e.db.QueryRow(fmt.Sprintf(sqlIsEmpty, quote(e.cfg.Table))).Scan(&n)

	return n == 0
}

// Upsert --------------------------------------------------------------------
// rowid = fnv64(id);  path column keeps the *original* string id.
func (e *Engine) Upsert(id string, vals map[string]string) error {
	if id == "" {
		return errors.New("ftsengine: empty id")
	}
	rowid := fnv64a(id)

	numCols := 1 + len(e.cfg.Columns) // "path" + each column

	colNames := make([]string, 0, numCols)
	marks := make([]string, 0, numCols)
	args := make([]any, 1, numCols+1)
	args[0] = rowid

	colNames = append(colNames, "path")
	marks = append(marks, "?")
	args = append(args, id)

	for _, c := range e.cfg.Columns {
		colNames = append(colNames, quote(c.Name))
		marks = append(marks, "?")
		args = append(args, vals[c.Name])
	}

	const sqlInsertOrReplace = `INSERT OR REPLACE INTO %s (rowid,%s) VALUES (? , %s);`

	sqlQ := fmt.Sprintf(sqlInsertOrReplace,
		quote(e.cfg.Table), strings.Join(colNames, ","), strings.Join(marks, ","))

	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.db.Exec(sqlQ, args...)
	return err
}

func (e *Engine) Delete(id string) error {
	const sqlDeleteByRowID = `DELETE FROM %s WHERE rowid=?`

	_, err := e.db.Exec(fmt.Sprintf(sqlDeleteByRowID, quote(e.cfg.Table)), fnv64a(id))

	return err
}

// Search returns one page of results and, if more results exist,
// an opaque token for the next page.
func (e *Engine) Search(
	ctx context.Context,
	query string,
	pageToken string,
	pageSize int,
) (hits []Result, nextToken string, err error) {
	if pageSize <= 0 || pageSize > 10000 {
		pageSize = 10
	}

	// decode / reset token ---------------------------------------------
	var offset int
	if pageToken != "" {
		var t struct {
			Query  string `json:"q"`
			Offset int    `json:"o"`
		}
		b, err := base64.StdEncoding.DecodeString(pageToken)
		if err == nil {
			_ = json.Unmarshal(b, &t)
		}
		if t.Query == query { // token belongs to same query
			offset = t.Offset
		}
	}

	// bm25 weight parameters (one per column) --------------------------
	var weights []any
	for _, c := range e.cfg.Columns {
		if c.Weight == 0 {
			weights = append(weights, float64(1))
		} else {
			weights = append(weights, c.Weight)
		}
	}

	const sqlSearch = `SELECT path, bm25(%s%s) AS s
			FROM %s WHERE %s MATCH ?
			ORDER BY s, rowid
			LIMIT ? OFFSET ?;` // Table/columns injected safely

	sqlQ := fmt.Sprintf(sqlSearch,
		quote(e.cfg.Table), paramPlaceholders(len(weights)),
		quote(e.cfg.Table), quote(e.cfg.Table))

	args := slices.Clone(weights)
	args = append(args, query, pageSize, offset)

	rows, err := e.db.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	for rows.Next() {
		var r Result
		if err := rows.Scan(&r.ID, &r.Score); err != nil {
			return nil, "", err
		}
		hits = append(hits, r)
	}

	// build next token --------------------------------------------------
	if len(hits) == pageSize {
		offset += pageSize
		buf, _ := json.Marshal(struct {
			Query  string `json:"q"`
			Offset int    `json:"o"`
		}{query, offset})
		nextToken = base64.StdEncoding.EncodeToString(buf)
	}
	return hits, nextToken, rows.Err()
}

/* ------------------------------------------------------------------ */
/*  helpers                                                           */
/* ------------------------------------------------------------------ */

func quote(id string) string { return `"` + strings.ReplaceAll(id, `"`, `""`) + `"` }

func paramPlaceholders(n int) string {
	sb := strings.Builder{}
	for range n {
		sb.WriteString(",?")
	}
	return sb.String()
}

func fnv64a(s string) int64 {
	const (
		offset uint64 = 14695981039346656037
		prime  uint64 = 1099511628211
	)
	h := offset
	for i := range len(s) {
		h ^= uint64(s[i])
		h *= prime
	}
	h &= 0x7fffffffffffffff // mask to 63 bits
	if h > math.MaxInt64 {
		// Should never happen, but just in case
		return math.MaxInt64
	}
	return int64(h)
}
