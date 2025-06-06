package conversationstore

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/conversationstore/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/dirstore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filenameprovider"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
)

/* ------------------------------------------------------------------ */
/*  Functional-options wiring                                         */
/* ------------------------------------------------------------------ */

type Option func(*ConversationCollection) error

func WithFilenameProvider(fp filenameprovider.Provider) Option {
	return func(cc *ConversationCollection) error {
		cc.fp = fp
		return nil
	}
}

func WithPartitionProvider(pp dirstore.PartitionProvider) Option {
	return func(cc *ConversationCollection) error {
		cc.pp = pp
		return nil
	}
}

func WithFTS(enabled bool) Option {
	return func(cc *ConversationCollection) error {
		cc.enableFTS = enabled
		return nil
	}
}

/* ------------------------------------------------------------------ */
/*  ConversationCollection                                            */
/* ------------------------------------------------------------------ */

type ConversationCollection struct {
	baseDir   string
	store     *dirstore.MapDirectoryStore
	fts       *ftsengine.Engine
	enableFTS bool // default = false

	// pluggable bits
	fp filenameprovider.Provider  // file-name builder / parser
	pp dirstore.PartitionProvider // directory partitioning
}

/*
NewConversationCollection creates a collection with sensible defaults
(UUID-v7 file names under yyyyMM partitions).  Callers may override either
strategy via the Option functions above.

	baseDir is the root directory for the map-directory store.
*/
func NewConversationCollection(baseDir string, opts ...Option) (*ConversationCollection, error) {
	defFP := filenameprovider.UUIDv7Provider{}
	defPP := dirstore.MonthPartitionProvider{TimeFn: defFP.CreatedAt}

	cc := &ConversationCollection{
		baseDir: filepath.Clean(baseDir),
		fp:      &defFP,
		pp:      &defPP,
	}

	for _, o := range opts {
		if err := o(cc); err != nil {
			return nil, err
		}
	}

	/* ------------- optional full-text engine ------------------------ */
	if cc.enableFTS {
		var err error
		cc.fts, err = ftsengine.New(ftsengine.Config{
			DBPath: filepath.Join(baseDir, "conversations.fts.sqlite"),
			Table:  "conversations",
			Columns: []ftsengine.Column{
				{Name: "title", Weight: 1},
				{Name: "system", Weight: 2},
				{Name: "user", Weight: 3},
				{Name: "assistant", Weight: 4},
				{Name: "function", Weight: 5},
				{Name: "feedback", Weight: 6},
			},
		})
		if err != nil {
			return nil, err
		}
		rebuildIfEmpty(baseDir, cc.fts)
	}

	/* ------------- MapDirectoryStore -------------------------------- */
	optsDir := []dirstore.Option{dirstore.WithPartitionProvider(cc.pp)}
	if cc.fts != nil {
		optsDir = append(optsDir, dirstore.WithListeners(NewFTSListner(cc.fts)))
	}
	store, err := dirstore.NewMapDirectoryStore(baseDir, true, optsDir...)
	if err != nil {
		return nil, err
	}
	cc.store = store
	return cc, nil
}

/* ------------------------------------------------------------------ */
/*  internal helpers                                                  */
/* ------------------------------------------------------------------ */

func (cc *ConversationCollection) fileNameFromConversation(c spec.Conversation) (string, error) {
	return cc.fp.Build(filenameprovider.FileInfo{
		ID:        c.ID,
		Title:     c.Title,
		CreatedAt: c.CreatedAt,
	})
}

/* ------------------------------------------------------------------ */
/*  Public API                                                        */
/* ------------------------------------------------------------------ */

func (cc *ConversationCollection) SaveConversation(
	ctx context.Context,
	req *spec.SaveConversationRequest,
) (*spec.SaveConversationResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or request body cannot be nil")
	}

	fn, err := cc.fileNameFromConversation(*req.Body)
	if err != nil {
		return nil, err
	}

	data, err := encdec.StructWithJSONTagsToMap(req.Body)
	if err != nil {
		return nil, err
	}
	if err := cc.store.SetFileData(fn, data); err != nil {
		return nil, err
	}
	return &spec.SaveConversationResponse{}, nil
}

func (cc *ConversationCollection) DeleteConversation(
	ctx context.Context,
	req *spec.DeleteConversationRequest,
) (*spec.DeleteConversationResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}
	fn, _ := cc.fp.Build(filenameprovider.FileInfo{ID: req.ID, Title: req.Title})
	if err := cc.store.DeleteFile(fn); err != nil {
		return nil, err
	}
	// purge from FTS (absolute path = docID)
	if cc.fts != nil {
		full := filepath.Join(cc.baseDir, cc.pp.GetPartitionDir(fn), fn)
		_ = cc.fts.Delete(full)
	}
	return &spec.DeleteConversationResponse{}, nil
}

func (cc *ConversationCollection) GetConversation(
	ctx context.Context,
	req *spec.GetConversationRequest,
) (*spec.GetConversationResponse, error) {
	if req == nil {
		return nil, errors.New("request or request body cannot be nil")
	}
	fn, _ := cc.fp.Build(filenameprovider.FileInfo{ID: req.ID, Title: req.Title})
	raw, err := cc.store.GetFileData(fn, false)
	if err != nil {
		return nil, err
	}

	var convo spec.Conversation
	if err := encdec.MapToStructWithJSONTags(raw, &convo); err != nil {
		return nil, err
	}
	return &spec.GetConversationResponse{Body: &convo}, nil
}

// The titles returned here are not from the conversation itself, but sanitized names with alpha numeric chars only.
func (cc *ConversationCollection) ListConversations(
	ctx context.Context,
	req *spec.ListConversationsRequest,
) (*spec.ListConversationsResponse, error) {
	token := ""
	if req != nil {
		token = req.Token
	}
	files, next, err := cc.store.ListFiles("desc", token)
	if err != nil {
		return nil, err
	}

	items := make([]spec.ConversationItem, 0, len(files))
	for _, f := range files {
		info, err := cc.fp.Parse(filepath.Base(f))
		if err != nil {
			continue // corrupted / foreign file – skip
		}
		items = append(items, spec.ConversationItem{
			ID:        info.ID,
			Title:     info.Title,
			CreatedAt: info.CreatedAt,
		})
	}
	return &spec.ListConversationsResponse{
		Body: &spec.ListConversationsResponseBody{
			ConversationItems: items,
			NextPageToken:     &next,
		},
	}, nil
}

func (cc *ConversationCollection) SearchConversations(
	ctx context.Context,
	req *spec.SearchConversationsRequest,
) (*spec.SearchConversationsResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}
	if cc.fts == nil {
		return nil, errors.New("full-text search is disabled")
	}
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 10
	}

	hits, next, err := cc.fts.Search(ctx, req.Query, req.Token, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]spec.ConversationItem, 0, len(hits))
	for _, h := range hits {
		base := filepath.Base(h.ID)
		info, err := cc.fp.Parse(base)
		if err != nil {
			continue
		}
		items = append(items, spec.ConversationItem{
			ID:        info.ID,
			Title:     info.Title,
			CreatedAt: info.CreatedAt,
		})
	}
	return &spec.SearchConversationsResponse{
		Body: &spec.SearchConversationsResponseBody{
			ConversationItems: items,
			NextPageToken:     &next,
		},
	}, nil
}

func (cc *ConversationCollection) AddMessageToConversation(
	ctx context.Context,
	req *spec.AddMessageToConversationRequest,
) (*spec.AddMessageToConversationResponse, error) {
	convoResp, err := cc.GetConversation(ctx,
		&spec.GetConversationRequest{ID: req.ID, Title: req.Body.Title})
	if err != nil {
		return nil, err
	}

	convoResp.Body.Messages = append(convoResp.Body.Messages, req.Body.NewMessage)
	convoResp.Body.ModifiedAt = time.Now()

	if _, err := cc.SaveConversation(
		ctx, &spec.SaveConversationRequest{Body: convoResp.Body}); err != nil {
		return nil, err
	}
	return &spec.AddMessageToConversationResponse{}, nil
}
