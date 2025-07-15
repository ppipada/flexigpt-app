package builtin

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewAsyncRebuilder(t *testing.T) {
	tests := []struct {
		name           string
		maxAge         time.Duration
		expectedMaxAge time.Duration
	}{
		{
			name:           "positive maxAge",
			maxAge:         time.Hour,
			expectedMaxAge: time.Hour,
		},
		{
			name:           "zero maxAge",
			maxAge:         0,
			expectedMaxAge: alwaysStale,
		},
		{
			name:           "negative maxAge",
			maxAge:         -time.Hour,
			expectedMaxAge: alwaysStale,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := func() error { return nil }
			r := NewAsyncRebuilder(tt.maxAge, fn)

			if r.maxAge != tt.expectedMaxAge {
				t.Errorf("expected maxAge %v, got %v", tt.expectedMaxAge, r.maxAge)
			}
			if r.lastRun != 0 {
				t.Errorf("expected lastRun to be 0, got %d", r.lastRun)
			}
			if r.running != 0 {
				t.Errorf("expected running to be 0, got %d", r.running)
			}
		})
	}
}

func TestAsyncRebuilder_Force(t *testing.T) {
	tests := []struct {
		name        string
		fnError     error
		expectError bool
	}{
		{
			name:        "successful force",
			fnError:     nil,
			expectError: false,
		},
		{
			name:        "force with error",
			fnError:     errors.New("rebuild failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var callCount int32
			fn := func() error {
				atomic.AddInt32(&callCount, 1)
				return tt.fnError
			}

			r := NewAsyncRebuilder(time.Hour, fn)
			err := r.Force()

			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got %v", err)
			}

			if atomic.LoadInt32(&callCount) != 1 {
				t.Errorf("expected fn to be called once, got %d", callCount)
			}

			if !tt.expectError && r.lastRun == 0 {
				t.Error("expected lastRun to be updated after successful Force")
			}
			if tt.expectError && r.lastRun != 0 {
				t.Error("expected lastRun to remain 0 after failed Force")
			}
		})
	}
}

func TestAsyncRebuilder_MarkFresh(t *testing.T) {
	r := NewAsyncRebuilder(time.Hour, func() error { return nil })

	if r.lastRun != 0 {
		t.Error("expected initial lastRun to be 0")
	}

	before := time.Now().UnixNano()
	r.MarkFresh()
	after := time.Now().UnixNano()

	if r.lastRun < before || r.lastRun > after {
		t.Errorf("expected lastRun to be between %d and %d, got %d", before, after, r.lastRun)
	}
}

func TestAsyncRebuilder_Trigger(t *testing.T) {
	tests := []struct {
		name           string
		maxAge         time.Duration
		initialLastRun time.Time
		expectRun      bool
		description    string
	}{
		{
			name:           "fresh snapshot",
			maxAge:         time.Hour,
			initialLastRun: time.Now().Add(-30 * time.Minute),
			expectRun:      false,
			description:    "should not run when snapshot is fresh",
		},
		{
			name:           "stale snapshot",
			maxAge:         time.Hour,
			initialLastRun: time.Now().Add(-2 * time.Hour),
			expectRun:      true,
			description:    "should run when snapshot is stale",
		},
		{
			name:           "never run before",
			maxAge:         time.Hour,
			initialLastRun: time.Time{},
			expectRun:      true,
			description:    "should run when never run before",
		},
		{
			name:           "always stale maxAge",
			maxAge:         alwaysStale,
			initialLastRun: time.Now(),
			expectRun:      true,
			description:    "should always run with alwaysStale maxAge",
		},
		{
			name:           "boundary case - exactly maxAge",
			maxAge:         time.Hour,
			initialLastRun: time.Now().Add(-time.Hour),
			expectRun:      true,
			description:    "should run when exactly at maxAge boundary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var callCount int32
			done := make(chan struct{})

			fn := func() error {
				atomic.AddInt32(&callCount, 1)
				close(done)
				return nil
			}

			r := NewAsyncRebuilder(tt.maxAge, fn)
			if !tt.initialLastRun.IsZero() {
				atomic.StoreInt64(&r.lastRun, tt.initialLastRun.UnixNano())
			}

			r.Trigger()

			if tt.expectRun {
				select {
				case <-done:
					// Expected.
				case <-time.After(100 * time.Millisecond):
					t.Error("expected function to be called but it wasn't")
				}
			} else {
				select {
				case <-done:
					t.Error("expected function not to be called but it was")
				case <-time.After(50 * time.Millisecond):
					// Expected.
				}
			}

			expectedCalls := int32(0)
			if tt.expectRun {
				expectedCalls = 1
			}
			if atomic.LoadInt32(&callCount) != expectedCalls {
				t.Errorf("expected %d calls, got %d", expectedCalls, callCount)
			}
		})
	}
}

func TestAsyncRebuilder_TriggerWithError(t *testing.T) {
	var callCount int32
	done := make(chan struct{})

	fn := func() error {
		atomic.AddInt32(&callCount, 1)
		close(done)
		return errors.New("rebuild failed")
	}

	r := NewAsyncRebuilder(time.Hour, fn)
	r.Trigger()

	select {
	case <-done:
		// Expected.
	case <-time.After(100 * time.Millisecond):
		t.Error("expected function to be called")
	}

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	// LastRun should not be updated on error.
	if r.lastRun != 0 {
		t.Error("expected lastRun to remain 0 after error")
	}
}

func TestAsyncRebuilder_TriggerWithPanic(t *testing.T) {
	var callCount int32
	done := make(chan struct{})

	fn := func() error {
		atomic.AddInt32(&callCount, 1)
		close(done)
		panic("test panic")
	}

	r := NewAsyncRebuilder(time.Hour, fn)
	r.Trigger()

	select {
	case <-done:
		// Expected.
	case <-time.After(100 * time.Millisecond):
		t.Error("expected function to be called")
	}

	// Give some time for panic recovery.
	time.Sleep(10 * time.Millisecond)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	// Verify that running flag is reset even after panic.
	if atomic.LoadInt32(&r.running) != 0 {
		t.Error("expected running flag to be reset after panic")
	}
}

func TestAsyncRebuilder_ConcurrentTriggers(t *testing.T) {
	var callCount int32
	var startedCount int32
	started := make(chan struct{})
	proceed := make(chan struct{})

	fn := func() error {
		atomic.AddInt32(&startedCount, 1)
		started <- struct{}{}
		<-proceed // Block until we signal to proceed
		atomic.AddInt32(&callCount, 1)
		return nil
	}

	r := NewAsyncRebuilder(time.Hour, fn)

	// Start multiple triggers concurrently.
	const numTriggers = 10
	var wg sync.WaitGroup
	wg.Add(numTriggers)

	for range numTriggers {
		go func() {
			defer wg.Done()
			r.Trigger()
		}()
	}

	// Wait for one goroutine to start.
	select {
	case <-started:
		// Expected.
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected at least one goroutine to start")
	}

	// Give other triggers a chance to try.
	time.Sleep(50 * time.Millisecond)

	// Only one should have started.
	if atomic.LoadInt32(&startedCount) != 1 {
		t.Errorf("expected exactly 1 goroutine to start, got %d", startedCount)
	}

	// Allow the running goroutine to complete.
	close(proceed)
	<-r.IsDone()
	wg.Wait()

	// Verify only one completed.
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected exactly 1 call to complete, got %d", callCount)
	}

	// Verify running flag is reset.
	if atomic.LoadInt32(&r.running) != 0 {
		t.Error("expected running flag to be reset")
	}
}

func TestAsyncRebuilder_TriggerAfterCompletion(t *testing.T) {
	var callCount int32

	fn := func() error {
		atomic.AddInt32(&callCount, 1)
		return nil
	}

	r := NewAsyncRebuilder(alwaysStale, fn) // Always stale so it always runs

	// First trigger.
	done1 := make(chan struct{})
	go func() {
		r.Trigger()
		done1 <- struct{}{}
	}()

	// Wait for first trigger to complete.
	select {
	case <-done1:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("first trigger didn't complete")
	}

	// Give it a moment to fully complete.
	time.Sleep(10 * time.Millisecond)

	// Second trigger should also run since we use alwaysStale.
	done2 := make(chan struct{})
	go func() {
		r.Trigger()
		done2 <- struct{}{}
	}()

	select {
	case <-done2:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("second trigger didn't complete")
	}

	// Give it a moment to fully complete.
	time.Sleep(10 * time.Millisecond)

	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestAsyncRebuilder_RaceConditionStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	var callCount int32
	var completedCount int32

	fn := func() error {
		atomic.AddInt32(&callCount, 1)
		// Simulate some work.
		time.Sleep(time.Millisecond)
		atomic.AddInt32(&completedCount, 1)
		return nil
	}

	r := NewAsyncRebuilder(alwaysStale, fn)

	// Launch many concurrent triggers.
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range 10 {
				r.Trigger()
				time.Sleep(time.Microsecond) // Small delay to increase chance of races
			}
		}()
	}

	wg.Wait()

	// Wait for any remaining goroutines to complete.
	time.Sleep(100 * time.Millisecond)

	started := atomic.LoadInt32(&callCount)
	completed := atomic.LoadInt32(&completedCount)

	if started != completed {
		t.Errorf("mismatch between started (%d) and completed (%d) calls", started, completed)
	}

	// We should have some calls, but not necessarily all triggers should result in calls.
	if completed == 0 {
		t.Error("expected at least some calls to be made")
	}

	// Verify final state.
	if atomic.LoadInt32(&r.running) != 0 {
		t.Error("expected running flag to be 0 after all operations complete")
	}
}

func TestAsyncRebuilder_MaxAgeEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		maxAge time.Duration
		setup  func(*AsyncRebuilder)
		expect bool
	}{
		{
			name:   "minimum positive duration",
			maxAge: time.Nanosecond,
			setup: func(r *AsyncRebuilder) {
				r.MarkFresh()
				time.Sleep(2 * time.Nanosecond) // Ensure it's stale
			},
			expect: true,
		},
		{
			name:   "maximum duration",
			maxAge: time.Duration(1<<63 - 1), // Max int64
			setup: func(r *AsyncRebuilder) {
				r.MarkFresh()
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			done := make(chan struct{})

			fn := func() error {
				called = true
				close(done)
				return nil
			}

			r := NewAsyncRebuilder(tt.maxAge, fn)
			tt.setup(r)
			r.Trigger()

			if tt.expect {
				select {
				case <-done:
					// Expected.
				case <-time.After(100 * time.Millisecond):
					t.Error("expected function to be called")
				}
			} else {
				select {
				case <-done:
					t.Error("expected function not to be called")
				case <-time.After(50 * time.Millisecond):
					// Expected.
				}
			}

			if called != tt.expect {
				t.Errorf("expected called=%v, got %v", tt.expect, called)
			}
		})
	}
}
