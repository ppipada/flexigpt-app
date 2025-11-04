package filenameprovider

import (
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var nonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9]`)

type UUIDv7Provider struct{}

// Build "<uuid>_<sanitised-title>.json".
func (p *UUIDv7Provider) Build(info FileInfo) (string, error) {
	if info.ID == "" {
		return "", errors.New("missing ID")
	}
	title := info.Title
	if title == "" {
		title = "New Conversation"
	}
	if len(title) > 64 {
		title = title[:64]
	}
	title = nonAlphaNum.ReplaceAllString(title, "_")

	return fmt.Sprintf("%s_%s.json", info.ID, title), nil
}

func (p *UUIDv7Provider) CreatedAt(filename string) (time.Time, error) {
	info, err := p.Parse(filename)
	if err != nil {
		return time.Time{}, err
	}
	return info.CreatedAt, nil
}

// Parse does the opposite of build.
// But, given that build is lossy i.e it looses non alpha numeric chars it is not a exact repro of title.
func (p *UUIDv7Provider) Parse(filename string) (*FileInfo, error) {
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, filepath.Ext(base))

	parts := strings.SplitN(base, "_", 2)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid file name: %s", filename)
	}
	id := parts[0]
	title := ""
	if len(parts) == 2 {
		title = strings.ReplaceAll(parts[1], "_", " ")
	}

	created, err := ExtractTimeFromUUIDv7(id)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		ID:        id,
		Title:     title,
		CreatedAt: created,
	}, nil
}

func ExtractTimeFromUUIDv7(uuidStr string) (time.Time, error) {
	if len(uuidStr) != 36 {
		return time.Time{}, fmt.Errorf("invalid UUIDv7: %s", uuidStr)
	}
	uuidStr = strings.ReplaceAll(uuidStr, "-", "")
	b, err := hex.DecodeString(uuidStr)
	if err != nil {
		return time.Time{}, err
	}
	var ms int64
	for i := range 6 {
		ms = (ms << 8) | int64(b[i])
	}
	return time.UnixMilli(ms), nil
}
