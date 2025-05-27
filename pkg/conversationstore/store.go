package conversationstore

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/conversationstore/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/dirstore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
	"github.com/google/uuid"
)

func GetDateFromUUIDv7(uuidStr string) (time.Time, error) {
	// Check if the UUID is valid and has the correct length
	if len(uuidStr) != 36 {
		return time.Time{}, errors.New("invalid UUIDv7 string")
	}

	// Remove dashes from the UUID
	uuidStr = uuidStr[:8] + uuidStr[9:13] + uuidStr[14:18] + uuidStr[19:23] + uuidStr[24:]

	// Parse the UUID into bytes
	uuidBytes, err := hex.DecodeString(uuidStr)
	if err != nil {
		return time.Time{}, err
	}

	// Prepare a slice to hold the 8-byte timestamp
	timestampBytes := make([]byte, 8)

	// Set the first 6 bytes from the UUID bytes (first 48 bits)
	copy(timestampBytes[2:], uuidBytes[:6])

	// Convert the 8-byte array into a 64-bit integer
	timestampMs := int64(0)
	for i := range 8 {
		timestampMs = (timestampMs << 8) | int64(timestampBytes[i])
	}

	// Convert the timestamp from milliseconds to a time.Time object
	return time.Unix(0, timestampMs*int64(time.Millisecond)), nil
}

func InitConversation(title string) (*spec.Conversation, error) {
	if title == "" {
		title = "New Conversation"
	}
	if len(title) > 64 {
		title = title[:64]
	}

	c := spec.Conversation{}
	u, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	c.ID = u.String()
	c.Title = title
	c.CreatedAt = time.Now()
	c.ModifiedAt = time.Now()
	c.Messages = []spec.ConversationMessage{}

	return &c, nil
}

type ConversationCollection struct {
	store *dirstore.MapDirectoryStore
}

func InitConversationCollection(convoCollection *ConversationCollection, baseDir string) error {
	partitionProvider := &dirstore.MonthBasedPartitionProvider{}
	store, err := dirstore.NewMapDirectoryStore(
		baseDir,
		true,
		dirstore.WithPartitionProvider(partitionProvider),
	)
	if err != nil {
		return err
	}
	convoCollection.store = store
	return nil
}

func (cc *ConversationCollection) GetConversationFilename(conversation spec.Conversation) string {
	// Create a regular expression to match non-alphanumeric characters.
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)

	// Replace non-alphanumeric characters in the title with underscores.
	sanitizedTitle := re.ReplaceAllString(conversation.Title, "_")

	// Format the filename using the conversation ID and sanitized title.
	filename := fmt.Sprintf("%s_%s.json", conversation.ID, sanitizedTitle)

	return filename
}

func (cc *ConversationCollection) getConversationFilenameUsingIDTitle(id, title string) string {
	c := spec.Conversation{}
	c.ID = id
	c.Title = title
	return cc.GetConversationFilename(c)
}

func (cc *ConversationCollection) SaveConversation(
	ctx context.Context,
	req *spec.SaveConversationRequest,
) (*spec.SaveConversationResponse, error) {
	filename := cc.GetConversationFilename(*req.Body)
	data, err := encdec.StructWithJSONTagsToMap(req.Body)
	if err != nil {
		return nil, err
	}
	err = cc.store.SetFileData(filename, data)
	if err != nil {
		return nil, err
	}
	return &spec.SaveConversationResponse{}, nil
}

func (cc *ConversationCollection) DeleteConversation(
	ctx context.Context,
	req *spec.DeleteConversationRequest,
) (*spec.DeleteConversationResponse, error) {
	err := cc.store.DeleteFile(cc.getConversationFilenameUsingIDTitle(req.ID, req.Title))
	if err != nil {
		return nil, err
	}
	return &spec.DeleteConversationResponse{}, nil
}

func (cc *ConversationCollection) GetConversation(
	ctx context.Context,
	req *spec.GetConversationRequest,
) (*spec.GetConversationResponse, error) {
	data, err := cc.store.GetFileData(
		cc.getConversationFilenameUsingIDTitle(req.ID, req.Title),
		false,
	)
	if err != nil {
		return nil, err
	}
	convo := spec.Conversation{}
	err = encdec.MapToStructWithJSONTags(data, &convo)
	if err != nil {
		return nil, err
	}
	return &spec.GetConversationResponse{Body: &convo}, nil
}

func (cc *ConversationCollection) ListConversations(
	ctx context.Context,
	req *spec.ListConversationsRequest,
) (*spec.ListConversationsResponse, error) {
	token := ""
	if req != nil {
		token = req.Token
	}
	files, nextToken, err := cc.store.ListFiles("desc", token)
	if err != nil {
		return nil, err
	}

	convoItems := make([]spec.ConversationItem, 0, len(files))
	for _, file := range files {
		filename := filepath.Base(file)
		parts := strings.Split(strings.TrimSuffix(filename, ".json"), "_")
		id := parts[0]
		title := strings.Join(parts[1:], " ")
		d, err := GetDateFromUUIDv7(id)
		if err != nil {
			slog.Warn("Could not get date. Skipping.", "ID", id)
			continue
		}
		convo := spec.ConversationItem{
			ID:        id,
			Title:     title,
			CreatedAt: d,
		}
		convoItems = append(convoItems, convo)
	}

	resp := &spec.ListConversationsResponseBody{
		ConversationItems: convoItems,
		NextPageToken:     &nextToken,
	}
	return &spec.ListConversationsResponse{Body: resp}, nil
}

func (cc *ConversationCollection) AddMessageToConversation(
	ctx context.Context,
	req *spec.AddMessageToConversationRequest,
) (*spec.AddMessageToConversationResponse, error) {
	conversation, err := cc.GetConversation(
		ctx,
		&spec.GetConversationRequest{ID: req.ID, Title: req.Body.Title},
	)
	if err != nil {
		return nil, err
	}

	conversation.Body.Messages = append(conversation.Body.Messages, req.Body.NewMessage)
	conversation.Body.ModifiedAt = time.Now()
	_, err = cc.SaveConversation(ctx, &spec.SaveConversationRequest{Body: conversation.Body})
	if err != nil {
		return nil, err
	}
	return &spec.AddMessageToConversationResponse{}, nil
}
