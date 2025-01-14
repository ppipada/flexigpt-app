package stdio

import (
	"io"
	"net/http"

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

// For actual runs os.Stdin, os.Stdout can be passed as reader and writer respectively
func GetServer(r io.Reader, w io.Writer, handler http.Handler) *stdioNet.Server {
	// Create the MessageFramer
	framer := &stdioNet.LineFramer{}

	// Create the MessageHandler, below is a jsonrpc packet on stdio to http adapter
	requestParams := RequestParams{
		Method: http.MethodPost,
		URL:    JSONRPCEndpoint,
		Header: make(http.Header),
	}
	messageHandler := NewHTTPMessageHandler(handler, requestParams)
	stdconn := stdioNet.NewStdioConn(r, w)
	server := stdioNet.NewServer(stdconn, framer, messageHandler)
	return server
}

// For actual runs os.Stdout and os.Stdin can be passed as reader and writer respectively
func GetClient(r io.Reader, w io.Writer) *stdioNet.Client {
	framer := &stdioNet.LineFramer{}

	clientConn := stdioNet.NewStdioConn(r, w)
	client := stdioNet.NewClient(clientConn, framer)

	return client
}
