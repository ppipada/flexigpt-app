package fts

import (
	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
)

const (
	compareColumn    = "mtime" // compareColumn is the field used for FTS sort order.
	ftsSyncBatchSize = 1000    // ftsSyncBatchSize is the batch size for FTS sync.
	enabledTrue      = "true"  // enabledTrue is the string value for enabled templates.
	enabledFalse     = "false" // enabledFalse is the string value for disabled templates.
	newline          = "\n"    // newline is the newline character.
	// Batch size is deliberately small - the number of built-ins is tiny.
	upsertBatchSize = 256
)

// Every built-in document ID is prefixed with this constant so that existing path-parsing code can still split bundleDir / fileName.
const (
	BuiltInDocPrefix  = "bi/"
	sqliteDBTableName = "prompttemplates"
)

type ftsDoc struct {
	Slug        bundleitemutils.ItemSlug `fts:"slug"`
	DisplayName string                   `fts:"displayName"`
	Desc        string                   `fts:"desc"`
	Messages    string                   `fts:"messages"`
	Tags        string                   `fts:"tags"`
	Enabled     string                   `fts:"enabled"`
	BundleID    bundleitemutils.BundleID `fts:"bundleID"`
	MTime       string                   `fts:"mtime"`
}

func (d ftsDoc) ToMap() map[string]string {
	return map[string]string{
		"slug":        string(d.Slug),
		"displayName": d.DisplayName,
		"desc":        d.Desc,
		"messages":    d.Messages,
		"tags":        d.Tags,
		"enabled":     d.Enabled,
		"bundleID":    string(d.BundleID),
		"mtime":       d.MTime,
	}
}

var ftsColumns = []ftsengine.Column{
	{Name: "slug", Weight: 1},
	{Name: "displayName", Weight: 2},
	{Name: "desc", Weight: 3},
	{Name: "messages", Weight: 4},
	{Name: "tags", Weight: 5},
	{Name: "enabled", Unindexed: true},
	{Name: "bundleID", Unindexed: true},
	{Name: "mtime", Unindexed: true},
}
