package spec

import "github.com/flexigpt/flexiui/pkggo/mcpsdk/spec/customtype"

// A progress token, used to associate progress notifications with the original
// request.
type ProgressToken customtype.IntString

// An opaque token used to represent a cursor for pagination.
type Cursor string

type Request struct {
	_ struct{} `json:"-"      additionalProperties:"true"`
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method"                             yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params *customtype.AdditionalParams `json:"params,omitempty" yaml:"params,omitempty" mapstructure:"params,omitempty"`
}

type Notification struct {
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params *customtype.AdditionalParams `json:"params,omitempty" yaml:"params,omitempty" mapstructure:"params,omitempty"`
}

type Result customtype.AdditionalParams

// A uniquely identifying ID for a request in JSON-RPC.
type RequestId customtype.IntString

// A request that expects a response.
type JSONRPCRequest struct {
	// Id corresponds to the JSON schema field "id".
	Id RequestId `json:"id" yaml:"id" mapstructure:"id"`

	// Jsonrpc corresponds to the JSON schema field "jsonrpc".
	Jsonrpc string `json:"jsonrpc" yaml:"jsonrpc" mapstructure:"jsonrpc"`

	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params *customtype.AdditionalParams `json:"params,omitempty" yaml:"params,omitempty" mapstructure:"params,omitempty"`
}

// A notification which does not expect a response.
type JSONRPCNotification struct {
	// Jsonrpc corresponds to the JSON schema field "jsonrpc".
	Jsonrpc string `json:"jsonrpc" yaml:"jsonrpc" mapstructure:"jsonrpc"`

	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params *customtype.AdditionalParams `json:"params,omitempty" yaml:"params,omitempty" mapstructure:"params,omitempty"`
}

// A successful (non-error) response to a request.
type JSONRPCResponse struct {
	// Id corresponds to the JSON schema field "id".
	Id RequestId `json:"id" yaml:"id" mapstructure:"id"`

	// Jsonrpc corresponds to the JSON schema field "jsonrpc".
	Jsonrpc string `json:"jsonrpc" yaml:"jsonrpc" mapstructure:"jsonrpc"`

	// Result corresponds to the JSON schema field "result".
	Result Result `json:"result" yaml:"result" mapstructure:"result"`
}

// A response to a request that indicates an error occurred.
type JSONRPCError struct {
	// Id corresponds to the JSON schema field "id".
	Id RequestId `json:"id" yaml:"id" mapstructure:"id"`

	// Jsonrpc corresponds to the JSON schema field "jsonrpc".
	Jsonrpc string `json:"jsonrpc" yaml:"jsonrpc" mapstructure:"jsonrpc"`

	// Error corresponds to the JSON schema field "error".
	Error JSONRPCErrorError `json:"error" yaml:"error" mapstructure:"error"`
}

type JSONRPCErrorError struct {
	// The error type that occurred.
	Code int `json:"code" yaml:"code" mapstructure:"code"`

	// A short description of the error. The message SHOULD be limited to a concise
	// single sentence.
	Message string `json:"message" yaml:"message" mapstructure:"message"`

	// Additional information about the error. The value of this member is defined by
	// the sender (e.g. detailed error information, nested errors etc.).
	Data interface{} `json:"data,omitempty" yaml:"data,omitempty" mapstructure:"data,omitempty"`
}

type EmptyResult = Result

type PaginatedRequest struct {
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params *PaginatedRequestParams `json:"params,omitempty" yaml:"params,omitempty" mapstructure:"params,omitempty"`
}

type PaginatedRequestParams struct {
	_ struct{} `json:"-"                additionalProperties:"true"`
	// An opaque token representing the current pagination position.
	// If provided, the server should return results starting after this cursor.
	Cursor *Cursor `json:"cursor,omitempty"                             yaml:"cursor,omitempty" mapstructure:"cursor,omitempty"`
}

type PaginatedResult struct {
	// This result property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their responses.
	Meta map[string]interface{} `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`

	// An opaque token representing the pagination position after the last returned
	// result.
	// If present, there may be more results available.
	NextCursor *Cursor `json:"nextCursor,omitempty" yaml:"nextCursor,omitempty" mapstructure:"nextCursor,omitempty"`
}

type Annotations struct {
	// Describes who the intended customer of this object or data is.
	//
	// It can include multiple entries to indicate content useful for multiple
	// audiences (e.g., `["user", "assistant"]`).
	Audience []Role `json:"audience,omitempty" yaml:"audience,omitempty" mapstructure:"audience,omitempty"`

	// Describes how important this data is for operating the server.
	//
	// A value of 1 means "most important," and indicates that the data is
	// effectively required, while 0 means "least important," and indicates that
	// the data is entirely optional.
	Priority *float64 `json:"priority,omitempty" yaml:"priority,omitempty" mapstructure:"priority,omitempty"`
}

type Content struct {
	// Type corresponds to the JSON schema field "type".
	Type ContentType `json:"type"                  yaml:"type"                  mapstructure:"type"`
	// Annotations corresponds to the JSON schema field "annotations".
	Annotations *Annotations `json:"annotations,omitempty" yaml:"annotations,omitempty" mapstructure:"annotations,omitempty"`
	// TextContent only: The text content of the message.
	Text *string `json:"text"                  yaml:"text"                  mapstructure:"text"`

	// ImageContent only: The base64-encoded image data.
	Data *string `json:"data" yaml:"data" mapstructure:"data"`

	// ImageContent only: The MIME type of the image. Different providers may support different image types.
	MimeType *string `json:"mimeType" yaml:"mimeType" mapstructure:"mimeType"`

	// EmbeddedResourceContent only: Resource corresponds to the JSON schema field "resource".
	Resource *ResourceContent `json:"resource" yaml:"resource" mapstructure:"resource"`
}

type ResourceContent struct {
	// The MIME type of this resource, if known.
	MimeType *string `json:"mimeType,omitempty" yaml:"mimeType,omitempty" mapstructure:"mimeType,omitempty"`

	// The text of the item. This must only be set if the item can actually be
	// represented as text (not binary data).
	Text *string `json:"text,omitempty" yaml:"text" mapstructure:"text"`

	// A base64-encoded string representing the binary data of the item.
	Blob *string `json:"blob,omitempty" yaml:"blob" mapstructure:"blob"`

	// The URI of this resource.
	Uri string `json:"uri" yaml:"uri" mapstructure:"uri"`
}
