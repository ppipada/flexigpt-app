package localregistry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
)

// JSONToolFunc is the low-level function signature stored in the registry.
type JSONToolFunc func(ctx context.Context, in json.RawMessage) (json.RawMessage, error)

// GoRegistry provides lookup/register for Go tools by name, with json.RawMessage I/O.
type GoRegistry struct {
	mu      sync.RWMutex
	toolMap map[string]JSONToolFunc
	timeout time.Duration
}
type GoRegistryOption func(*GoRegistry) error

// DefaultGoRegistry is a package-level global registry with a 5s timeout.
// It is created during package initialization and panics on failure.
var DefaultGoRegistry = mustNewGoRegistry(WithCallTimeout(5 * time.Second))

// mustNewGoRegistry panics if NewGoRegistry fails.
// This is useful for package-level initialization.
func mustNewGoRegistry(opts ...GoRegistryOption) *GoRegistry {
	r, err := NewGoRegistry(opts...)
	if err != nil {
		panic(fmt.Errorf("localregistry: failed to create registry: %w", err))
	}
	return r
}

func WithCallTimeout(d time.Duration) GoRegistryOption {
	return func(gr *GoRegistry) error {
		gr.timeout = d
		return nil
	}
}

func NewGoRegistry(opts ...GoRegistryOption) (*GoRegistry, error) {
	r := &GoRegistry{toolMap: make(map[string]JSONToolFunc)}
	for _, o := range opts {
		if err := o(r); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func RegisterTyped[T any, R any](
	r *GoRegistry,
	name string,
	fn func(context.Context, T) (R, error),
) error {
	if name == "" || fn == nil {
		return errors.New("invalid registration: empty name or nil func")
	}
	wrapped := typedToJSON(fn)
	return r.Register(name, wrapped)
}

func (r *GoRegistry) Register(name string, fn JSONToolFunc) error {
	if name == "" || fn == nil {
		return errors.New("invalid registration: empty name or nil func")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.toolMap[name]; exists {
		return fmt.Errorf("go-tool already registered: %s", name)
	}
	r.toolMap[name] = fn
	return nil
}

func (r *GoRegistry) Call(
	ctx context.Context,
	name string,
	in json.RawMessage,
) (json.RawMessage, error) {
	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	}

	fn, ok := r.Lookup(name)
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return fn(ctx, in)
}

func (r *GoRegistry) Lookup(name string) (JSONToolFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.toolMap[name]
	return fn, ok
}

// typedToJSON wraps a typed function (ctx, T) -> (R, error) into a JSONToolFunc
// that strictly decodes input into T and encodes output R to JSON.
func typedToJSON[T any, R any](fn func(context.Context, T) (R, error)) JSONToolFunc {
	return func(ctx context.Context, in json.RawMessage) (json.RawMessage, error) {
		// Decode input strictly into T (rejects unknown fields and trailing data).
		args, err := encdec.DecodeJSONRaw[T](in)
		if err != nil {
			return nil, fmt.Errorf("invalid input: %w", err)
		}

		out, err := fn(ctx, args)
		if err != nil {
			return nil, err
		}
		return encdec.EncodeToJSONRaw(out)
	}
}
