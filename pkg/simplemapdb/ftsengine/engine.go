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
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	_ "github.com/glebarez/go-sqlite" // pure-Go SQLite/FTS5 driver
)

// Result is returned by Search().
type Result struct {
	ID    string  // string id stored in the "path" column
	Score float64 // bm25
}

type Engine struct {
	db  *sql.DB
	cfg Config
	hsh string     // schema checksum
	mu  sync.Mutex // serialises write-queries
}

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

func (c *Config) Validate() error {
	if len(c.Columns) == 0 {
		return errors.New("ftsengine: need ≥1 column")
	}
	if c.DBPath == "" {
		return errors.New("ftsengine: DBPath missing")
	}

	if strings.TrimSpace(c.Table) == "" {
		return errors.New("ftsengine: empty table name")
	}
	seen := make(map[string]struct{})
	for _, col := range c.Columns {
		if strings.TrimSpace(col.Name) == "" {
			return errors.New("ftsengine: column with empty name")
		}
		if _, dup := seen[col.Name]; dup {
			return fmt.Errorf("ftsengine: duplicate column %q", col.Name)
		}
		seen[col.Name] = struct{}{}
	}
	return nil
}

func New(cfg Config) (*Engine, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
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
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

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
	_ = json.NewEncoder(h).Encode(e.cfg)
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

func (e *Engine) IsEmpty(ctx context.Context) (bool, error) {
	const sqlIsEmpty = `SELECT count(*) FROM %s`
	var n int
	if err := e.db.QueryRowContext(
		ctx, fmt.Sprintf(sqlIsEmpty, quote(e.cfg.Table)),
	).Scan(&n); err != nil {
		return false, err
	}
	return n == 0, nil
}

// Upsert inserts a new document, or replaces the existing one whose
// string id is `path`.  The logic works with every SQLite ≥ 3.9 because
// it uses INSERT and INSERT OR REPLACE, both supported by FTS5.
func (e *Engine) Upsert(
	ctx context.Context,
	id string,
	vals map[string]string,
) error {
	if id == "" {
		return errors.New("ftsengine: empty id")
	}

	e.mu.Lock() // serialise writes inside *this* process
	defer e.mu.Unlock()

	/* --------------------------------------------------------------
	   1) Does a row with this path already exist?
	----------------------------------------------------------------*/
	var rowid int64
	const sqlSearchRow = `SELECT rowid FROM %s WHERE path=?`
	err := e.db.QueryRowContext(
		ctx,
		fmt.Sprintf(sqlSearchRow, quote(e.cfg.Table)),
		id,
	).Scan(&rowid)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err // real DB error
	}
	exists := !errors.Is(err, sql.ErrNoRows)

	/* --------------------------------------------------------------
	   2) Build column lists and args
	----------------------------------------------------------------*/
	numCols := 1 + len(e.cfg.Columns) // path + N user columns
	colNames := make([]string, 0, numCols)
	marks := make([]string, 0, numCols)
	args := make([]any, 0, numCols+1) // maybe +1 rowid

	// path column
	colNames = append(colNames, "path")
	marks = append(marks, "?")
	args = append(args, id)

	// user-defined columns
	for _, c := range e.cfg.Columns {
		colNames = append(colNames, quote(c.Name))
		marks = append(marks, "?")
		args = append(args, vals[c.Name])
	}

	/* --------------------------------------------------------------
	   3) Choose statement variant
	----------------------------------------------------------------*/
	var sqlQ string

	if exists {
		// prepend rowid for INSERT OR REPLACE
		colNames = append([]string{"rowid"}, colNames...)
		marks = append([]string{"?"}, marks...)
		args = append([]any{rowid}, args...)
		const sqlInsertOrReplace = `INSERT OR REPLACE INTO %s (%s) VALUES (%s);`
		sqlQ = fmt.Sprintf(
			sqlInsertOrReplace,
			quote(e.cfg.Table),
			strings.Join(colNames, ","),
			strings.Join(marks, ","),
		)
	} else {
		// fresh insert, rowid is omitted
		const sqlInsert = `INSERT INTO %s (%s) VALUES (%s);`
		sqlQ = fmt.Sprintf(
			sqlInsert,
			quote(e.cfg.Table),
			strings.Join(colNames, ","),
			strings.Join(marks, ","),
		)
	}

	_, err = e.db.ExecContext(ctx, sqlQ, args...)
	return err
}

func (e *Engine) Delete(ctx context.Context, id string) error {
	const sqlDel = `DELETE FROM %s WHERE path=?`
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.db.ExecContext(ctx,
		fmt.Sprintf(sqlDel, quote(e.cfg.Table)), id)
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
			ORDER BY s ASC, rowid
			LIMIT ? OFFSET ?;`

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

func (e *Engine) Close() error { return e.db.Close() }

/* ------------------------------------------------------------------ */
/*  helpers                                                           */
/* ------------------------------------------------------------------ */

func quote(id string) string { return `"` + strings.ReplaceAll(id, `"`, `""`) + `"` }

func paramPlaceholders(n int) string {
	if n == 0 {
		return ""
	}
	return strings.Repeat(",?", n)
}
