package localregistry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ppipada/flexigpt-app/internal/jsonutil"
	"github.com/ppipada/flexigpt-app/internal/tool/spec"
)

// ToolFunc is the low-level function signature stored in the registry.
// It receives JSON-encoded args and returns one or more tool-store outputs.
type ToolFunc func(ctx context.Context, in json.RawMessage) ([]spec.ToolStoreOutputUnion, error)

// GoRegistry provides lookup/register for Go tools by name, with json.RawMessage I/O.
type GoRegistry struct {
	mu      sync.RWMutex
	toolMap map[string]ToolFunc
	timeout time.Duration
}
type GoRegistryOption func(*GoRegistry) error

func WithCallTimeout(d time.Duration) GoRegistryOption {
	return func(gr *GoRegistry) error {
		gr.timeout = d
		return nil
	}
}

func NewGoRegistry(opts ...GoRegistryOption) (*GoRegistry, error) {
	r := &GoRegistry{toolMap: make(map[string]ToolFunc)}
	for _, o := range opts {
		if err := o(r); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// RegisterOutputs registers a typed tool function that directly returns
// []ToolStoreOutputUnion.
//
//	fn: func(ctx, T) ([]spec.ToolStoreOutputUnion, error)
func RegisterOutputs[T any](
	r *GoRegistry,
	name string,
	fn func(context.Context, T) ([]spec.ToolStoreOutputUnion, error),
) error {
	if name == "" || fn == nil {
		return errors.New("invalid registration: empty name or nil func")
	}
	wrapped := typedToOutputs(fn)
	return r.Register(name, wrapped)
}

// RegisterTypedAsText registers a typed tool function whose output R is JSON-
// encodable. The JSON representation of R is wrapped into a single text block.
//
//	fn: func(ctx, T) (R, error)
func RegisterTypedAsText[T, R any](
	r *GoRegistry,
	name string,
	fn func(context.Context, T) (R, error),
) error {
	if name == "" || fn == nil {
		return errors.New("invalid registration: empty name or nil func")
	}
	wrapped := typedToText(fn)
	return r.Register(name, wrapped)
}

func (r *GoRegistry) Register(name string, fn ToolFunc) error {
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
) ([]spec.ToolStoreOutputUnion, error) {
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

func (r *GoRegistry) Lookup(name string) (ToolFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.toolMap[name]
	return fn, ok
}

// typedToOutputs wraps a typed function (ctx, T) -> ([]ToolStoreOutputUnion, error)
// into a ToolFunc that strictly decodes input into T.
func typedToOutputs[T any](
	fn func(context.Context, T) ([]spec.ToolStoreOutputUnion, error),
) ToolFunc {
	return func(ctx context.Context, in json.RawMessage) ([]spec.ToolStoreOutputUnion, error) {
		// Decode input strictly into T (rejects unknown fields and trailing data).
		args, err := jsonutil.DecodeJSONRaw[T](in)
		if err != nil {
			return nil, fmt.Errorf("invalid input: %w", err)
		}
		return fn(ctx, args)
	}
}

// typedToText wraps a typed function (ctx, T) -> (R, error) into a ToolFunc
// that JSON-encodes R and returns it as a single text output block.
func typedToText[T, R any](fn func(context.Context, T) (R, error)) ToolFunc {
	return func(ctx context.Context, in json.RawMessage) ([]spec.ToolStoreOutputUnion, error) {
		// Decode input strictly into T (rejects unknown fields and trailing data).
		args, err := jsonutil.DecodeJSONRaw[T](in)
		if err != nil {
			return nil, fmt.Errorf("invalid input: %w", err)
		}

		out, err := fn(ctx, args)
		if err != nil {
			return nil, err
		}
		raw, err := jsonutil.EncodeToJSONRaw(out)
		if err != nil {
			return nil, fmt.Errorf("encode output: %w", err)
		}

		text := string(raw)
		if text == "" || text == "null" {
			return nil, nil
		}
		return []spec.ToolStoreOutputUnion{
			{
				Kind: spec.ToolStoreOutputKindText,
				TextItem: &spec.ToolStoreOutputText{
					Text: text,
				},
			},
		}, nil
	}
}
