// Package booloverlay provides a thread-safe feature-flag store that is
// persisted as JSON through simplemapdb/filestore.
package booloverlay

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
)

// GroupID is the compile-time type for group names stored on disk.
type GroupID string

// KeyID is the compile-time type for individual entry names.
type KeyID string

// Root is the complete in-memory image of the file.
type Root map[GroupID]map[KeyID]Flag

// Key identifies a single flag entry.
type Key interface {
	Group() GroupID
	ID() KeyID
}

// Flag represents one feature-flag record.
type Flag struct {
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

type Store struct {
	mu  sync.RWMutex
	db  *filestore.MapFileStore
	reg map[GroupID]struct{}
}

// Option customises store construction.
type Option func(*Store) error

// WithKeyType registers the group of key type K and guarantees its presence on disk.
func WithKeyType[K Key]() Option {
	var zero K
	group := zero.Group()

	return func(s *Store) error {
		if _, dup := s.reg[group]; dup {
			return fmt.Errorf("booloverlay: duplicate registration for group %q", group)
		}
		s.reg[group] = struct{}{}

		s.mu.Lock()
		defer s.mu.Unlock()

		root, err := s.readRoot()
		if err != nil {
			return err
		}
		if _, ok := root[group]; !ok {
			root[group] = make(map[KeyID]Flag)
			return s.writeRoot(root)
		}
		return nil
	}
}

// NewStore opens or creates the file and applies options.
func NewStore(path string, opts ...Option) (*Store, error) {
	fs, err := filestore.NewMapFileStore(
		path,
		map[string]any{},
		filestore.WithCreateIfNotExists(true),
		filestore.WithAutoFlush(true),
		filestore.WithEncoderDecoder(encdec.JSONEncoderDecoder{}),
	)
	if err != nil {
		return nil, err
	}

	st := &Store{
		db:  fs,
		reg: make(map[GroupID]struct{}),
	}
	for _, opt := range opts {
		if err := opt(st); err != nil {
			return nil, err
		}
	}
	return st, nil
}

// GetFlag returns the stored flag, a presence indicator and an error.
func (s *Store) GetFlag(k Key) (Flag, bool, error) {
	group := k.Group()
	if err := s.ensureRegistered(group); err != nil {
		return Flag{}, false, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	root, err := s.readRoot()
	if err != nil {
		return Flag{}, false, err
	}
	if grp, ok := root[group]; ok {
		if flag, ok2 := grp[k.ID()]; ok2 {
			return flag, true, nil
		}
	}
	return Flag{}, false, nil
}

// SetFlag stores the enabled state and updates timestamps.
func (s *Store) SetFlag(k Key, enabled bool) (Flag, error) {
	group := k.Group()
	if err := s.ensureRegistered(group); err != nil {
		return Flag{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	root, err := s.readRoot()
	if err != nil {
		return Flag{}, err
	}
	if root[group] == nil {
		root[group] = make(map[KeyID]Flag)
	}

	now := time.Now().UTC()
	flag, exists := root[group][k.ID()]
	if !exists {
		flag = Flag{
			Enabled:    enabled,
			CreatedAt:  now,
			ModifiedAt: now,
		}
	} else {
		flag.Enabled = enabled
		flag.ModifiedAt = now
	}
	root[group][k.ID()] = flag

	if err := s.writeRoot(root); err != nil {
		return Flag{}, err
	}
	return flag, nil
}

// Delete removes the flag entry if present.
func (s *Store) Delete(k Key) error {
	group := k.Group()
	if err := s.ensureRegistered(group); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	root, err := s.readRoot()
	if err != nil {
		return err
	}
	if grp, ok := root[group]; ok {
		delete(grp, k.ID())
	}
	return s.writeRoot(root)
}

func (s *Store) ensureRegistered(group GroupID) error {
	if _, ok := s.reg[group]; !ok {
		return fmt.Errorf("booloverlay: group %q is not registered", group)
	}
	return nil
}

// readRoot converts the loosely typed map from filestore into Root.
func (s *Store) readRoot() (Root, error) {
	raw, err := s.db.GetAll(false)
	if err != nil {
		return nil, err
	}

	// Fast path for an empty file.
	if len(raw) == 0 {
		return make(Root), nil
	}

	// Marshal then unmarshal to leverage the JSON type conversion rules.
	jsonData, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var root Root
	if err := json.Unmarshal(jsonData, &root); err != nil {
		return nil, err
	}
	if root == nil {
		root = make(Root)
	}
	return root, nil
}

// writeRoot converts Root into the map[string]any format expected by filestore.
func (s *Store) writeRoot(root Root) error {
	mp, err := encdec.StructWithJSONTagsToMap(root)
	if err != nil {
		return err
	}
	return s.db.SetAll(mp)
}
