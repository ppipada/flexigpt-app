package spec

import "github.com/flexigpt/flexiui/pkg/mcpsdk/jsonrpc"

// A progress token, used to associate progress notifications with the original
// request.
type ProgressToken jsonrpc.IntString

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
