// Thread-safe, reflection-free feature-toggle store.
//
//   • One JSON file keeps a map[string]bool per *group* ("bundles", "templates", ...).
//   • A *key type* decides at compile time
//         1) the group it belongs to   — Key.Group()
//         2) how the value is encoded  — Key.ID()
//   • Every key type must be registered via WithKeyType when the store
//     is created.  Anything else is rejected with a clear error.

package booloverlay

import (
	"fmt"
	"sync"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
)

// Key represents an identifier that can be toggled.
//
// Group() MUST return the *same* constant string for every value of the
// concrete type and MUST be implemented on a *value* receiver so that it
// can be called on the zero value in WithKeyType.
//
// ID() encodes the *value*; anything returning a unique string is fine.
type Key interface {
	Group() string
	ID() string
}

type Store struct {
	mu  sync.RWMutex
	db  *filestore.MapFileStore
	reg map[string]struct{}
}

type Option func(*Store) error

// WithKeyType registers the group for K and guarantees that the group
// exists in the file.
//
// Note: Group() **must** be defined on the *value* receiver of K.
func WithKeyType[K Key]() Option {
	var zero K // value receiver required
	group := zero.Group()

	return func(s *Store) error {
		if _, dup := s.reg[group]; dup {
			return fmt.Errorf("booloverlay: duplicate registration for group %q", group)
		}
		s.reg[group] = struct{}{}

		// Ensure the group exists on disk.
		s.mu.Lock()
		defer s.mu.Unlock()

		root, err := s.readRoot()
		if err != nil {
			return err
		}
		if _, ok := root[group]; !ok {
			root[group] = map[string]bool{}
			return s.writeRoot(root)
		}
		return nil
	}
}

// NewStore opens (or creates) the JSON file and applies all options.
func NewStore(filePath string, opts ...Option) (*Store, error) {
	fs, err := filestore.NewMapFileStore(
		filePath,
		map[string]any{},                      // default empty object
		filestore.WithCreateIfNotExists(true), // create when missing
		filestore.WithAutoFlush(true),         // flush on each write
		filestore.WithEncoderDecoder(encdec.JSONEncoderDecoder{}),
	)
	if err != nil {
		return nil, err
	}

	st := &Store{db: fs, reg: make(map[string]struct{})}
	for _, opt := range opts {
		if err := opt(st); err != nil {
			return nil, err
		}
	}
	return st, nil
}

// IsEnabled returns the stored flag for k, or def when missing / on error.
func (s *Store) IsEnabled(k Key, def bool) (bool, error) {
	group := k.Group()
	if err := s.ensureRegistered(group); err != nil {
		return def, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isEnabled(group, k.ID(), def)
}

// SetEnabled writes k ← enabled (creates the entry, group already exists).
func (s *Store) SetEnabled(k Key, enabled bool) error {
	group := k.Group()
	if err := s.ensureRegistered(group); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.setEnabled(group, k.ID(), enabled)
}

// Delete removes k from the store (NOP if it never existed).
func (s *Store) Delete(k Key) error {
	group := k.Group()
	if err := s.ensureRegistered(group); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deleteKey(group, k.ID())
}

func (s *Store) ensureRegistered(group string) error {
	if _, ok := s.reg[group]; !ok {
		return fmt.Errorf("booloverlay: group %q is not registered", group)
	}
	return nil
}

func (s *Store) readRoot() (map[string]map[string]bool, error) {
	raw, err := s.db.GetAll(false)
	if err != nil {
		return nil, err
	}
	root := make(map[string]map[string]bool, len(raw))
	for g, v := range raw {
		switch mm := v.(type) {
		case map[string]any:
			sub := make(map[string]bool, len(mm))
			for k, vv := range mm {
				if b, ok := vv.(bool); ok {
					sub[k] = b
				}
			}
			root[g] = sub
		case map[string]bool:
			root[g] = mm
		}
	}
	return root, nil
}

func (s *Store) writeRoot(root map[string]map[string]bool) error {
	raw := make(map[string]any, len(root))
	for g, mm := range root {
		sub := make(map[string]any, len(mm))
		for k, v := range mm {
			sub[k] = v
		}
		raw[g] = sub
	}
	return s.db.SetAll(raw)
}

func (s *Store) isEnabled(group, key string, def bool) (bool, error) {
	root, err := s.readRoot()
	if err != nil {
		return def, err
	}
	if sub, ok := root[group]; ok {
		if v, ok := sub[key]; ok {
			return v, nil
		}
	}
	return def, nil
}

func (s *Store) setEnabled(group, key string, enabled bool) error {
	root, err := s.readRoot()
	if err != nil {
		return err
	}
	sub := root[group]
	sub[key] = enabled
	return s.writeRoot(root)
}

func (s *Store) deleteKey(group, key string) error {
	root, err := s.readRoot()
	if err != nil {
		return err
	}
	if sub, ok := root[group]; ok {
		delete(sub, key)
	}
	return s.writeRoot(root)
}
