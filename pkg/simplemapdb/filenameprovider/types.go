package filenameprovider

import "time"

// FileInfo is the logical information that can be extracted from a file name.
type FileInfo struct {
	ID        string
	Title     string
	CreatedAt time.Time
}

// Provider converts between logical information and a physical file name.
type Provider interface {
	// Build turns the logical conversation data into a file name.
	Build(info FileInfo) (string, error)

	// Parse does the inverse and must work with the file names returned by Build.
	Parse(filename string) (*FileInfo, error)

	// CreatedAt is a convenience wrapper, it may call Parse under the hood.
	CreatedAt(filename string) (time.Time, error)
}
