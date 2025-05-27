package spec

import (
	jsonrpcReqResp "github.com/ppipada/flexigpt-app/pkg/jsonrpc/reqresp"
)

// A ping, issued by either the server or the client, to check that the other party
// is still alive. The receiver must promptly respond, or else may be disconnected.
type PingRequest jsonrpcReqResp.Request[*AdditionalParams]

// A ping response should send out a empty result.
type PingResponse jsonrpcReqResp.Response[AdditionalParams]

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
type CancelledNotification jsonrpcReqResp.Notification[CancelledNotificationParams]

type CancelledNotificationParams struct {
	// This property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their requests.
	Meta map[string]any `json:"_meta,omitempty"`
	_    struct{}       `json:"-"               additionalProperties:"true"`

	// The ID of the request to cancel.
	//
	// This MUST correspond to the ID of a request previously issued in the same
	// direction.
	RequestID jsonrpcReqResp.RequestID `json:"requestId"`

	// An optional string describing the reason for the cancellation. This MAY be
	// logged or presented to the user.
	Reason *string `json:"reason,omitempty"`
}

// An out-of-band notification used to inform the receiver of a progress update for
// a long-running request.
type ProgressNotification jsonrpcReqResp.Notification[ProgressNotificationParams]

type ProgressNotificationParams struct {
	// This property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their requests.
	Meta map[string]any `json:"_meta,omitempty"`
	_    struct{}       `json:"-"               additionalProperties:"true"`

	// The progress thus far. This should increase every time progress is made, even
	// if the total is unknown.
	Progress float64 `json:"progress"`

	// The progress token which was given in the initial request, used to associate
	// this notification with the request that is proceeding.
	ProgressToken ProgressToken `json:"progressToken"`

	// Total number of items to process (or total progress required), if known.
	Total *float64 `json:"total,omitempty"`
}

// This notification is sent from the client to the server after initialization has finished.
type InitializedNotification jsonrpcReqResp.Notification[*AdditionalParams]

// This request is sent from the client to the server when it first connects,
// asking it to begin initialization.
type InitializeRequest jsonrpcReqResp.Request[InitializeRequestParams]

type InitializeRequestParams struct {
	// This property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their requests.
	Meta map[string]any `json:"_meta,omitempty"`
	_    struct{}       `json:"-"               additionalProperties:"true"`

	// The latest version of the Model Context Protocol that the client supports. The
	// client MAY decide to support older versions as well.
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ServerClientInfo   `json:"clientInfo"`
}

// Describes the name and version of an MCP implementation.
type ServerClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Capabilities a client may support. Known capabilities are defined here, in this
// schema, but this is not a closed set: any client can define its own, additional
// capabilities.
type ClientCapabilities struct {
	// Experimental, non-standard capabilities that the client supports.
	Experimental *ClientCapabilitiesExperimental `json:"experimental,omitempty"`
	Roots        *ClientCapabilitiesRoots        `json:"roots,omitempty"`
	Sampling     *ClientCapabilitiesSampling     `json:"sampling,omitempty"`
}

// Experimental, non-standard capabilities that the client supports.
type ClientCapabilitiesExperimental map[string]map[string]any

// Present if the client supports listing roots.
type ClientCapabilitiesRoots struct {
	// Whether the client supports notifications for changes to the roots list.
	ListChanged *bool `json:"listChanged,omitempty"`
}

// Present if the client supports sampling from an LLM.
type ClientCapabilitiesSampling map[string]any

// After receiving an initialize request from the client, the server sends this response.

type InitializeResponse jsonrpcReqResp.Response[*InitializeResult]

type InitializeResult struct {
	// This result property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their responses.
	Meta map[string]any `json:"_meta,omitempty"`
	// The version of the Model Context Protocol that the server wants to use. This
	// may not match the version that the client requested. If the client cannot
	// support this version, it MUST disconnect.
	ProtocolVersion string `json:"protocolVersion"`

	Capabilities ServerCapabilities `json:"capabilities"`

	// ServerInfo corresponds to the JSON schema field "serverInfo".
	ServerInfo ServerClientInfo `json:"serverInfo"`

	// Instructions describing how to use the server and its features.
	//
	// This can be used by clients to improve the LLM's understanding of available
	// tools, resources, etc. It can be thought of like a "hint" to the model. For
	// example, this information MAY be added to the system prompt.
	Instructions *string `json:"instructions,omitempty"`
}

// Capabilities that a server may support. Known capabilities are defined here, in
// this schema, but this is not a closed set: any server can define its own,
// additional capabilities.
type ServerCapabilities struct {
	// Experimental, non-standard capabilities that the server supports.
	Experimental *ServerCapabilitiesExperimental `json:"experimental,omitempty"`

	// Present if the server supports sending log messages to the client.
	Logging *ServerCapabilitiesLogging `json:"logging,omitempty"`

	// Present if the server offers any prompt templates.
	Prompts *ServerCapabilitiesPrompts `json:"prompts,omitempty"`

	// Present if the server offers any resources to read.
	Resources *ServerCapabilitiesResources `json:"resources,omitempty"`

	// Present if the server offers any tools to call.
	Tools *ServerCapabilitiesTools `json:"tools,omitempty"`
}

// Experimental, non-standard capabilities that the server supports.
type ServerCapabilitiesExperimental map[string]map[string]any

// Present if the server supports sending log messages to the client.
type ServerCapabilitiesLogging map[string]any

// Present if the server offers any prompt templates.
type ServerCapabilitiesPrompts struct {
	// Whether this server supports notifications for changes to the prompt list.
	ListChanged *bool `json:"listChanged,omitempty"`
}

// Present if the server offers any resources to read.
type ServerCapabilitiesResources struct {
	// Whether this server supports notifications for changes to the resource list.
	ListChanged *bool `json:"listChanged,omitempty"`

	// Whether this server supports subscribing to resource updates.
	Subscribe *bool `json:"subscribe,omitempty"`
}

// Present if the server offers any tools to call.
type ServerCapabilitiesTools struct {
	// Whether this server supports notifications for changes to the tool list.
	ListChanged *bool `json:"listChanged,omitempty"`
}
