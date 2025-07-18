package spec

import "time"

type PutConversationRequest struct {
	ID   string `path:"id" required:"true"`
	Body *PutConversationRequestBody
}

type PutConversationRequestBody struct {
	Title      string                `json:"title"      required:"true"`
	CreatedAt  time.Time             `json:"createdAt"  required:"true"`
	ModifiedAt time.Time             `json:"modifiedAt" required:"true"`
	Messages   []ConversationMessage `json:"messages"   required:"true"`
}

type PutConversationResponse struct{}

type PutMessagesToConversationRequest struct {
	ID   string `path:"id" required:"true"`
	Body *PutMessagesToConversationRequestBody
}

type PutMessagesToConversationRequestBody struct {
	Title    string                `json:"title"    required:"true"`
	Messages []ConversationMessage `json:"messages" required:"true"`
}

type PutMessagesToConversationResponse struct{}

type DeleteConversationRequest struct {
	ID    string `path:"id" required:"true"`
	Title string `          required:"true" query:"title"`
}

type DeleteConversationResponse struct{}

type GetConversationRequest struct {
	ID    string `path:"id" required:"true"`
	Title string `          required:"true" query:"title"`
}

type GetConversationResponse struct {
	Body *Conversation
}

type ListConversationsRequest struct {
	PageSize  int    `query:"pageSize"`
	PageToken string `query:"pageToken"`
}

type ListConversationsResponse struct {
	Body *ListConversationsResponseBody
}

type ListConversationsResponseBody struct {
	ConversationListItems []ConversationListItem `json:"conversationListItems"`
	NextPageToken         *string                `json:"nextPageToken,omitempty"`
}

// ConversationListItem represents a conversation with basic details.
type ConversationListItem struct {
	ID             string     `json:"id"`
	SanatizedTitle string     `json:"sanatizedTitle"`
	ModifiedAt     *time.Time `json:"modifiedAt"`
}

type SearchConversationsRequest struct {
	Query     string `query:"q"         required:"true"`
	PageToken string `query:"pageToken"`
	PageSize  int    `query:"pageSize"`
}

type SearchConversationsResponse struct {
	Body *SearchConversationsResponseBody
}

type SearchConversationsResponseBody struct {
	ConversationListItems []ConversationListItem `json:"conversationListItems"`
	NextPageToken         *string                `json:"nextPageToken,omitempty"`
}
