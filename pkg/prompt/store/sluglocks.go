package store

import (
	"sync"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
)

// slugLocks manages a RW-mutex per bundleID|slug for concurrency control on templates.
// Note: This implementation allows indefinite growth of the lock map, which is acceptable
// for the current use case as the number of unique bundle|slug combinations is expected
// to be relatively small and bounded by actual usage patterns.
type slugLocks struct {
	mu sync.Mutex
	m  map[string]*sync.RWMutex
}

// newSlugLocks creates a new slugLocks instance.
func newSlugLocks() *slugLocks {
	return &slugLocks{m: map[string]*sync.RWMutex{}}
}

// lockKey returns the mutex for a given bundleID and slug, creating it if necessary.
func (l *slugLocks) lockKey(
	bundleID bundleitemutils.BundleID,
	slug bundleitemutils.ItemSlug,
) *sync.RWMutex {
	k := string(bundleID) + "|" + string(slug)
	l.mu.Lock()
	defer l.mu.Unlock()
	if lk, ok := l.m[k]; ok {
		return lk
	}
	lk := &sync.RWMutex{}
	l.m[k] = lk
	return lk
}
