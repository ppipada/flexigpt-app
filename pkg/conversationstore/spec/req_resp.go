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
	Token string `query:"token"`
}

type ListConversationsResponse struct {
	Body *ListConversationsResponseBody
}

type ListConversationsResponseBody struct {
	ConversationItems []ConversationItem `json:"conversationItems"`
	NextPageToken     *string            `json:"nextPageToken"`
}

type SearchConversationsRequest struct {
	Query string `query:"query"    required:"true"`
	Token string `query:"token"`
	// Default is 10.
	PageSize int `query:"pageSize"`
}

type SearchConversationsResponse struct {
	Body *SearchConversationsResponseBody
}

type SearchConversationsResponseBody struct {
	ConversationItems []ConversationItem `json:"conversationItems"`
	NextPageToken     *string            `json:"nextPageToken"`
}
