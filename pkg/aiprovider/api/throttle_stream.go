package api

import (
	"strings"
	"sync"
	"time"
)

const (
	FlushInterval  = 256 * time.Millisecond
	FlushChunkSize = 512
)

// NewBufferedStreamer returns two functions:
//   - write(chunk)  -> use this instead of onStreamData
//   - flush()       -> call once when streaming is finished
func NewBufferedStreamer(
	onStreamData func(string) error,
	flushInterval time.Duration,
	maxSize int,
) (write func(string) error, flush func()) {
	var mu sync.Mutex
	var buf strings.Builder
	ticker := time.NewTicker(flushInterval)
	done := make(chan struct{})

	// Background goroutine time-based flush.
	go func() {
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				if buf.Len() > 0 {
					data := buf.String()
					buf.Reset()
					mu.Unlock()
					_ = onStreamData(data)
				} else {
					mu.Unlock()
				}
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	// Returns the wrapped write.
	write = func(chunk string) error {
		mu.Lock()
		buf.WriteString(chunk)
		over := buf.Len() >= maxSize
		if over {
			data := buf.String()
			buf.Reset()
			mu.Unlock()
			// Size-based flush.
			return onStreamData(data)
		}
		mu.Unlock()
		return nil
	}

	// Flush everything, stop ticker.
	flush = func() {
		close(done)
		mu.Lock()
		if buf.Len() > 0 {
			data := buf.String()
			buf.Reset()
			mu.Unlock()
			_ = onStreamData(data)
			return
		}
		mu.Unlock()
	}

	return write, flush
}
