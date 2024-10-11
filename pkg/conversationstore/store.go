package conversationstore

import (
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/flexigpt/flexiui/pkg/conversationstore/spec"
	"github.com/flexigpt/flexiui/pkg/simplemapdb/dirstore"
	"github.com/flexigpt/flexiui/pkg/simplemapdb/encdec"
	"github.com/google/uuid"
)

func GetDateFromUUIDv7(uuid string) (time.Time, error) {
	// Check if the UUID is valid and has the correct length
	if len(uuid) != 36 {
		return time.Time{}, errors.New("invalid UUIDv7 string")
	}

	// Remove dashes from the UUID
	uuid = uuid[:8] + uuid[9:13] + uuid[14:18] + uuid[19:23] + uuid[24:]

	// Parse the UUID into bytes
	uuidBytes, err := hex.DecodeString(uuid)
	if err != nil {
		return time.Time{}, err
	}

	// Prepare a slice to hold the 8-byte timestamp
	timestampBytes := make([]byte, 8)

	// Set the first 6 bytes from the UUID bytes (first 48 bits)
	copy(timestampBytes[2:], uuidBytes[:6])

	// Convert the 8-byte array into a 64-bit integer
	timestampMs := int64(0)
	for i := 0; i < 8; i++ {
		timestampMs = (timestampMs << 8) | int64(timestampBytes[i])
	}

	// Convert the timestamp from milliseconds to a time.Time object
	return time.Unix(0, timestampMs*int64(time.Millisecond)), nil
}

type ConversationCollection struct {
	store *dirstore.MapDirectoryStore
}

func InitConversationCollection(convoCollection *ConversationCollection, baseDir string) error {
	partitionProvider := &dirstore.MonthBasedPartitionProvider{}
	store, err := dirstore.NewMapDirectoryStore(baseDir, true, dirstore.WithPartitionProvider(partitionProvider))
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

func (cc *ConversationCollection) SaveConversation(conversation *spec.Conversation) error {
	filename := cc.GetConversationFilename(*conversation)
	data, err := encdec.StructWithJSONTagsToMap(conversation)
	if err != nil {
		return err
	}
	return cc.store.SetFileData(filename, data)
}

func (cc *ConversationCollection) DeleteConversation(id, title string) error {
	return cc.store.DeleteFile(cc.getConversationFilenameUsingIDTitle(id, title))
}

func (cc *ConversationCollection) GetConversation(id, title string) (*spec.Conversation, error) {
	data, err := cc.store.GetFileData(cc.getConversationFilenameUsingIDTitle(id, title), false)
	if err != nil {
		return nil, err
	}
	convo := spec.Conversation{}
	err = encdec.MapToStructWithJSONTags(data, &convo)
	if err != nil {
		return nil, err
	}
	return &convo, nil
}

func (cc *ConversationCollection) ListConversations(token string) (*spec.ListResponse, error) {
	files, nextToken, err := cc.store.ListFiles("desc", token)
	if err != nil {
		return nil, err
	}

	var convoItems []spec.ConversationItem
	for _, file := range files {
		filename := filepath.Base(file)
		parts := strings.Split(strings.TrimSuffix(filename, ".json"), "_")
		id := parts[0]
		title := strings.Join(parts[1:], " ")
		d, err := GetDateFromUUIDv7(id)
		if err != nil {
			fmt.Printf("Could not get date from ID: %s, skipping", id)
			continue
		}
		convo := spec.ConversationItem{
			ID:        id,
			Title:     title,
			CreatedAt: d,
		}
		convoItems = append(convoItems, convo)
	}

	resp := &spec.ListResponse{}
	resp.ConversationItems = convoItems
	resp.NextPageToken = &nextToken
	return resp, nil
}

func (cc *ConversationCollection) AddMessageToConversation(id, title string, newMessage spec.ConversationMessage) error {
	conversation, err := cc.GetConversation(id, title)
	if err != nil {
		return err
	}

	conversation.Messages = append(conversation.Messages, newMessage)
	conversation.ModifiedAt = time.Now()
	return cc.SaveConversation(conversation)
}
