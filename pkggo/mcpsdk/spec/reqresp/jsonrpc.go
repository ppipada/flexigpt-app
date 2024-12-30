package reqresp

import "github.com/flexigpt/flexiui/pkggo/mcpsdk/spec/types"

// A uniquely identifying ID for a request in JSON-RPC.
type RequestId types.IntString

// A progress token, used to associate progress notifications with the original
// request.
type ProgressToken types.IntString

// A request that expects a response.
type JSONRPCV2Request[T any] struct {
	JSONRPC string    `json:"jsonrpc"          enum:"2.0"`
	Id      RequestId `json:"id"`
	Method  string    `json:"method"`
	Params  T         `json:"params,omitempty"` // normally this could be *AdditionalParams
}

type JSONRPCV2Error struct {
	// The error type that occurred.
	Code int `json:"code"`

	// A short description of the error. The message SHOULD be limited to a concise
	// single sentence.
	Message string `json:"message"`

	// Additional information about the error. The value of this member is defined by
	// the sender (e.g. detailed error information, nested errors etc.).
	Data interface{} `json:"data,omitempty"`
}

// A successful (non-error) response to a request.
type JSONRPCV2Response[T any] struct {
	JSONRPC string          `json:"jsonrpc"          enum:"2.0"`
	Id      RequestId       `json:"id"`
	Result  T               `json:"result,omitempty"` // normally this could be *AdditionalParams
	Error   *JSONRPCV2Error `json:"error,omitempty"`
}

// A notification which does not expect a response.
type JSONRPCV2Notification[T any] struct {
	JSONRPC string `json:"jsonrpc"          enum:"2.0"`
	Method  string `json:"method"`
	Params  T      `json:"params,omitempty"` // normally this could be *AdditionalParams
}

// An opaque token used to represent a cursor for pagination.
type Cursor string

type PaginatedRequestParams struct {
	// This property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their requests.
	Meta map[string]interface{} `json:"_meta,omitempty"`
	_    struct{}               `json:"-"               additionalProperties:"true"`

	// An opaque token representing the current pagination position.
	// If provided, the server should return results starting after this cursor.
	Cursor *Cursor `json:"cursor,omitempty"`
}

type PaginatedResultParams struct {
	// This property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their requests.
	Meta map[string]interface{} `json:"_meta,omitempty"`
	_    struct{}               `json:"-"               additionalProperties:"true"`

	// An opaque token representing the pagination position after the last returned result.
	// If present, there may be more results available.
	NextCursor *Cursor `json:"nextCursor,omitempty"`
}
