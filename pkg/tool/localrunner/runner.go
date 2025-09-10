package localrunner

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// GoToolRunner invokes a registered Go tool by name.
type GoToolRunner struct {
	funcName string
	reg      GoRegistry
	timeout  time.Duration
}

// Functional options for GoToolRunner.
type GoOption func(*GoToolRunner)

// WithGoTimeout sets a hard timeout for the Go tool invocation.
func WithGoTimeout(d time.Duration) GoOption {
	return func(r *GoToolRunner) {
		r.timeout = d
	}
}

func NewGoToolRunner(funcName string, reg GoRegistry, opts ...GoOption) (*GoToolRunner, error) {
	if funcName == "" {
		return nil, errors.New("go tool func name is empty")
	}
	if reg == nil {
		reg = globalGoRegistry
	}
	// Validate it exists at construction-time.
	if _, ok := reg.Lookup(funcName); !ok {
		return nil, fmt.Errorf("go tool func not found: %s", funcName)
	}
	r := &GoToolRunner{funcName: funcName, reg: reg}
	for _, o := range opts {
		o(r)
	}
	return r, nil
}

// Run executes the Go tool function with an optional timeout.
func (r *GoToolRunner) Run(
	ctx context.Context,
	args map[string]any,
) (output any, metaData map[string]any, err error) {
	fn, ok := r.reg.Lookup(r.funcName)
	if !ok {
		return nil, nil, fmt.Errorf("go tool func not found at runtime: %s", r.funcName)
	}

	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	}

	output, err = fn(ctx, args)
	metaData = map[string]any{
		"type":     "go",
		"funcName": r.funcName,
	}
	return output, metaData, err
}
