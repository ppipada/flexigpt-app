package example

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/flexigpt/flexiui/pkg/mcpsdk/transport/stdio"
)

type StdIOOptions struct {
	Debug bool `doc:"Enable debug logs" default:"false"`
}

func SetupStdIOTransport() http.Handler {
	// Use default go router
	router := http.NewServeMux()

	api := humago.New(router, huma.DefaultConfig("Example JSONRPC API", "1.0.0"))
	// Add any middlewares
	api.UseMiddleware(loggingMiddleware)
	handler := PanicRecoveryMiddleware(router)

	// Init the servers method and notifications handlers
	methodMap := GetMethodHandlers()
	notificationMap := GetNotificationHandlers()
	stdio.Register(api, methodMap, notificationMap)

	return handler
}

func GetStdIOServerCLI() humacli.CLI {
	// Redirect logs from the log package to stderr
	// This is necessary for stdio transport
	log.SetOutput(os.Stderr)

	cli := humacli.New(func(hooks humacli.Hooks, opts *StdIOOptions) {
		log.Printf("Options are %+v\n", opts)
		handler := SetupStdIOTransport()
		// Create the server with the handler and request parameters
		server := stdio.GetServer(os.Stdin, os.Stdout, handler)

		// Start the server
		hooks.OnStart(func() {
			if err := server.Serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
