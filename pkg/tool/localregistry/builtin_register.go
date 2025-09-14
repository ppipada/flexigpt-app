package localregistry

import (
	"fmt"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/builtin/gotool"
)

// DefaultGoRegistry is a package-level global registry with a 5s timeout.
// It is created during package initialization and panics on failure.
var DefaultGoRegistry *GoRegistry

// mustNewGoRegistry panics if NewGoRegistry fails.
// This is useful for package-level initialization.
func mustNewGoRegistry(opts ...GoRegistryOption) *GoRegistry {
	r, err := NewGoRegistry(opts...)
	if err != nil {
		panic(fmt.Errorf("localregistry: failed to create registry: %w", err))
	}
	return r
}

func init() {
	DefaultGoRegistry = mustNewGoRegistry(WithCallTimeout(5 * time.Second))
	if err := RegisterTyped(DefaultGoRegistry, gotool.ReadFileFuncID, gotool.ReadFile); err != nil {
		panic(err)
	}
	if err := RegisterTyped(DefaultGoRegistry, gotool.ListDirectoryFuncID, gotool.ListDirectory); err != nil {
		panic(err)
	}
	if err := RegisterTyped(DefaultGoRegistry, gotool.SearchFilesFuncID, gotool.SearchFiles); err != nil {
		panic(err)
	}
}
