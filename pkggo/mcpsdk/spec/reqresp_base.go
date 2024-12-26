package spec

// A ping, issued by either the server or the client, to check that the other party
// is still alive. The receiver must promptly respond, or else may be disconnected.
type PingRequest Request

// This request is sent from the client to the server when it first connects,
// asking it to begin initialization.
type InitializeRequest struct {
	_ struct{} `json:"-"      additionalProperties:"true"`
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method"                             yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params InitializeRequestParams `json:"params" yaml:"params" mapstructure:"params"`
}

type InitializeRequestParams struct {
	// Capabilities corresponds to the JSON schema field "capabilities".
	Capabilities ClientCapabilities `json:"capabilities" yaml:"capabilities" mapstructure:"capabilities"`

	// ClientInfo corresponds to the JSON schema field "clientInfo".
	ClientInfo Implementation `json:"clientInfo" yaml:"clientInfo" mapstructure:"clientInfo"`

	// The latest version of the Model Context Protocol that the client supports. The
	// client MAY decide to support older versions as well.
	ProtocolVersion string `json:"protocolVersion" yaml:"protocolVersion" mapstructure:"protocolVersion"`
}

// Capabilities a client may support. Known capabilities are defined here, in this
// schema, but this is not a closed set: any client can define its own, additional
// capabilities.
type ClientCapabilities struct {
	// Experimental, non-standard capabilities that the client supports.
	Experimental ClientCapabilitiesExperimental `json:"experimental,omitempty" yaml:"experimental,omitempty" mapstructure:"experimental,omitempty"`

	// Present if the client supports listing roots.
	Roots *ClientCapabilitiesRoots `json:"roots,omitempty" yaml:"roots,omitempty" mapstructure:"roots,omitempty"`

	// Present if the client supports sampling from an LLM.
	Sampling ClientCapabilitiesSampling `json:"sampling,omitempty" yaml:"sampling,omitempty" mapstructure:"sampling,omitempty"`
}

// Experimental, non-standard capabilities that the client supports.
type ClientCapabilitiesExperimental map[string]map[string]interface{}

// Present if the client supports listing roots.
type ClientCapabilitiesRoots struct {
	// Whether the client supports notifications for changes to the roots list.
	ListChanged *bool `json:"listChanged,omitempty" yaml:"listChanged,omitempty" mapstructure:"listChanged,omitempty"`
}

// Present if the client supports sampling from an LLM.
type ClientCapabilitiesSampling map[string]interface{}

// Describes the name and version of an MCP implementation.
type Implementation struct {
	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// Version corresponds to the JSON schema field "version".
	Version string `json:"version" yaml:"version" mapstructure:"version"`
}

// After receiving an initialize request from the client, the server sends this
// response.
type InitializeResult struct {
	_ struct{} `json:"-" additionalProperties:"true"`

	// This result property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their responses.
	Meta map[string]interface{} `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`

	// Capabilities corresponds to the JSON schema field "capabilities".
	Capabilities ServerCapabilities `json:"capabilities" yaml:"capabilities" mapstructure:"capabilities"`

	// Instructions describing how to use the server and its features.
	//
	// This can be used by clients to improve the LLM's understanding of available
	// tools, resources, etc. It can be thought of like a "hint" to the model. For
	// example, this information MAY be added to the system prompt.
	Instructions *string `json:"instructions,omitempty" yaml:"instructions,omitempty" mapstructure:"instructions,omitempty"`

	// The version of the Model Context Protocol that the server wants to use. This
	// may not match the version that the client requested. If the client cannot
	// support this version, it MUST disconnect.
	ProtocolVersion string `json:"protocolVersion" yaml:"protocolVersion" mapstructure:"protocolVersion"`

	// ServerInfo corresponds to the JSON schema field "serverInfo".
	ServerInfo Implementation `json:"serverInfo" yaml:"serverInfo" mapstructure:"serverInfo"`
}

// Capabilities that a server may support. Known capabilities are defined here, in
// this schema, but this is not a closed set: any server can define its own,
// additional capabilities.
type ServerCapabilities struct {
	// Experimental, non-standard capabilities that the server supports.
	Experimental ServerCapabilitiesExperimental `json:"experimental,omitempty" yaml:"experimental,omitempty" mapstructure:"experimental,omitempty"`

	// Present if the server supports sending log messages to the client.
	Logging ServerCapabilitiesLogging `json:"logging,omitempty" yaml:"logging,omitempty" mapstructure:"logging,omitempty"`

	// Present if the server offers any prompt templates.
	Prompts *ServerCapabilitiesPrompts `json:"prompts,omitempty" yaml:"prompts,omitempty" mapstructure:"prompts,omitempty"`

	// Present if the server offers any resources to read.
	Resources *ServerCapabilitiesResources `json:"resources,omitempty" yaml:"resources,omitempty" mapstructure:"resources,omitempty"`

	// Present if the server offers any tools to call.
	Tools *ServerCapabilitiesTools `json:"tools,omitempty" yaml:"tools,omitempty" mapstructure:"tools,omitempty"`
}

// Experimental, non-standard capabilities that the server supports.
type ServerCapabilitiesExperimental map[string]map[string]interface{}

// Present if the server supports sending log messages to the client.
type ServerCapabilitiesLogging map[string]interface{}

// Present if the server offers any prompt templates.
type ServerCapabilitiesPrompts struct {
	// Whether this server supports notifications for changes to the prompt list.
	ListChanged *bool `json:"listChanged,omitempty" yaml:"listChanged,omitempty" mapstructure:"listChanged,omitempty"`
}

// Present if the server offers any resources to read.
type ServerCapabilitiesResources struct {
	// Whether this server supports notifications for changes to the resource list.
	ListChanged *bool `json:"listChanged,omitempty" yaml:"listChanged,omitempty" mapstructure:"listChanged,omitempty"`

	// Whether this server supports subscribing to resource updates.
	Subscribe *bool `json:"subscribe,omitempty" yaml:"subscribe,omitempty" mapstructure:"subscribe,omitempty"`
}

// Present if the server offers any tools to call.
type ServerCapabilitiesTools struct {
	// Whether this server supports notifications for changes to the tool list.
	ListChanged *bool `json:"listChanged,omitempty" yaml:"listChanged,omitempty" mapstructure:"listChanged,omitempty"`
}

// This notification can be sent by either side to indicate that it is cancelling a
// previously-issued request.
//
// The request SHOULD still be in-flight, but due to communication latency, it is
// always possible that this notification MAY arrive after the request has already
// finished.
//
// This notification indicates that the result will be unused, so any associated
// processing SHOULD cease.
//
// A client MUST NOT attempt to cancel its `initialize` request.
type CancelledNotification struct {
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params CancelledNotificationParams `json:"params" yaml:"params" mapstructure:"params"`
}

type CancelledNotificationParams struct {
	// An optional string describing the reason for the cancellation. This MAY be
	// logged or presented to the user.
	Reason *string `json:"reason,omitempty" yaml:"reason,omitempty" mapstructure:"reason,omitempty"`

	// The ID of the request to cancel.
	//
	// This MUST correspond to the ID of a request previously issued in the same
	// direction.
	RequestId RequestId `json:"requestId" yaml:"requestId" mapstructure:"requestId"`
}

// This notification is sent from the client to the server after initialization has
// finished.
type InitializedNotification Request

// An out-of-band notification used to inform the receiver of a progress update for
// a long-running request.
type ProgressNotification struct {
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params ProgressNotificationParams `json:"params" yaml:"params" mapstructure:"params"`
}

type ProgressNotificationParams struct {
	// The progress thus far. This should increase every time progress is made, even
	// if the total is unknown.
	Progress float64 `json:"progress" yaml:"progress" mapstructure:"progress"`

	// The progress token which was given in the initial request, used to associate
	// this notification with the request that is proceeding.
	ProgressToken ProgressToken `json:"progressToken" yaml:"progressToken" mapstructure:"progressToken"`

	// Total number of items to process (or total progress required), if known.
	Total *float64 `json:"total,omitempty" yaml:"total,omitempty" mapstructure:"total,omitempty"`
}

// Unions in the schema
type ClientNotification interface{}
type ClientRequest interface{}
type ClientResult interface{}
type JSONRPCMessage interface{}
type ServerNotification interface{}
type ServerRequest interface{}
type ServerResult interface{}
