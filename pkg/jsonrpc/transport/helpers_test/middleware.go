package helpers_test

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/danielgtaylor/huma/v2"
)

// This is a huma middleware.
// Either a huma middleware can be added or a http handler middleware can be added.
func LoggingMiddleware(ctx huma.Context, next func(huma.Context)) {
	// Log.Printf("Received request: %v %v", ctx.URL().RawPath, ctx.Operation().Path).
	next(ctx)
	// Log.Printf("Responded to request: %v %v", ctx.URL().RawPath, ctx.Operation().Path).
}

// This is a http handler middleware.
// PanicRecoveryMiddleware recovers from panics in handlers.
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic to stderr.
				log.Printf("Recovered from panic: %+v", err)

				// Optionally, log the stack trace.
				log.Printf("%s", debug.Stack())

				// Return a 500 Internal Server Error.
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
