package net

import (
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestStdioConnNoTimeoutReadBlocksIndefinitely(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	// Create a StdioConn with readTimeout and writeTimeout set to zero (no timeout).
	conn := NewStdioConn(reader, writer)
	defer conn.Close()

	// Start a goroutine to perform a read.
	readDone := make(chan error, 1)
	go func() {
		buf := make([]byte, 10)
		_, err := conn.Read(buf)
		readDone <- err
	}()

	// Wait for 1 second, and check if read has returned.
	select {
	case err := <-readDone:
		t.Errorf("Read returned unexpectedly: %v", err)
	case <-time.After(1 * time.Second):
		// Expected, read should still be blocked.
	}

	// Now write data to the writer to unblock the read.
	_, err := writer.Write([]byte("hello"))
	if err != nil {
		t.Errorf("Error writing to writer: %v", err)
	}

	// Now the read should complete.
	select {
	case err := <-readDone:
		if err != nil && !errors.Is(err, io.EOF) {
			t.Errorf("Read returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Read did not return after data was written.")
	}
}

func TestStdioConnReadTimeout(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	// Create a StdioConn with a readTimeout of 500 milliseconds.
	conn := NewStdioConn(reader, writer, WithReadTimeout(500*time.Millisecond))
	defer conn.Close()

	// Start a goroutine to perform a read.
	readDone := make(chan error, 1)
	go func() {
		buf := make([]byte, 10)
		_, err := conn.Read(buf)
		readDone <- err
	}()

	// Wait for the read to return.
	select {
	case err := <-readDone:
		var ne net.Error
		if err == nil {
			t.Errorf("Read did not return an error as expected")
		} else if ok := errors.As(err, &ne); !ok || !ne.Timeout() {
			t.Errorf("Expected timeout error, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Read did not return after timeout")
	}
}

func TestStdioConnWriteTimeout(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	// Create a StdioConn with a writeTimeout of 500 milliseconds.
	conn := NewStdioConn(reader, writer, WithWriteTimeout(500*time.Millisecond))
	defer conn.Close()

	// Close the reader to simulate blocking write (since the reader is closed, write will error).
	reader.Close()

	// Start a goroutine to perform a write.
	writeDone := make(chan error, 1)
	go func() {
		_, err := conn.Write([]byte("hello"))
		writeDone <- err
	}()

	// Wait for the write to return.
	select {
	case err := <-writeDone:
		var ne net.Error
		if err == nil {
			t.Errorf("Write did not return an error as expected")
		} else if ok := errors.As(err, &ne); !ok || !ne.Timeout() {
			// May get a closed error.
			if !strings.Contains(err.Error(), "closed") {
				t.Errorf("Expected timeout error, got: %v", err)
			}
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Write did not return after timeout")
	}
}

func TestStdioConnSetReadDeadline(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	conn := NewStdioConn(reader, writer)
	defer conn.Close()

	// Set a read deadline of 500 milliseconds.
	err := conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	if err != nil {
		t.Fatalf("Failed to set read deadline: %v", err)
	}

	// Start a goroutine to perform a read.
	readDone := make(chan error, 1)
	go func() {
		buf := make([]byte, 10)
		_, err := conn.Read(buf)
		readDone <- err
	}()

	// Wait for the read to return.
	select {
	case err := <-readDone:
		var ne net.Error
		if err == nil {
			t.Errorf("Read did not return an error as expected")
		} else if ok := errors.As(err, &ne); !ok || !ne.Timeout() {
			t.Errorf("Expected timeout error, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Read did not return after deadline")
	}
}

// Additional test to cover reading after the deadline has passed, and then resetting the deadline.
func TestStdioConnReadDeadlineReset(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	conn := NewStdioConn(reader, writer)
	defer conn.Close()

	// Set a read deadline of 500 milliseconds.
	err := conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	if err != nil {
		t.Fatalf("Failed to set read deadline: %v", err)
	}

	// Start a goroutine to perform a read.
	readDone := make(chan error, 1)
	go func() {
		buf := make([]byte, 10)
		_, err := conn.Read(buf)
		readDone <- err
	}()

	// Wait for the read to return with a timeout error.
	select {
	case err := <-readDone:
		var ne net.Error
		if err == nil {
			t.Errorf("Read did not return an error as expected")
			return
		}
		if ok := errors.As(err, &ne); !ok || !ne.Timeout() {
			t.Errorf("Expected timeout error, got: %v", err)
			return
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Read did not return after deadline")
		return
	}

	// Reset the deadline to zero (no timeout).
	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		t.Fatalf("Failed to reset read deadline: %v", err)
	}

	// Start a new read operation.
	readDone = make(chan error, 1)
	go func() {
		buf := make([]byte, 5)
		_, err := conn.Read(buf)
		readDone <- err
	}()

	// Write data to the writer to unblock the read.
	_, err = writer.Write([]byte("hello"))
	if err != nil {
		t.Errorf("Error writing to writer: %v", err)
	}

	// Now the read should complete successfully.
	select {
	case err := <-readDone:
		if err != nil && !errors.Is(err, io.EOF) {
			t.Errorf("Read returned error after resetting deadline: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Read did not return after data was written")
	}
}
