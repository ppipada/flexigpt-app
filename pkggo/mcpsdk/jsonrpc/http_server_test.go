//go:build integration

package jsonrpc

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humacli"
)

// Example main server
type Options struct {
	Host  string `doc:"Host to listen on" default:"localhost"`
	Port  int    `doc:"Port to listen on" default:"8080"`
	Debug bool   `doc:"Enable debug logs" default:"false"`
}

func StartHTTPServer() {

	cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
		log.Printf("Options are %+v\n", opts)
		router := getRouter()
		server := http.Server{
			Addr:    fmt.Sprintf("%s:%d", opts.Host, opts.Port),
			Handler: router,
		}
		// Create the HTTP server.
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

	cli.Run()

}

// go test -v -tags=integration -run TestRunServer -count=1 ./pkggo/mcpsdk/jsonrpc/ &
func TestRunServer(t *testing.T) {
	// Start the server
	StartHTTPServer()
}
