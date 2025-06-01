package conversationstore

import (
	"context"
	"path/filepath"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/conversationstore/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/dirstore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filenameprovider"
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

/* ------------------------------------------------------------------ */
/*  ConversationCollection                                            */
/* ------------------------------------------------------------------ */

type ConversationCollection struct {
	store *dirstore.MapDirectoryStore

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
	// default components ------------------------------------------------
	defFP := filenameprovider.UUIDv7Provider{}
	defPP := dirstore.MonthPartitionProvider{TimeFn: defFP.CreatedAt}

	cc := &ConversationCollection{
		fp: &defFP,
		pp: &defPP,
	}

	// apply caller overrides -------------------------------------------
	for _, o := range opts {
		if err := o(cc); err != nil {
			return nil, err
		}
	}

	// wire the underlying store ----------------------------------------
	store, err := dirstore.NewMapDirectoryStore(
		baseDir,
		true,
		dirstore.WithPartitionProvider(cc.pp),
	)
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
	fn, _ := cc.fp.Build(filenameprovider.FileInfo{ID: req.ID, Title: req.Title})
	if err := cc.store.DeleteFile(fn); err != nil {
		return nil, err
	}
	return &spec.DeleteConversationResponse{}, nil
}

func (cc *ConversationCollection) GetConversation(
	ctx context.Context,
	req *spec.GetConversationRequest,
) (*spec.GetConversationResponse, error) {
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

// The titles returned here are not from the conversation itself, but sanitized names with alpha numeric chars only
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
