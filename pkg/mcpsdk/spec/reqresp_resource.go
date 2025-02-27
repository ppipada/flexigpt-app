package spec

import (
	"github.com/flexigpt/flexiui/pkg/mcpsdk/jsonrpc"
)

// Sent from the client to request a list of resources the server has.
type ListResourcesRequest jsonrpc.Request[*PaginatedRequestParams]

// The server's response to a resources/list request from the client.
type ListResourcesResponse jsonrpc.Response[*ListPromptsResult]

type ListResourcesResult struct {
	_ struct{} `json:"-"         additionalProperties:"true"`
	PaginatedResultParams
	Resources []Resource `json:"resources"`
}

// A known resource that the server is capable of reading.
type Resource struct {
	Annotations *Annotations `json:"annotations,omitempty"`

	// The URI of this resource.
	URI string `json:"uri"`

	// A human-readable name for this resource.
	//
	// This can be used by clients to populate UI elements.
	Name string `json:"name"`

	// A description of what this resource represents.
	//
	// This can be used by clients to improve the LLM's understanding of available
	// resources. It can be thought of like a "hint" to the model.
	Description *string `json:"description,omitempty"`

	// The MIME type of this resource, if known.
	MimeType *string `json:"mimeType,omitempty"`
}

// Sent from the client to the server, to read a specific resource URI.
type ReadResourceRequest jsonrpc.Request[ReadResourceRequestParams]

type ReadResourceRequestParams struct {
	_    struct{}               `json:"-"               additionalProperties:"true"`
	Meta map[string]any `json:"_meta,omitempty"`
	// The URI of the resource to read. The URI can use any protocol; it is up to the
	// server how to interpret it.
	URI string `json:"uri"`
}

// The server's response to a resources/read request from the client.
type (
	ReadResourceResponse jsonrpc.Response[*ReadResourceResult]
	ReadResourceResult   struct {
		// This result property is reserved by the protocol to allow clients and servers
		// to attach additional metadata to their responses.
		Meta     map[string]any `json:"_meta,omitempty"`
		Contents []ResourceContent      `json:"contents"`
	}
)

// Sent from the client to request a list of resource templates the server has.
type ListResourceTemplatesRequest jsonrpc.Request[*PaginatedRequestParams]

// The server's response to a resources/templates/list request from the client.
type (
	ListResourceTemplatesResponse jsonrpc.Response[*ListResourceTemplatesResult]
	ListResourceTemplatesResult   struct {
		// This result property is reserved by the protocol to allow clients and servers
		// to attach additional metadata to their responses.
		Meta map[string]any `json:"_meta,omitempty"`
		// ResourceTemplates corresponds to the JSON schema field "resourceTemplates".
		ResourceTemplates []ResourceTemplate `json:"resourceTemplates"`
	}
)

// Sent from the client to request resources/updated notifications from the server
// whenever a particular resource changes.
type SubscribeRequest jsonrpc.Request[SubscribeRequestParams]

type SubscribeRequestParams struct {
	// This result property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their responses.
	Meta map[string]any `json:"_meta,omitempty"`
	// The URI of the resource to subscribe to. The URI can use any protocol; it is up
	// to the server how to interpret it.
	URI string `json:"uri"`
}

// Sent from the client to request cancellation of resources/updated notifications
// from the server. This should follow a previous resources/subscribe request.
type UnsubscribeRequest jsonrpc.Request[UnsubscribeRequestParams]

type UnsubscribeRequestParams struct {
	// This result property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their responses.
	Meta map[string]any `json:"_meta,omitempty"`
	// The URI of the resource to unsubscribe from.
	URI string `json:"uri"`
}

// An optional notification from the server to the client, informing it that the
// list of resources it can read from has changed. This may be issued by servers
// without any previous subscription from the client.
type ResourceListChangedNotification jsonrpc.Notification[*AdditionalParams]

// A notification from the server to the client, informing it that a resource has
// changed and may need to be read again. This should only be sent if the client
// previously sent a resources/subscribe request.
type ResourceUpdatedNotification jsonrpc.Notification[ResourceUpdatedNotificationParams]

type ResourceUpdatedNotificationParams struct {
	// This result property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their responses.
	Meta map[string]any `json:"_meta,omitempty"`

	// The URI of the resource that has been updated. This might be a sub-resource of
	// the one that the client actually subscribed to.
	URI string `json:"uri"`
}
