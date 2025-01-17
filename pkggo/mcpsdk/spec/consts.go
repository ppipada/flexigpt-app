package spec

const (
	LatestProtocolVersion = "2024-11-05"

	// BaseMethods.
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

	MethodCompletion = "completion/complete"

	MethodRootsList                     = "roots/list"
	MethodRootsNotificationsListChanged = "notifications/roots/list_changed"
)
