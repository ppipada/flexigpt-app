package middleware

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

// WithRecoveryResp is a helper that recovers from any panic, logs the stack trace,
// and returns an error to the caller. T must match the response type of your function.
func WithRecoveryResp[T any](fn func() (T, error)) (T, error) {
	var (
		result T
		err    error
	)

	defer func() {
		if r := recover(); r != nil {
			// Log the panic plus stack trace.
			slog.Error("Panic recovered",
				slog.Any("panic", r),
				slog.String("stacktrace", string(debug.Stack())),
			)

			// Overwrite err so the caller sees we failed.
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()

	result, err = fn()
	if err != nil {
		slog.Error("Response", "Error", err.Error())
		slog.Error(string(debug.Stack()))
	}
	return result, err
}
