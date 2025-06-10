package conversationstore

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

const (
	tag        = "Conversations"
	pathPrefix = "/conversations"
)

func InitConversationStoreHandlers(api huma.API, conversationStoreAPI *ConversationCollection) {
	huma.Register(api, huma.Operation{
		OperationID: "list-conversations",
		Method:      http.MethodGet,
		Path:        pathPrefix,
		Summary:     "List conversations",
		Description: "List conversation",
		Tags:        []string{tag},
	}, conversationStoreAPI.ListConversations)

	huma.Register(api, huma.Operation{
		OperationID: "search-conversations",
		Method:      http.MethodGet,
		Path:        pathPrefix + "/search",
		Summary:     "Search conversations",
		Description: "Search conversation",
		Tags:        []string{tag},
	}, conversationStoreAPI.SearchConversations)

	huma.Register(api, huma.Operation{
		OperationID: "put-conversation",
		Method:      http.MethodPut,
		Path:        pathPrefix + "/{id}",
		Summary:     "Put a conversation",
		Description: "Put a conversation",
		Tags:        []string{tag},
	}, conversationStoreAPI.PutConversation)

	huma.Register(api, huma.Operation{
		OperationID: "put-messages-to-conversation",
		Method:      http.MethodPut,
		Path:        pathPrefix + "/{id}/messages",
		Summary:     "Put messages to a conversation",
		Description: "Put messages to a conversation",
		Tags:        []string{tag},
	}, conversationStoreAPI.PutMessagesToConversation)

	huma.Register(api, huma.Operation{
		OperationID: "delete-conversation",
		Method:      http.MethodDelete,
		Path:        pathPrefix + "/{id}",
		Summary:     "Delete a conversation",
		Description: "Delete a conversation",
		Tags:        []string{tag},
	}, conversationStoreAPI.DeleteConversation)

	huma.Register(api, huma.Operation{
		OperationID: "get-conversation",
		Method:      http.MethodGet,
		Path:        pathPrefix + "/{id}",
		Summary:     "Get a conversation",
		Description: "Get a conversation",
		Tags:        []string{tag},
	}, conversationStoreAPI.GetConversation)
}
