package httponly

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/flexigpt/flexiui/pkg/mcpsdk/transport/helpers_test"
)

// CLI options can be added as needed.
type Options struct {
	Host  string `doc:"Host to listen on" default:"localhost"`
	Port  int    `doc:"Port to listen on" default:"8080"`
	Debug bool   `doc:"Enable debug logs" default:"false"`
}

func SetupHTTPOnlyTransport() http.Handler {
	// Use default go router
	router := http.NewServeMux()

	api := humago.New(router, huma.DefaultConfig("Example JSONRPC API", "1.0.0"))
	// Add any middlewares
	api.UseMiddleware(helpers_test.LoggingMiddleware)
	handler := helpers_test.PanicRecoveryMiddleware(router)

	// Init the servers method and notifications handlers
	methodMap := helpers_test.GetMethodHandlers()
	notificationMap := helpers_test.GetNotificationHandlers()

	Register(api, methodMap, notificationMap)
	return handler
}

func GetHTTPServerCLI() humacli.CLI {
	cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
		log.Printf("Options are %+v\n", opts)
		handler := SetupHTTPOnlyTransport()
		// Initialize the http server
		server := http.Server{
			Addr:              fmt.Sprintf("%s:%d", opts.Host, opts.Port),
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		}

		// Hook the HTTP server.
		hooks.OnStart(func() {
			if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
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
