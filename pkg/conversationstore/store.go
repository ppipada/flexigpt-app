package conversationstore

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/conversationstore/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/dirstore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filenameprovider"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
)

type ConversationCollection struct {
	baseDir   string
	store     *dirstore.MapDirectoryStore
	fts       *ftsengine.Engine
	enableFTS bool
	// File-name builder / parser.
	fp filenameprovider.Provider
	// Directory partitioning.
	pp dirstore.PartitionProvider
}

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

// NewConversationCollection creates a collection with sensible defaults
// (UUID-v7 file names under yyyyMM partitions).  Callers may override either
// strategy via the Option functions above.
//
//	baseDir is the root directory for the map-directory store.
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

	// Optional full-text engine.
	if cc.enableFTS {
		var err error
		cc.fts, err = ftsengine.NewEngine(ftsengine.Config{
			BaseDir:    baseDir,
			DBFileName: "conversations.fts.sqlite",
			Table:      "conversations",
			Columns: []ftsengine.Column{
				{Name: "title", Weight: 1},
				{Name: "system", Weight: 2},
				{Name: "user", Weight: 3},
				{Name: "assistant", Weight: 4},
				{Name: "function", Weight: 5},
				{Name: "feedback", Weight: 6},
				{Name: "mtime", Unindexed: true},
			},
		})
		if err != nil {
			return nil, err
		}
		StartRebuild(
			context.Background(),
			baseDir,
			cc.fts,
		)
	}

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

func (cc *ConversationCollection) fileNameFromConversation(c spec.Conversation) (string, error) {
	return cc.fp.Build(filenameprovider.FileInfo{
		ID:        c.ID,
		Title:     c.Title,
		CreatedAt: c.CreatedAt,
	})
}

func (cc *ConversationCollection) PutConversation(
	ctx context.Context,
	req *spec.PutConversationRequest,
) (*spec.PutConversationResponse, error) {
	if req == nil || req.Body == nil || req.ID == "" || req.Body.Title == "" {
		return nil, errors.New("request or request body cannot be nil")
	}
	if req.ID == "" || req.Body.Title == "" {
		return nil, errors.New("request ID an title are required")
	}

	// Get filename from info.
	fn, _ := cc.fp.Build(filenameprovider.FileInfo{
		ID:        req.ID,
		Title:     req.Body.Title,
		CreatedAt: req.Body.CreatedAt,
	})
	partitionDirName := cc.pp.GetPartitionDir(fn)

	// Check if there are files with same id as prefix
	// We don't iterate as we expect only 1 file max with the id prefix of uuid.
	files, _, err := cc.store.ListFiles(
		dirstore.ListingConfig{
			FilenamePrefix:   req.ID,
			PageSize:         10,
			FilterPartitions: []string{partitionDirName},
		},
		"",
	)
	if err != nil {
		return nil, err
	}
	// If there is a file, that means its a replace of full conversation
	// May be title has also changed
	// Remove the current file and add new.
	for idx := range files {
		err := cc.store.DeleteFile(filepath.Base(files[idx]))
		if err != nil {
			slog.Warn("Put conversation remove existing file", "error", err)
		}
	}

	currentConversation := &spec.Conversation{}
	currentConversation.ID = req.ID
	currentConversation.Title = req.Body.Title
	currentConversation.CreatedAt = req.Body.CreatedAt
	currentConversation.ModifiedAt = req.Body.ModifiedAt
	currentConversation.Messages = req.Body.Messages

	data, err := encdec.StructWithJSONTagsToMap(currentConversation)
	if err != nil {
		return nil, err
	}
	if err := cc.store.SetFileData(fn, data); err != nil {
		return nil, err
	}
	return &spec.PutConversationResponse{}, nil
}

func (cc *ConversationCollection) PutMessagesToConversation(
	ctx context.Context,
	req *spec.PutMessagesToConversationRequest,
) (*spec.PutMessagesToConversationResponse, error) {
	if req == nil || req.Body == nil || req.Body.Messages == nil || len(req.Body.Messages) == 0 {
		return nil, errors.New("request or request body cannot be nil")
	}

	convoResp, err := cc.GetConversation(ctx,
		&spec.GetConversationRequest{ID: req.ID, Title: req.Body.Title})
	if err != nil {
		return nil, err
	}

	currentConversation := convoResp.Body
	currentConversation.ModifiedAt = time.Now()
	currentConversation.Messages = req.Body.Messages

	fn, err := cc.fileNameFromConversation(*currentConversation)
	if err != nil {
		return nil, err
	}

	data, err := encdec.StructWithJSONTagsToMap(currentConversation)
	if err != nil {
		return nil, err
	}
	if err := cc.store.SetFileData(fn, data); err != nil {
		return nil, err
	}

	return &spec.PutMessagesToConversationResponse{}, nil
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
	// Purge from FTS (absolute path = docID).
	if cc.fts != nil {
		full := filepath.Join(cc.baseDir, cc.pp.GetPartitionDir(fn), fn)
		_ = cc.fts.Delete(ctx, full)
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
	files, next, err := cc.store.ListFiles(
		dirstore.ListingConfig{SortOrder: dirstore.SortOrderDescending},
		token,
	)
	if err != nil {
		return nil, err
	}

	items := make([]spec.ConversationItem, 0, len(files))
	for _, f := range files {
		info, err := cc.fp.Parse(filepath.Base(f))
		if err != nil {
			// Corrupted/foreign file skip.
			continue
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
