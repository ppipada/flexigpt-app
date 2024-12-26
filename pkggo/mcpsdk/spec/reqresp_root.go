package spec

// Sent from the server to request a list of root URIs from the client. Roots allow
// servers to ask for specific directories or files to operate on. A common example
// for roots is providing a set of repositories or directories a server should
// operate
// on.
//
// This request is typically used when the server needs to understand the file
// system
// structure or access specific locations that the client has permission to read
// from.
type ListRootsRequest Request

// The client's response to a roots/list request from the server.
// This result contains an array of Root objects, each representing a root
// directory
// or file that the server can operate on.
type ListRootsResult struct {
	// This result property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their responses.
	Meta map[string]interface{} `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`

	// Roots corresponds to the JSON schema field "roots".
	Roots []Root `json:"roots" yaml:"roots" mapstructure:"roots"`
}

// Represents a root directory or file that the server can operate on.
type Root struct {
	// An optional name for the root. This can be used to provide a human-readable
	// identifier for the root, which may be useful for display purposes or for
	// referencing the root in other parts of the application.
	Name *string `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name,omitempty"`

	// The URI identifying the root. This *must* start with file:// for now.
	// This restriction may be relaxed in future versions of the protocol to allow
	// other URI schemes.
	Uri string `json:"uri" yaml:"uri" mapstructure:"uri"`
}

// A notification from the client to the server, informing it that the list of
// roots has changed.
// This notification should be sent whenever the client adds, removes, or modifies
// any root.
// The server should then request an updated list of roots using the
// ListRootsRequest.
type RootsListChangedNotification Notification
