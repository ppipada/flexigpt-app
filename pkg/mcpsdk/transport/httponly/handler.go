package httponly

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/flexigpt/flexiui/pkg/mcpsdk/jsonrpc"
)

const (
	JSONRPCEndpoint = "/jsonrpc"
)

func Register(api huma.API,
	methodMap map[string]jsonrpc.IMethodHandler,
	notificationMap map[string]jsonrpc.INotificationHandler,
) {
	// Get default operation
	op := jsonrpc.GetDefaultOperation()
	op.Path = JSONRPCEndpoint
	// Register the methods
	jsonrpc.Register(api, op, methodMap, notificationMap, nil, nil)
}
