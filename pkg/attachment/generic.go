package attachment

import "strings"

// GenericRef preserves legacy handle-style references (PRs, doc indexes, etc.).
type GenericRef struct {
	Handle string `json:"handle"`
}

func (ref *GenericRef) PopulateRef() error {
	ref.Handle = strings.TrimSpace(ref.Handle)

	return nil
}
