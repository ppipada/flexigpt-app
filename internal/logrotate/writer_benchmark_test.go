package logrotate

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func BenchmarkWriter(b *testing.B) {
	tests := []struct {
		name        string
		messages    int
		messageSize int
		writers     int
	}{
		{"1000Messages_100BytesPerMessage_1Writer", 1000, 100, 1},
		{"1000Messages_100BytesPerMessage_2Writers", 1000, 100, 2},
		{"1000Messages_100BytesPerMessage_4Writers", 1000, 100, 4},
		{"100000Messages_100BytesPerMessage_1Writer", 100000, 100, 1},
		{"100000Messages_100BytesPerMessage_2Writers", 100000, 100, 2},
		{"100000Messages_100BytesPerMessage_4Writers", 100000, 100, 4},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			benchmarkWriter(b, tt.messages, tt.messageSize, tt.writers)
		})
	}
}

func benchmarkWriter(b *testing.B, messages, messageSize, writers int) {
	b.Helper()
	slogOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, slogOpts))

	dir := b.TempDir()

	for b.Loop() {
		w, err := New(logger, Options{
			Directory: dir,
		})
		if err != nil {
			b.Fatalf("failed to create writer: %v", err)
		}

		var wg sync.WaitGroup
		errCh := make(chan error, writers)

		for i := range writers {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				for range messages {
					if _, err := w.Write([]byte(strings.Repeat(strconv.Itoa(i), messageSize))); err != nil {
						errCh <- fmt.Errorf("failed to write: %w", err)
						return
					}
				}
			}(i)
		}

		wg.Wait()
		close(errCh)

		for err := range errCh {
			if err != nil {
				b.Error(err)
			}
		}

		if err := w.Close(); err != nil {
			b.Fatalf("failed to close writer: %v", err)
		}
	}
}
