package httpsse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/flexigpt/flexiui/pkggo/mcpsdk/jsonrpc"
)

type Options struct {
	Host  string `doc:"Host to listen on" default:"localhost"`
	Port  int    `doc:"Port to listen on" default:"8080"`
	Debug bool   `doc:"Enable debug logs" default:"false"`
}

func loggingMiddleware(ctx huma.Context, next func(huma.Context)) {
	// log.Printf("Received request: %v %v", ctx.URL().RawPath, ctx.Operation().Path)
	next(ctx)
	// log.Printf("Responded to request: %v %v", ctx.URL().RawPath, ctx.Operation().Path)
}

var DefaultJSONFormat = huma.Format{
	Marshal: func(w io.Writer, v any) error {
		return json.NewEncoder(w).Encode(v)
	},
	Unmarshal: func(data []byte, v any) error {
		// log.Printf("Trying to unmarshal %v", string(data))
		err := json.Unmarshal(data, v)
		// log.Printf("err %v", err)
		return err
	},
}

func GetRouter() (*http.ServeMux, huma.API) {
	router := http.NewServeMux()
	config := huma.DefaultConfig("Example JSONRPC API", "1.0.0")
	config.Formats = map[string]huma.Format{
		"application/json": DefaultJSONFormat,
		"json":             DefaultJSONFormat,
	}
	huma.NewError = jsonrpc.GetStatusError

	api := humago.New(router, config)
	api.UseMiddleware(loggingMiddleware)

	return router, api
}

func GetHTTPServerCLI() humacli.CLI {

	cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
		log.Printf("Options are %+v\n", opts)
		router, api := GetRouter()
		InitJSONRPChandlers(api)
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

	return cli
}

func StartHTTPServer() {
	cli := GetHTTPServerCLI()
	cli.Run()
}
