package store

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/conversation/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/dirstore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
	"github.com/ppipada/mapstore-go/uuidv7filename"
)

type ConversationCollection struct {
	baseDir   string
	enableFTS bool
	store     *dirstore.MapDirectoryStore
	fts       *ftsengine.Engine
	pp        dirstore.PartitionProvider
}

type Option func(*ConversationCollection) error

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
	defPP := dirstore.MonthPartitionProvider{
		TimeFn: func(fileKey dirstore.FileKey) (time.Time, error) {
			u, err := uuidv7filename.Parse(fileKey.FileName)
			if err != nil {
				return time.Time{}, err
			}
			return u.Time, nil
		},
	}

	cc := &ConversationCollection{
		baseDir: filepath.Clean(baseDir),
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
	info, err := uuidv7filename.Build(req.ID, req.Body.Title, spec.ConversationFileExtension)
	if err != nil {
		return nil, err
	}
	filename := info.FileName
	partitionDirName, err := cc.pp.GetPartitionDir(dirstore.FileKey{FileName: filename})
	if err != nil {
		return nil, err
	}

	// Check if there are files with same id as prefix
	// We don't iterate as we expect only 1 file max with the id prefix of uuid.
	fileEntries, _, err := cc.store.ListFiles(
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
	for idx := range fileEntries {
		err := cc.store.DeleteFile(
			dirstore.FileKey{FileName: filepath.Base(fileEntries[idx].BaseRelativePath)},
		)
		if err != nil {
			slog.Warn("put conversation remove existing file", "error", err)
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
	if err := cc.store.SetFileData(dirstore.FileKey{FileName: filename}, data); err != nil {
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
		&spec.GetConversationRequest{ID: req.ID, Title: req.Body.Title, ForceFetch: false})
	if err != nil {
		return nil, err
	}

	currentConversation := convoResp.Body
	currentConversation.ModifiedAt = time.Now()
	currentConversation.Messages = req.Body.Messages

	filename, err := cc.fileNameFromConversation(*currentConversation)
	if err != nil {
		return nil, err
	}

	data, err := encdec.StructWithJSONTagsToMap(currentConversation)
	if err != nil {
		return nil, err
	}
	if err := cc.store.SetFileData(dirstore.FileKey{FileName: filename}, data); err != nil {
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
	info, err := uuidv7filename.Build(req.ID, req.Title, spec.ConversationFileExtension)
	if err != nil {
		return nil, err
	}
	filename := info.FileName

	if err := cc.store.DeleteFile(dirstore.FileKey{FileName: filename}); err != nil {
		return nil, err
	}
	slog.Info("delete conversation", "file", filename)
	return &spec.DeleteConversationResponse{}, nil
}

func (cc *ConversationCollection) GetConversation(
	ctx context.Context,
	req *spec.GetConversationRequest,
) (*spec.GetConversationResponse, error) {
	if req == nil || req.Title == "" || req.ID == "" {
		return nil, errors.New("request or request body cannot be nil")
	}
	info, err := uuidv7filename.Build(req.ID, req.Title, spec.ConversationFileExtension)
	if err != nil {
		return nil, err
	}
	filename := info.FileName

	// Force get the data for single get.
	raw, err := cc.store.GetFileData(dirstore.FileKey{FileName: filename}, req.ForceFetch)
	if err != nil {
		return nil, err
	}

	var convo spec.Conversation
	if err := encdec.MapToStructWithJSONTags(raw, &convo); err != nil {
		return nil, err
	}
	return &spec.GetConversationResponse{Body: &convo}, nil
}

// ListConversations
// The titles returned here are not from the conversation itself, but sanitized names with alpha
// numeric chars only.
func (cc *ConversationCollection) ListConversations(
	ctx context.Context,
	req *spec.ListConversationsRequest,
) (*spec.ListConversationsResponse, error) {
	token := ""
	pageSize := spec.DefaultPageSize
	if req != nil {
		token = req.PageToken
		if req.PageSize > 0 && req.PageSize <= spec.MaxPageSize {
			pageSize = req.PageSize
		}
	}
	fileEntries, next, err := cc.store.ListFiles(
		dirstore.ListingConfig{SortOrder: dirstore.SortOrderDescending, PageSize: pageSize},
		token,
	)
	if err != nil {
		return nil, err
	}

	items := make([]spec.ConversationListItem, 0, len(fileEntries))
	for _, f := range fileEntries {
		info, err := uuidv7filename.Parse(filepath.Base(f.BaseRelativePath))
		if err != nil {
			// Corrupted/foreign file skip.
			continue
		}
		fileModTime := f.FileInfo.ModTime()
		items = append(items, spec.ConversationListItem{
			ID:             info.ID,
			SanatizedTitle: info.Suffix,
			ModifiedAt:     &fileModTime,
		})
	}
	return &spec.ListConversationsResponse{
		Body: &spec.ListConversationsResponseBody{
			ConversationListItems: items,
			NextPageToken:         &next,
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
	pageSize := spec.DefaultPageSize
	if req.PageSize > 0 && req.PageSize <= spec.MaxPageSize {
		pageSize = req.PageSize
	}

	hits, next, err := cc.fts.Search(ctx, req.Query, req.PageToken, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]spec.ConversationListItem, 0, len(hits))
	for _, h := range hits {
		info, err := uuidv7filename.Parse(filepath.Base(h.ID))
		if err != nil {
			continue
		}
		items = append(items, spec.ConversationListItem{
			ID:             info.ID,
			SanatizedTitle: info.Suffix,
		})
	}
	return &spec.SearchConversationsResponse{
		Body: &spec.SearchConversationsResponseBody{
			ConversationListItems: items,
			NextPageToken:         &next,
		},
	}, nil
}

func (cc *ConversationCollection) fileNameFromConversation(c spec.Conversation) (string, error) {
	info, err := uuidv7filename.Build(c.ID, c.Title, spec.ConversationFileExtension)
	if err != nil {
		return "", err
	}
	return info.FileName, nil
}
