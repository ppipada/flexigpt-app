package localrunner

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// GoToolFunc is the signature for registered Go tool functions.
// Implementers must be deterministic and thread-safe.
type (
	GoToolFunc func(ctx context.Context, args map[string]any) (any, error)
	// GoRegistry provides lookup/register for Go tools by name.
	GoRegistry interface {
		Lookup(name string) (GoToolFunc, bool)
		Register(name string, fn GoToolFunc) error
	}
)

type defaultGoRegistry struct {
	mu sync.RWMutex
	m  map[string]GoToolFunc
}

var globalGoRegistry = newDefaultGoRegistry()

func newDefaultGoRegistry() *defaultGoRegistry {
	return &defaultGoRegistry{m: make(map[string]GoToolFunc)}
}

// RegisterGoTool registers a GoToolFunc in the global registry.
func RegisterGoTool(name string, fn GoToolFunc) error {
	return globalGoRegistry.Register(name, fn)
}

// DefaultGoRegistry exposes the global registry instance.
func DefaultGoRegistry() GoRegistry {
	return globalGoRegistry
}

func (r *defaultGoRegistry) Lookup(name string) (GoToolFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.m[name]
	return fn, ok
}

func (r *defaultGoRegistry) Register(name string, fn GoToolFunc) error {
	if name == "" || fn == nil {
		return errors.New("invalid registration: empty name or nil func")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.m[name]; exists {
		return fmt.Errorf("go-tool already registered: %s", name)
	}
	r.m[name] = fn
	return nil
}
