package example

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/flexigpt/flexiui/pkggo/mcpsdk/transport/httpsse"
)

// CLI options can be added as needed
type Options struct {
	Host  string `doc:"Host to listen on" default:"localhost"`
	Port  int    `doc:"Port to listen on" default:"8080"`
	Debug bool   `doc:"Enable debug logs" default:"false"`
}

func SetupSSETransport() http.Handler {
	// Use default go router
	router := http.NewServeMux()

	api := humago.New(router, huma.DefaultConfig("Example JSONRPC API", "1.0.0"))
	// Add any middlewares
	api.UseMiddleware(loggingMiddleware)
	handler := PanicRecoveryMiddleware(router)

	// Init the servers method and notifications handlers
	methodMap := GetMethodHandlers()
	notificationMap := GetNotificationHandlers()

	// Register the SSE endpoint and post endpoint
	sseTransport := httpsse.NewSSETransport(httpsse.JSONRPCEndpoint)
	sseTransport.Register(api, methodMap, notificationMap)
	return handler
}

func GetHTTPServerCLI() humacli.CLI {

	cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
		log.Printf("Options are %+v\n", opts)
		handler := SetupSSETransport()
		// Initialize the http server
		server := http.Server{
			Addr:    fmt.Sprintf("%s:%d", opts.Host, opts.Port),
			Handler: handler,
		}

		// Hook the HTTP server.
		hooks.OnStart(func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

func StartHTTPServer() {
	cli := GetHTTPServerCLI()
	cli.Run()
}
