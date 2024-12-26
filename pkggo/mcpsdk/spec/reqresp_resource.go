package spec

// Sent from the client to request a list of resources the server has.
type ListResourcesRequest PaginatedRequest

// The server's response to a resources/list request from the client.
type ListResourcesResult struct {
	PaginatedResult

	// Resources corresponds to the JSON schema field "resources".
	Resources []Resource `json:"resources" yaml:"resources" mapstructure:"resources"`
}

// A known resource that the server is capable of reading.
type Resource struct {
	// Annotations corresponds to the JSON schema field "annotations".
	Annotations *Annotations `json:"annotations,omitempty" yaml:"annotations,omitempty" mapstructure:"annotations,omitempty"`

	// A description of what this resource represents.
	//
	// This can be used by clients to improve the LLM's understanding of available
	// resources. It can be thought of like a "hint" to the model.
	Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`

	// The MIME type of this resource, if known.
	MimeType *string `json:"mimeType,omitempty" yaml:"mimeType,omitempty" mapstructure:"mimeType,omitempty"`

	// A human-readable name for this resource.
	//
	// This can be used by clients to populate UI elements.
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// The URI of this resource.
	Uri string `json:"uri" yaml:"uri" mapstructure:"uri"`
}

// Sent from the client to request a list of resource templates the server has.
type ListResourceTemplatesRequest PaginatedRequest

// The server's response to a resources/templates/list request from the client.
type ListResourceTemplatesResult struct {
	PaginatedResult

	// ResourceTemplates corresponds to the JSON schema field "resourceTemplates".
	ResourceTemplates []ResourceTemplate `json:"resourceTemplates" yaml:"resourceTemplates" mapstructure:"resourceTemplates"`
}

// A template description for resources available on the server.
type ResourceTemplate struct {
	// Annotations corresponds to the JSON schema field "annotations".
	Annotations *Annotations `json:"annotations,omitempty" yaml:"annotations,omitempty" mapstructure:"annotations,omitempty"`

	// A description of what this template is for.
	//
	// This can be used by clients to improve the LLM's understanding of available
	// resources. It can be thought of like a "hint" to the model.
	Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`

	// The MIME type for all resources that match this template. This should only be
	// included if all resources matching this template have the same type.
	MimeType *string `json:"mimeType,omitempty" yaml:"mimeType,omitempty" mapstructure:"mimeType,omitempty"`

	// A human-readable name for the type of resource this template refers to.
	//
	// This can be used by clients to populate UI elements.
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// A URI template (according to RFC 6570) that can be used to construct resource
	// URIs.
	UriTemplate string `json:"uriTemplate" yaml:"uriTemplate" mapstructure:"uriTemplate"`
}

// Sent from the client to the server, to read a specific resource URI.
type ReadResourceRequest struct {
	_ struct{} `json:"-"      additionalProperties:"true"`
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method"                             yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params ReadResourceRequestParams `json:"params" yaml:"params" mapstructure:"params"`
}

type ReadResourceRequestParams struct {
	_ struct{} `json:"-"   additionalProperties:"true"`
	// The URI of the resource to read. The URI can use any protocol; it is up to the
	// server how to interpret it.
	Uri string `json:"uri"                             yaml:"uri" mapstructure:"uri"`
}

// The server's response to a resources/read request from the client.
type ReadResourceResult struct {
	// This result property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their responses.
	Meta map[string]interface{} `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`

	// Contents corresponds to the JSON schema field "contents".
	Contents []ResourceContent `json:"contents" yaml:"contents" mapstructure:"contents"`
}

// Sent from the client to request resources/updated notifications from the server
// whenever a particular resource changes.
type SubscribeRequest struct {
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params SubscribeRequestParams `json:"params" yaml:"params" mapstructure:"params"`
}

type SubscribeRequestParams struct {
	// The URI of the resource to subscribe to. The URI can use any protocol; it is up
	// to the server how to interpret it.
	Uri string `json:"uri" yaml:"uri" mapstructure:"uri"`
}

// Sent from the client to request cancellation of resources/updated notifications
// from the server. This should follow a previous resources/subscribe request.
type UnsubscribeRequest struct {
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params UnsubscribeRequestParams `json:"params" yaml:"params" mapstructure:"params"`
}

type UnsubscribeRequestParams struct {
	// The URI of the resource to unsubscribe from.
	Uri string `json:"uri" yaml:"uri" mapstructure:"uri"`
}

// An optional notification from the server to the client, informing it that the
// list of resources it can read from has changed. This may be issued by servers
// without any previous subscription from the client.
type ResourceListChangedNotification Notification

// A notification from the server to the client, informing it that a resource has
// changed and may need to be read again. This should only be sent if the client
// previously sent a resources/subscribe request.
type ResourceUpdatedNotification struct {
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params ResourceUpdatedNotificationParams `json:"params" yaml:"params" mapstructure:"params"`
}

type ResourceUpdatedNotificationParams struct {
	// The URI of the resource that has been updated. This might be a sub-resource of
	// the one that the client actually subscribed to.
	Uri string `json:"uri" yaml:"uri" mapstructure:"uri"`
}
