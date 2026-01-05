package fts

import (
	"github.com/ppipada/mapstore-go/ftsengine"

	"github.com/ppipada/flexigpt-app/internal/bundleitemutils"
)

const (
	// CompareColumn is the field used for FTS sort order.
	compareColumn    = "mtime"
	ftsSyncBatchSize = 1000
	enabledTrue      = "true"
	enabledFalse     = "false"
	newline          = "\n"

	// Batch size is deliberately small - the number of built-ins is tiny.
	upsertBatchSize = 256
)

// Every built-in document ID is prefixed with this constant so that existing
// path-parsing code can still split bundleDir / fileName.
const (
	BuiltInDocPrefix  = "bi/"
	sqliteDBTableName = "tools"
)

// One row in the FTS table.
type ftsDoc struct {
	Slug        bundleitemutils.ItemSlug `fts:"slug"`
	DisplayName string                   `fts:"displayName"`
	Desc        string                   `fts:"desc"`

	Args     string `fts:"args"`
	Impl     string `fts:"impl"`
	ImplMeta string `fts:"implMeta"`

	Tags     string                   `fts:"tags"`
	Enabled  string                   `fts:"enabled"`
	BundleID bundleitemutils.BundleID `fts:"bundleID"`
	MTime    string                   `fts:"mtime"`
}

// ToMap converts the doc to a string map for the engine.
func (d ftsDoc) ToMap() map[string]string {
	return map[string]string{
		"slug":        string(d.Slug),
		"displayName": d.DisplayName,
		"desc":        d.Desc,
		"args":        d.Args,
		"impl":        d.Impl,
		"implMeta":    d.ImplMeta,
		"tags":        d.Tags,
		"enabled":     d.Enabled,
		"bundleID":    string(d.BundleID),
		"mtime":       d.MTime,
	}
}

// All indexed / stored columns for the tools table.
var ftsColumns = []ftsengine.Column{
	{Name: "slug", Weight: 1},
	{Name: "displayName", Weight: 2},
	{Name: "desc", Weight: 3},
	{Name: "args", Weight: 4},
	{Name: "tags", Weight: 5},
	{Name: "impl", Weight: 6},
	{Name: "implMeta", Weight: 7},

	{Name: "enabled", Unindexed: true},
	{Name: "bundleID", Unindexed: true},
	{Name: "mtime", Unindexed: true},
}
