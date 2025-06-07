package spec

type SaveConversationRequest struct {
	Body *Conversation
}

type SaveConversationResponse struct{}

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
	// default is 10
	PageSize int `query:"pageSize"`
}

type SearchConversationsResponse struct {
	Body *SearchConversationsResponseBody
}

type SearchConversationsResponseBody struct {
	ConversationItems []ConversationItem `json:"conversationItems"`
	NextPageToken     *string            `json:"nextPageToken"`
}

type AddMessageToConversationRequest struct {
	ID   string `path:"id" required:"true"`
	Body *AddMessageToConversationRequestBody
}

type AddMessageToConversationRequestBody struct {
	Title      string              `json:"title"      required:"true"`
	NewMessage ConversationMessage `json:"newMessage" required:"true"`
}

type AddMessageToConversationResponse struct{}
