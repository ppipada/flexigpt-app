package conversationstore_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flexigpt/flexiui/pkg/conversationstore"
	"github.com/flexigpt/flexiui/pkg/conversationstore/spec"
	"github.com/google/uuid"
)

func TestInitConversation(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		expectedTitle string
	}{
		{"Empty title", "", "New Conversation"},
		{"Short title", "Chat with AI", "Chat with AI"},
		{
			"Long title",
			"This is a very long title that should be truncated to fit within the maximum allowed length of 64 characters",
			"This is a very long title that should be truncated to fit within",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convo, err := conversationstore.InitConversation(tt.title)
			if err != nil {
				t.Errorf("Expected valid convo, got error: %v", err)
			}
			if convo.Title != tt.expectedTitle {
				t.Errorf("Expected title %s, got %s", tt.expectedTitle, convo.Title)
			}
			if len(convo.Title) > 64 {
				t.Errorf(
					"Expected title to be truncated to 64 characters, got %d",
					len(convo.Title),
				)
			}
			if _, err := uuid.Parse(convo.ID); err != nil {
				t.Errorf("Expected valid UUID, got error: %v", err)
			}
		})
	}
}

func getConvo(id, title string) spec.Conversation {
	c := spec.Conversation{}
	c.ID = id
	c.Title = title
	return c
}

func TestGetConversationFilename(t *testing.T) {
	cc := &conversationstore.ConversationCollection{}
	tests := []struct {
		name         string
		conversation spec.Conversation
		expected     string
	}{
		{"Simple title", getConvo("123", "Chat"), "123_Chat.json"},
		{"Title with special chars", getConvo("123", "Chat with AI!"), "123_Chat_with_AI_.json"},
		{"Empty title", getConvo("123", ""), "123_.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := cc.GetConversationFilename(tt.conversation)
			if filename != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, filename)
			}
		})
	}
}

func TestConversationCollection(t *testing.T) {
	baseDir := filepath.Join(os.TempDir(), "conversationstore_test")
	defer os.RemoveAll(baseDir)
	cc := &conversationstore.ConversationCollection{}
	err := conversationstore.InitConversationCollection(cc, baseDir)
	if err != nil {
		t.Fatalf("Failed to create conversation collection: %v", err)
	}

	t.Run("Save and Get Conversation", func(t *testing.T) {
		validConvo, err := conversationstore.InitConversation("Test Conversation")
		if err != nil {
			t.Errorf("Failed to init conversation: %v", err)
		}
		emptyConvo, err := conversationstore.InitConversation("")
		if err != nil {
			t.Errorf("Failed to init conversation: %v", err)
		}
		tests := []struct {
			name         string
			conversation spec.Conversation
			expectError  bool
		}{
			{"Valid conversation", *validConvo, false},
			{"Empty title", *emptyConvo, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx := t.Context()
				_, err := cc.SaveConversation(
					ctx,
					&spec.SaveConversationRequest{Body: &tt.conversation},
				)
				if (err != nil) != tt.expectError {
					t.Fatalf("Expected error: %v, got: %v", tt.expectError, err)
				}

				retrievedConvo, err := cc.GetConversation(
					ctx,
					&spec.GetConversationRequest{
						ID:    tt.conversation.ID,
						Title: tt.conversation.Title,
					},
				)
				if err != nil {
					t.Fatalf("Failed to get conversation: %v", err)
				}

				if retrievedConvo.Body.ID != tt.conversation.ID ||
					retrievedConvo.Body.Title != tt.conversation.Title {
					t.Errorf("Retrieved conversation does not match saved conversation")
				}
			})
		}
	})

	t.Run("Delete Conversation", func(t *testing.T) {
		ctx := t.Context()
		convo, err := conversationstore.InitConversation("To Be Deleted")
		if err != nil {
			t.Errorf("Failed to init conversation: %v", err)
		}
		_, err = cc.SaveConversation(ctx, &spec.SaveConversationRequest{Body: convo})
		if err != nil {
			t.Errorf("Failed to save conversation: %v", err)
		}

		_, err = cc.DeleteConversation(
			ctx,
			&spec.DeleteConversationRequest{ID: convo.ID, Title: convo.Title},
		)
		if err != nil {
			t.Errorf("Failed to delete conversation: %v", err)
		}

		_, err = cc.GetConversation(
			ctx,
			&spec.GetConversationRequest{ID: convo.ID, Title: convo.Title},
		)
		if err == nil {
			t.Errorf("Expected error when getting deleted conversation, got none")
		}
	})

	t.Run("Add Message to Conversation", func(t *testing.T) {
		ctx := t.Context()
		convo, err := conversationstore.InitConversation("Message Test")
		if err != nil {
			t.Errorf("Failed to init conversation: %v", err)
		}
		_, err = cc.SaveConversation(ctx, &spec.SaveConversationRequest{Body: convo})
		if err != nil {
			t.Errorf("Failed to save conversation: %v", err)
		}
		tests := []struct {
			name        string
			message     spec.ConversationMessage
			expectError bool
		}{
			{
				"Valid message",
				spec.ConversationMessage{
					ID:      "msg1",
					Role:    spec.ConversationRoleUser,
					Content: "Hello",
				},
				false,
			},
			{
				"Empty content",
				spec.ConversationMessage{ID: "msg2", Role: spec.ConversationRoleUser, Content: ""},
				false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx := t.Context()
				_, err := cc.AddMessageToConversation(
					ctx,
					&spec.AddMessageToConversationRequest{
						ID: convo.ID,
						Body: &spec.AddMessageToConversationRequestBody{
							Title:      convo.Title,
							NewMessage: tt.message,
						},
					},
				)
				if (err != nil) != tt.expectError {
					t.Fatalf("Expected error: %v, got: %v", tt.expectError, err)
				}

				retrievedConvo, err := cc.GetConversation(
					ctx,
					&spec.GetConversationRequest{ID: convo.ID, Title: convo.Title},
				)
				if err != nil {
					t.Fatalf("Failed to get conversation: %v", err)
				}

				if len(retrievedConvo.Body.Messages) == 0 ||
					retrievedConvo.Body.Messages[len(retrievedConvo.Body.Messages)-1].Content != tt.message.Content {
					t.Errorf("Message not added correctly to conversation")
				}
			})
		}
	})
}

func TestConversationCollectionListing(t *testing.T) {
	baseDir := filepath.Join(os.TempDir(), "conversationstore_test_list")
	defer os.RemoveAll(baseDir)

	cc := &conversationstore.ConversationCollection{}
	err := conversationstore.InitConversationCollection(cc, baseDir)
	if err != nil {
		t.Fatalf("Failed to create conversation collection: %v", err)
	}

	t.Run("List Conversations", func(t *testing.T) {
		ctx := t.Context()
		convo1, err := conversationstore.InitConversation("First Conversation")
		if err != nil {
			t.Errorf("Failed to init conversation: %v", err)
		}
		convo2, err := conversationstore.InitConversation("Second Conversation")
		if err != nil {
			t.Errorf("Failed to init conversation: %v", err)
		}
		_, err = cc.SaveConversation(ctx, &spec.SaveConversationRequest{Body: convo1})
		if err != nil {
			t.Errorf("Failed to save conversation: %v", err)
		}
		_, err = cc.SaveConversation(ctx, &spec.SaveConversationRequest{Body: convo2})
		if err != nil {
			t.Errorf("Failed to save conversation: %v", err)
		}
		convoItems, err := cc.ListConversations(ctx, nil)
		if err != nil {
			t.Fatalf("Failed to list conversations: %v", err)
		}

		if len(convoItems.Body.ConversationItems) != 2 {
			t.Errorf("Expected 2 conversations, got %d", len(convoItems.Body.ConversationItems))
		}
	})
}

func TestGetDateFromUUIDv7(t *testing.T) {
	tests := []struct {
		name        string
		uuid        string
		expectError bool
	}{
		{"Valid UUID", "018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e", false},
		{"Invalid UUID length", "123", true},
		{"Invalid UUID format", "invalid-uuid-format", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := conversationstore.GetDateFromUUIDv7(tt.uuid)
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}
		})
	}
}
