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
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"unicode"

	_ "github.com/glebarez/go-sqlite"
)

const (
	tokenizerOptions = "porter unicode61 remove_diacritics 1"
)

var ErrEmptyQuery = errors.New("empty query")

type Engine struct {
	db  *sql.DB
	cfg Config
	// Schema checksum.
	hsh string
	// Serializes write-queries.
	mu sync.Mutex
}

func NewEngine(cfg Config) (*Engine, error) {
	err := validateConfig(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.BaseDir != MemoryDBBaseDir {
		// Idempotent - harmless if it already exists.
		if err := os.MkdirAll(cfg.BaseDir, 0o770); err != nil {
			return nil, err
		}
	}

	dataSourceName := filepath.Join(
		cfg.BaseDir,
		cfg.DBFileName,
	)

	db, err := sql.Open("sqlite", dataSourceName+"?busy_timeout=5000&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)

	e := &Engine{db: db, cfg: cfg}
	e.hsh = schemaChecksum(e.cfg, tokenizerOptions)
	slog.Info("ftsengine bootstrap", "dbPath", dataSourceName)
	if err := e.bootstrap(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return e, nil
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

func (e *Engine) Delete(ctx context.Context, id string) error {
	const sqlDel = `DELETE FROM %s WHERE %s=?`
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.db.ExecContext(ctx,
		fmt.Sprintf(sqlDel, quote(e.cfg.Table), ColNameExternalID), id)
	return err
}

func (e *Engine) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// SQLite default.
	const maxVars = 999
	toAny := func(ss []string) []any {
		out := make([]any, len(ss))
		for i, s := range ss {
			out[i] = s
		}
		return out
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	for len(ids) != 0 {
		n := min(len(ids), maxVars)
		part := ids[:n]
		ids = ids[n:]

		var b strings.Builder
		for i := range part {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteByte('?')
		}
		const sqlDelete = `DELETE FROM %s WHERE %s IN (%s);`
		sqlQ := fmt.Sprintf(sqlDelete, quote(e.cfg.Table), ColNameExternalID, b.String())

		if _, err := e.db.ExecContext(ctx, sqlQ, toAny(part)...); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) Close() error { return e.db.Close() }

// Upsert inserts a new document, or replaces the existing one whose string id is present.
// The logic works with every SQLite ≥ 3.9 because it uses INSERT and INSERT OR REPLACE, both supported by FTS5.
// This is not multi process safe as this is serialized at application level.
func (e *Engine) Upsert(ctx context.Context, id string, vals map[string]string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.internalUpsert(ctx, nil, id, vals)
}

// BatchUpsert writes / updates all docs inside ONE transaction.
// The map key is the externalID, the value is the column map.
func (e *Engine) BatchUpsert(
	ctx context.Context,
	docs map[string]map[string]string,
) error {
	if len(docs) == 0 {
		return nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	commit := func(err error) error {
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		return tx.Commit()
	}

	// Gather existing rowids in one probe.
	ids := make([]string, 0, len(docs))
	for id := range docs {
		ids = append(ids, id)
	}
	existing, err := e.lookupRowIDs(ctx, tx, ids)
	if err != nil {
		return commit(err)
	}

	for id, vals := range docs {
		if err := e.internalUpsert(ctx, tx, id, vals, existing[id]); err != nil {
			return commit(err)
		}
	}
	return commit(nil)
}

// BatchList pages over the whole table ordered by `compareColumn` + rowid.
// If compareColumn == "" it falls back to ordering by rowid only (fast path).
// WantedCols limits the data that is returned to the caller.
// The slice must be a subset of cfg.Columns.
// Nil / empty means "all".
//
// Returns rows, an opaque nextToken ("" == no more rows) and an error.
func (e *Engine) BatchList(
	ctx context.Context,
	compareColumn string,
	wantedCols []string,
	pageToken string,
	pageSize int,
) (rows []ListResult, nextToken string, err error) {
	if pageSize <= 0 {
		pageSize = 1000
	}
	if pageSize > 10000 {
		pageSize = 10000
	}

	// Validate / canonicalise wantedCols.
	colExists := func(name string) bool {
		for _, c := range e.cfg.Columns {
			if c.Name == name {
				return true
			}
		}
		return false
	}
	if len(wantedCols) == 0 {
		wantedCols = make([]string, 0, len(e.cfg.Columns))
		for _, c := range e.cfg.Columns {
			wantedCols = append(wantedCols, c.Name)
		}
	} else {
		for _, n := range wantedCols {
			if !colExists(n) {
				return nil, "", fmt.Errorf("ftsengine: unknown column %q", n)
			}
		}
	}

	if compareColumn == "" {
		compareColumn = ColNameRowID
	} else if compareColumn != ColNameRowID && !colExists(compareColumn) {
		return nil, "", fmt.Errorf("ftsengine: unknown compare column %q", compareColumn)
	}

	// Decode continuation token.
	var (
		// TEXT comparison   (rowid compare: unused).
		lastCmp string
		// Always included to disambiguate duplicates.
		lastRID int64
	)
	if pageToken != "" {
		var t struct {
			C string `json:"c"`
			R int64  `json:"r"`
		}
		if b, _ := base64.StdEncoding.DecodeString(pageToken); len(b) > 0 {
			_ = json.Unmarshal(b, &t)
			lastCmp = t.C
			lastRID = t.R
		}
	}

	// Build SELECT list.
	selectCols := []string{ColNameRowID, ColNameExternalID}
	needCmpInSelect := compareColumn != ColNameRowID
	if needCmpInSelect {
		selectCols = append(selectCols, quote(compareColumn))
	}
	wantedColsNoCompare := make([]string, 0, len(wantedCols))
	for _, c := range wantedCols {
		if c == compareColumn {
			continue
		}
		selectCols = append(selectCols, quote(c))
		wantedColsNoCompare = append(wantedColsNoCompare, c)
	}

	// Build WHERE + ORDER BY.
	var where string
	var args []any
	if compareColumn == ColNameRowID {
		where = ColNameRowID + ">?"
		args = append(args, lastRID)
	} else {
		// Actual: (cmp > lastCmp) OR (cmp = lastCmp AND rowid > lastRID).
		where = fmt.Sprintf("(%s>? OR (%s=? AND %s>?))",
			quote(compareColumn), quote(compareColumn), ColNameRowID)
		args = append(args, lastCmp, lastCmp, lastRID)
	}

	// We fetch one extra row to know if more data exists.
	limitRows := pageSize + 1
	args = append(args, limitRows)

	const sqlSelect = `SELECT %s FROM %s WHERE %s ORDER BY %s,%s LIMIT ?;`
	sqlQ := fmt.Sprintf(sqlSelect,
		strings.Join(selectCols, ","),
		quote(e.cfg.Table),
		where,
		quote(compareColumn),
		ColNameRowID,
	)

	// One read-only tx per page.
	tx, err := e.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = tx.Rollback() }()

	r, err := tx.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, "", err
	}
	defer r.Close()

	// Prepare scan dest.
	destCount := len(selectCols)
	dest := make([]any, destCount)

	var ridHolder int64
	var idHolder string
	dest[0] = &ridHolder
	dest[1] = &idHolder

	cmpIdx := -1
	if needCmpInSelect {
		cmpIdx = 2
		dest[cmpIdx] = new(sql.NullString)
	}

	// Remaining wanted columns.
	valHolders := make([]sql.NullString, len(wantedColsNoCompare))
	// Position after rowid, externalID (+ maybe compareCol).
	off := 2
	if needCmpInSelect {
		off++
	}
	for i := range valHolders {
		dest[off+i] = &valHolders[i]
	}

	var haveMore bool
	for r.Next() {
		if err := r.Scan(dest...); err != nil {
			return nil, "", err
		}

		// If we've already collected pageSize rows, this is the +1 look-ahead.
		if len(rows) >= pageSize {
			haveMore = true
			break
		}

		vals := make(map[string]string, len(wantedCols))
		j := 0
		for _, col := range wantedCols {
			if col == compareColumn {
				// If user requested compareColumn, get it from cmpIdx.
				if cmpIdx >= 0 {
					if nv, ok := dest[cmpIdx].(*sql.NullString); ok && nv.Valid {
						vals[col] = nv.String
					}
				}
			} else {
				if valHolders[j].Valid {
					vals[col] = valHolders[j].String
				}
				j++
			}
		}
		rows = append(rows, ListResult{ID: idHolder, Values: vals})
		lastRID = ridHolder
		if cmpIdx >= 0 {
			if nv, ok := dest[cmpIdx].(*sql.NullString); ok && nv.Valid {
				lastCmp = nv.String
			}
		}
	}
	if err := r.Err(); err != nil {
		return nil, "", err
	}
	if err := tx.Commit(); err != nil {
		return nil, "", err
	}

	// Produce nextToken only if a further row exists.
	if haveMore {
		buf, _ := json.Marshal(struct {
			C string `json:"c"`
			R int64  `json:"r"`
		}{lastCmp, lastRID})
		nextToken = base64.StdEncoding.EncodeToString(buf)
	}
	return rows, nextToken, nil
}

// Search returns one page of results and, if more results exist,
// an opaque token for the next page.
// The query is treated as a search literal and not a fts5 expression.
func (e *Engine) Search(
	ctx context.Context,
	query string,
	pageToken string,
	pageSize int,
) (hits []SearchResult, nextToken string, err error) {
	if query == "" {
		return nil, "", ErrEmptyQuery
	}

	if pageSize <= 0 || pageSize > 10000 {
		pageSize = 10
	}

	// Decode / reset token.
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
		// Token belongs to same query.
		if t.Query == query {
			offset = t.Offset
		}
	}

	// Bm25 weight parameters, one per column.
	var weights []any
	for _, c := range e.cfg.Columns {
		if c.Weight == 0 {
			weights = append(weights, float64(1))
		} else {
			weights = append(weights, c.Weight)
		}
	}

	const sqlSearch = `SELECT %s, bm25(%s%s) AS s
			FROM %s WHERE %s MATCH ?
			ORDER BY s ASC, %s
			LIMIT ? OFFSET ?;`

	sqlQ := fmt.Sprintf(sqlSearch, ColNameExternalID,
		quote(e.cfg.Table), paramPlaceholders(len(weights)),
		quote(e.cfg.Table), e.cfg.Table, ColNameRowID)

	args := slices.Clone(weights)
	// Escape any embedded double quotes.
	// FTS5 has special chars like - * etc that only quote for SQL, not for token.
	cQ := cleanQueryWithOr(query)
	if cQ == "" {
		// Return empty result.
		return []SearchResult{}, "", nil
	}
	args = append(args, cQ, pageSize, offset)

	rows, err := e.db.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ID, &r.Score); err != nil {
			return nil, "", err
		}
		hits = append(hits, r)
	}

	// Build next token.
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

func (e *Engine) bootstrap(ctx context.Context) error {
	const sqlCreateMetaTable = `CREATE TABLE IF NOT EXISTS meta(k TEXT PRIMARY KEY,v TEXT);`
	const sqlSelectMetaHash = `SELECT v FROM meta WHERE k='h'`
	const sqlInsertMetaHash = `INSERT OR REPLACE INTO meta(k,v) VALUES('h',?)`
	const sqlDropTable = `DROP TABLE IF EXISTS %s`
	const sqlCreateVirtualTable = `CREATE VIRTUAL TABLE IF NOT EXISTS %s
		USING fts5 (%s,
			tokenize='%s');`
	const sqlDeleteAllRows = `DELETE FROM %s`

	// Meta for schema hash.
	if _, err := e.db.ExecContext(ctx, sqlCreateMetaTable); err != nil {
		return err
	}

	// Existing hash.
	var stored string
	_ = e.db.QueryRowContext(ctx, sqlSelectMetaHash).Scan(&stored)

	// Create / replace FTS virtual table.
	slog.Debug("fst-engine bootstrap", "previousChecksum", stored, "newChecksum", e.hsh)
	if stored != e.hsh {
		// Schema changed, clear previous rows.
		if stored != "" {
			slog.Info("fst-engine bootstrap: config checksum mismatch, delete all rows.")
			_, _ = e.db.ExecContext(ctx, fmt.Sprintf(sqlDeleteAllRows, quote(e.cfg.Table)))
		}
		slog.Info("fst-engine bootstrap: config checksum mismatch, create virtual table again.")
		_, _ = e.db.ExecContext(ctx, fmt.Sprintf(sqlDropTable, quote(e.cfg.Table)))

		var cols []string
		cols = append(cols, ColNameExternalID+" UNINDEXED")
		for _, c := range e.cfg.Columns {
			col := c.Name
			if c.Unindexed {
				col += " UNINDEXED"
			}
			cols = append(cols, col)
		}
		ddl := fmt.Sprintf(sqlCreateVirtualTable,
			quote(e.cfg.Table), strings.Join(cols, ","), tokenizerOptions)

		if _, err := e.db.ExecContext(ctx, ddl); err != nil {
			return err
		}
		_, _ = e.db.ExecContext(ctx, sqlInsertMetaHash, e.hsh)

	}
	return nil
}

func (e *Engine) lookupRowIDs(
	ctx context.Context,
	exec sqlExec,
	ids []string,
) (map[string]int64, error) {
	if len(ids) == 0 {
		return nil, errors.New("got empty id's for lookup")
	}
	var b strings.Builder
	for i := range ids {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('?')
	}

	sqlQ := fmt.Sprintf(`SELECT %s,%s FROM %s WHERE %s IN (%s);`,
		ColNameExternalID, ColNameRowID, quote(e.cfg.Table), ColNameExternalID, b.String())

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	rows, err := exec.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]int64, len(ids))
	for rows.Next() {
		var id string
		var rid int64
		if err := rows.Scan(&id, &rid); err != nil {
			return nil, err
		}
		out[id] = rid
	}
	return out, rows.Err()
}

// internalUpsert is shared by Upsert and BatchUpsert.
// If tx == nil the engine's *sql.DB is used, otherwise the provided *sql.Tx is used.
func (e *Engine) internalUpsert(
	ctx context.Context,
	tx *sql.Tx,
	id string,
	vals map[string]string,
	// Optional optimisation from BatchUpsert.
	knownRowID ...int64,
) error {
	if id == "" {
		return errors.New("ftsengine: empty id")
	}

	var exec sqlExec = e.db
	if tx != nil {
		exec = tx
	}

	// Determine whether the document already exists.
	var (
		exists bool
		rowid  int64
	)
	if len(knownRowID) == 1 && knownRowID[0] > 0 {
		// Caller already knows the rowid.
		exists = true
		rowid = knownRowID[0]
	} else {
		sqlQ := fmt.Sprintf(`SELECT %s FROM %s WHERE %s=?`, ColNameRowID, quote(e.cfg.Table), ColNameExternalID)
		rows, err := exec.QueryContext(ctx, sqlQ, id)
		if err != nil {
			return err
		}
		defer rows.Close()
		if rows.Next() {
			if err := rows.Scan(&rowid); err != nil {
				return err
			}
			exists = true
		}
		if err := rows.Err(); err != nil {
			return err
		}
	}

	// Build column list, placeholders and args slice.
	colNames := []string{ColNameExternalID}
	marks := []string{"?"}
	args := []any{id}

	for _, c := range e.cfg.Columns {
		colNames = append(colNames, quote(c.Name))
		marks = append(marks, "?")
		args = append(args, vals[c.Name])
	}

	// Choose INSERT vs INSERT OR REPLACE.
	var sqlQ string
	if exists {
		colNames = append([]string{ColNameRowID}, colNames...)
		marks = append([]string{"?"}, marks...)
		args = append([]any{rowid}, args...)

		sqlQ = fmt.Sprintf(`INSERT OR REPLACE INTO %s (%s) VALUES (%s);`,
			quote(e.cfg.Table),
			strings.Join(colNames, ","),
			strings.Join(marks, ","))
	} else {
		sqlQ = fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s);`,
			quote(e.cfg.Table),
			strings.Join(colNames, ","),
			strings.Join(marks, ","))
	}

	_, err := exec.ExecContext(ctx, sqlQ, args...)
	return err
}

// cleanQueryWithOr converts a raw string into `"a" OR "b" OR "c"`.
// Expect input: words separated by blanks.
func cleanQueryWithOr(q string) string {
	var tokens []string
	var buf strings.Builder

	flush := func() {
		if buf.Len() == 0 {
			return
		}
		tokens = append(tokens, buf.String())
		buf.Reset()
	}

	for _, r := range q {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			buf.WriteRune(r)
		} else {
			flush()
		}
	}
	// Final word.
	flush()

	// Nothing to search for, only non alphanumeric input.
	if len(tokens) == 0 {
		// Caller can skip the SQL.
		return ""
	}

	// Deduplicate *before* quoting.
	seen := make(map[string]struct{}, len(tokens))
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if len(t) == 1 {
			if !unicode.IsDigit(rune(t[0])) {
				// Skip 1 char strings.
				continue
			}
		}
		if _, ok := seen[t]; !ok {
			seen[t] = struct{}{}
			out = append(out, quote(t))
		}
	}

	if len(out) == 0 {
		return strings.Join(tokens, " OR ")
	}

	return strings.Join(out, " OR ")
}

func validateConfig(c Config) error {
	if len(c.Columns) == 0 {
		return errors.New("ftsengine: need ≥1 column")
	}
	if c.BaseDir == "" {
		return errors.New("ftsengine: DB BaseDir incorrect")
	}
	if c.BaseDir == MemoryDBBaseDir && c.DBFileName != "" {
		return errors.New("ftsengine: DB filename should be empty for memory db")
	}
	if c.BaseDir != MemoryDBBaseDir && c.DBFileName == "" {
		return errors.New("ftsengine: DB filename incorrect")
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

func schemaChecksum(cfg Config, extra string) string {
	h := sha256.New()
	// Write the extra string first.
	h.Write([]byte(extra))
	_ = json.NewEncoder(h).Encode(cfg)
	return hex.EncodeToString(h.Sum(nil))
}

func paramPlaceholders(n int) string {
	if n == 0 {
		return ""
	}
	return strings.Repeat(",?", n)
}

func quote(id string) string { return `"` + strings.ReplaceAll(id, `"`, `""`) + `"` }
