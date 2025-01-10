package stdio

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/flexigpt/flexiui/pkggo/mcpsdk/jsonrpc"
	stdioNet "github.com/flexigpt/flexiui/pkggo/mcpsdk/transport/stdio/net"
)

type StdIOOptions struct {
	Debug bool `doc:"Enable debug logs" default:"false"`
}

func init() {
	// Redirect logs from the log package to stderr
	log.SetOutput(os.Stderr)
}

///////// PanicRecoveryMiddleware ////////////

// PanicRecoveryMiddleware is a middleware that recovers from panics in handlers
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic to stderr
				log.Printf("Recovered from panic: %v", err)

				// Optionally, log the stack trace
				log.Printf("%s", debug.Stack())

				// Return a 500 Internal Server Error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

///////// Middleware ////////////

func loggingMiddleware(ctx huma.Context, next func(huma.Context)) {
	log.Printf("Received request: %v %v", ctx.URL().RawPath, ctx.Operation().Path)
	next(ctx)
	log.Printf("Responded to request: %v %v", ctx.URL().RawPath, ctx.Operation().Path)
}

var DefaultJSONFormat = huma.Format{
	Marshal: func(w io.Writer, v any) error {
		return json.NewEncoder(w).Encode(v)
	},
	Unmarshal: func(data []byte, v any) error {
		return json.Unmarshal(data, v)
	},
}

func GetRouter() (http.Handler, huma.API) {
	router := http.NewServeMux()
	config := huma.DefaultConfig("Example JSONRPC API", "1.0.0")
	config.Formats = map[string]huma.Format{
		"application/json": DefaultJSONFormat,
		"json":             DefaultJSONFormat,
	}
	huma.NewError = jsonrpc.GetStatusError

	api := humago.New(router, config)
	api.UseMiddleware(loggingMiddleware)

	// Wrap the router with the panic recovery middleware
	handler := PanicRecoveryMiddleware(router)

	return handler, api
}

func GetJSONRPCStdIOServer(handler http.Handler) *stdioNet.Server {
	// Create the StdioConn
	connCreator := func() net.Conn {
		return stdioNet.NewStdioConn(os.Stdin, os.Stdout)
	}

	// Create the StdioListener
	listener := stdioNet.NewStdioListener(connCreator)

	// Create the MessageFramer
	framer := &stdioNet.LineFramer{}

	// Create the MessageHandler
	requestParams := RequestParams{
		Method: "POST",
		URL:    "/jsonrpc",
		Header: make(http.Header),
	}
	messageHandler := NewHTTPMessageHandler(handler, requestParams)
	server := stdioNet.NewServer(listener, framer, messageHandler)
	return server
}

func GetStdIOServerCLI() humacli.CLI {
	cli := humacli.New(func(hooks humacli.Hooks, opts *StdIOOptions) {
		log.Printf("Options are %+v", opts)

		handler, api := GetRouter()
		InitJSONRPChandlers(api)

		// Create the server with the handler and request parameters
		server := GetJSONRPCStdIOServer(handler)

		// Start the server
		hooks.OnStart(func() {
			if err := server.Serve(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
		})

		hooks.OnStop(func() {
			// Gracefully shutdown your server here
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Shutdown(ctx)
		})
	})

	return cli
}

func GetAndRunStdIOCLI() {
	cli := GetStdIOServerCLI()
	cli.Run()
}
