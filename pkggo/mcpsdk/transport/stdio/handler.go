package stdio

import (
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/flexigpt/flexiui/pkggo/mcpsdk/jsonrpc"
	stdioNet "github.com/flexigpt/flexiui/pkggo/mcpsdk/transport/stdio/net"
)

const JSONRPCEndpoint = "/jsonrpc"

func Register(api huma.API,
	methodMap map[string]jsonrpc.IMethodHandler,
	notificationMap map[string]jsonrpc.INotificationHandler) {

	// Get default operation
	op := jsonrpc.GetDefaultOperation()
	op.Path = JSONRPCEndpoint
	// Register the methods
	jsonrpc.Register(api, op, methodMap, notificationMap)
}

func GetServer(handler http.Handler) *stdioNet.Server {
	// Create the MessageFramer
	framer := &stdioNet.LineFramer{}

	// Create the MessageHandler, below is a jsonrpc packet on stdio to http adapter
	requestParams := RequestParams{
		Method: http.MethodPost,
		URL:    JSONRPCEndpoint,
		Header: make(http.Header),
	}
	messageHandler := NewHTTPMessageHandler(handler, requestParams)
	stdconn := stdioNet.NewStdioConn(os.Stdin, os.Stdout)
	server := stdioNet.NewServer(stdconn, framer, messageHandler)
	return server
}
