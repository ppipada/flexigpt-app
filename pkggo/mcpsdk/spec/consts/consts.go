package consts

const (
	LATEST_PROTOCOL_VERSION = "2024-11-05"
	JSONRPC_VERSION         = "2.0"

	// Standard JSON-RPC error codes
	PARSE_ERROR      = -32700
	INVALID_REQUEST  = -32600
	METHOD_NOT_FOUND = -32601
	INVALID_PARAMS   = -32602
	INTERNAL_ERROR   = -32603

	// BaseMethods
	MethodInitialize = "initialize"
	MethodPing       = "ping"

	MethodNotificationsCancelled   = "notifications/cancelled"
	MethodNotificationsInitialized = "notifications/initialized"
	MethodNotificationsProgress    = "notifications/progress"

	MethodResourcesList                     = "resources/list"
	MethodResourcesTemplatesList            = "resources/templates/list"
	MethodResourcesRead                     = "resources/read"
	MethodResourcesSubscribe                = "resources/subscribe"
	MethodResourcesUnsubscribe              = "resources/unsubscribe"
	MethodResourcesNotificationsListChanged = "notifications/resources/list_changed"
	MethodResourcesNotificationsUpdated     = "notifications/resources/updated"

	MethodPromptsList                     = "prompts/list"
	MethodPromptsGet                      = "prompts/get"
	MethodPromptsNotificationsListChanged = "notifications/prompts/list_changed"

	MethodToolsList                     = "tools/list"
	MethodToolsCall                     = "tools/call"
	MethodToolsNotificationsListChanged = "notifications/tools/list_changed"

	MethodLoggingSetLevel             = "logging/setLevel"
	MethodLoggingNotificationsMessage = "notifications/message"

	MethodSamplingCreateMessage = "sampling/createMessage"
	MethodSamplingCompletion    = "completion/complete"

	MethodRootsList                     = "roots/list"
	MethodRootsNotificationsListChanged = "notifications/roots/list_changed"
)
