package conversationstore

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

const tag = "Conversations"
const pathPrefix = "/conversations"

func InitConversationStoreHandlers(api huma.API, conversationStoreAPI *ConversationCollection) {

	huma.Register(api, huma.Operation{
		OperationID: "save-conversation",
		Method:      http.MethodPut,
		Path:        pathPrefix + "/{id}",
		Summary:     "Save a conversation",
		Description: "Save a conversation",
		Tags:        []string{tag},
	}, conversationStoreAPI.SaveConversation)

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

	huma.Register(api, huma.Operation{
		OperationID: "add-message-to-conversation",
		Method:      http.MethodPatch,
		Path:        pathPrefix + "/{id}",
		Summary:     "Append a message to a conversation",
		Description: "Append a message to a conversation",
		Tags:        []string{tag},
	}, conversationStoreAPI.AddMessageToConversation)

	huma.Register(api, huma.Operation{
		OperationID: "list-conversations",
		Method:      http.MethodGet,
		Path:        pathPrefix,
		Summary:     "List conversations",
		Description: "List conversation",
		Tags:        []string{tag},
	}, conversationStoreAPI.ListConversations)

}
