package attachment

import "strings"

// HandleRef preserves legacy handle-style references (PRs, doc indexes, etc.).
type HandleRef struct {
	Handle string `json:"handle"`
}

func (ref *HandleRef) PopulateRef() error {
	ref.Handle = strings.TrimSpace(ref.Handle)

	return nil
}
