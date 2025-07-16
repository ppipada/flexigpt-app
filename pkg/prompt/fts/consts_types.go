package fts

import "github.com/ppipada/flexigpt-app/pkg/prompt/spec"

const (
	compareColumn    = "mtime" // compareColumn is the field used for FTS sort order.
	ftsSyncBatchSize = 1000    // ftsSyncBatchSize is the batch size for FTS sync.
	enabledTrue      = "true"  // enabledTrue is the string value for enabled templates.
	enabledFalse     = "false" // enabledFalse is the string value for disabled templates.
	newline          = "\n"    // newline is the newline character.
	// Every built-in document ID is prefixed with this constant so that existing path-parsing code can still split bundleDir / fileName.
	builtinDocPrefix = "builtin/"
	// Batch size is deliberately small - the number of built-ins is tiny.
	upsertBatchSize = 256
)

type ftsDoc struct {
	Slug        spec.TemplateSlug `fts:"slug"`
	DisplayName string            `fts:"displayName"`
	Desc        string            `fts:"desc"`
	Messages    string            `fts:"messages"`
	Tags        string            `fts:"tags"`
	Enabled     string            `fts:"enabled"`
	BundleID    spec.BundleID     `fts:"bundleId"`
	MTime       string            `fts:"mtime"`
}

func (d ftsDoc) ToMap() map[string]string {
	return map[string]string{
		"slug":        string(d.Slug),
		"displayName": d.DisplayName,
		"desc":        d.Desc,
		"messages":    d.Messages,
		"tags":        d.Tags,
		"enabled":     d.Enabled,
		"bundleId":    string(d.BundleID),
		"mtime":       d.MTime,
	}
}
