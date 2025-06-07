package net

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// echoHandler echoes back the received message.
type echoHandler struct{}

func (h *echoHandler) HandleMessage(writer io.Writer, msg []byte) {
	log.Printf("Start echo write: %s", string(msg))
	_, _ = writer.Write(msg)
	log.Printf("Done echo write: %s", string(msg))
}

// delayHandler introduces a delay before responding.
type delayHandler struct {
	delay time.Duration
}

func (h *delayHandler) HandleMessage(writer io.Writer, msg []byte) {
	time.Sleep(h.delay)
	_, _ = writer.Write(msg)
}

// errorHandler simulates an error by sending invalid responses.
type errorHandler struct{}

func (h *errorHandler) HandleMessage(writer io.Writer, msg []byte) {
	// Send back an invalid response (e.g., no request ID).
	_, _ = writer.Write([]byte("invalid response\n"))
}

// initClientServer initializes a client and server for testing.
func initClientServer(
	handler MessageHandler,
	options ...ClientOption,
) (*Client, *Server, func()) {
	// Create connected pipes to simulate stdin and stdout.
	// Client writes to serverReader.
	clientReader, serverWriter := io.Pipe()
	// Server writes to clientReader.
	serverReader, clientWriter := io.Pipe()

	// Create StdioConn for the client and server using the connected pipes.
	clientConn := NewStdioConn(
		clientReader,
		clientWriter,
		WithReadTimeout(time.Second*1),
		WithWriteTimeout(time.Second*1),
	)
	serverConn := NewStdioConn(
		serverReader,
		serverWriter,
		WithReadTimeout(time.Second*1),
		WithWriteTimeout(time.Second*1),
	)

	framer := &LineFramer{}
	server := NewServer(serverConn, framer, handler)
	go func() {
		err := server.Serve()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			// Log the error if it's not due to server shutdown.
			fmt.Printf("Server error: %v\n", err)
		}
	}()
	// Client setup.
	client := NewClient(clientConn, framer, options...)

	// Teardown function to close client and server.
	teardown := func() {
		client.Close()
		_ = server.Shutdown(context.Background())
	}

	return client, server, teardown
}

// Helper function to generate random bytes.
func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		// Handle error appropriately in a real application.
		fmt.Println("Error generating random bytes:", err)
		return nil
	}

	// Replace any occurrence of '\n' with another character, e.g., 'A'.
	for i := range b {
		if b[i] == '\n' {
			// Replace '\n' with 'A' or any other character you prefer.
			b[i] = 'A'
		}
	}

	return b
}

func TestSynchronousSingleRequest(t *testing.T) {
	message := []any{
		map[string]any{
			"jsonrpc": "2.0",
			"method":  "notify",
			"params":  map[string]any{"message": "Hello"},
		},
		map[string]any{
			"jsonrpc": "2.0",
			"method":  "notify",
			"params":  map[string]any{"message": "World"},
		},
	}
	msgReqBytes, _ := json.Marshal(message)
	tests := []struct {
		name          string
		message       []byte
		expectedReply []byte
	}{
		{
			name:          "Simple message",
			message:       []byte("Hello, Server!"),
			expectedReply: []byte("Hello, Server!"),
		},
		{
			name:          "Empty message",
			message:       []byte(""),
			expectedReply: []byte(""),
		},
		{
			name:          "Large message",
			message:       []byte(strings.Repeat("A", 1024)),
			expectedReply: []byte(strings.Repeat("A", 1024)),
		},
		{
			name: "Special characters",
			// "Hello World" in Japanese.
			message:       []byte("こんにちは世界"),
			expectedReply: []byte("こんにちは世界"),
		},
		{
			name:          "Binary data",
			message:       []byte{0x00, 0xFF, 0xAA, 0x55},
			expectedReply: []byte{0x00, 0xFF, 0xAA, 0x55},
		},
		{
			name:          "UTF-8 characters",
			message:       []byte("😊🌟🚀"),
			expectedReply: []byte("😊🌟🚀"),
		},
		{
			name:    "Random data",
			message: randomBytes(512),
			// We'll compare the lengths instead.
			expectedReply: nil,
		},
		{
			name:    "msg interface",
			message: msgReqBytes,
			// We'll compare the lengths instead.
			expectedReply: msgReqBytes,
		},
	}

	handler := &echoHandler{}
	client, _, teardown := initClientServer(handler)

	defer teardown()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set long timeouts to avoid interference.
			_ = client.conn.SetDeadline(time.Now().Add(10 * time.Second))

			reply, err := client.Send(tt.message)
			if err != nil {
				t.Fatalf("Send failed: %v", err)
			}

			if tt.expectedReply != nil {
				if !bytes.Equal(reply, tt.expectedReply) {
					t.Errorf("Expected reply %v, got %v", tt.expectedReply, reply)
				}
			} else {
				if len(reply) != len(tt.message) {
					t.Errorf("Expected reply length %d, got %d", len(tt.message), len(reply))
				}
			}
		})
	}
}

func TestConcurrentRequestsSingleClient(t *testing.T) {
	// Define the number of concurrent requests.
	numRequests := 10

	// Handlers and functions for request ID assignment and extraction.
	var idCounter int32
	assignID := func(msg []byte) (*string, []byte, error) {
		id := atomic.AddInt32(&idCounter, 1)
		idStr := strconv.Itoa(int(id))
		msgWithID := append([]byte(idStr+":"), msg...)
		return &idStr, msgWithID, nil
	}
	extractID := func(msg []byte) (*string, []byte, error) {
		parts := bytes.SplitN(msg, []byte(":"), 2)
		if len(parts) != 2 {
			return nil, nil, errors.New("invalid message format")
		}
		id := string(parts[0])
		return &id, parts[1], nil
	}

	handler := &echoHandler{}
	client, _, teardown := initClientServer(
		handler,
		WithRequestIDFunctions(assignID, extractID),
	)

	defer teardown()

	// Prepare test messages.
	messages := make([]string, numRequests)
	for i := range numRequests {
		messages[i] = fmt.Sprintf("Message %d", i)
	}

	// Use WaitGroup to synchronize the test.
	var wg sync.WaitGroup
	wg.Add(numRequests)

	// Send requests concurrently.
	for i := range numRequests {
		go func(i int) {
			defer wg.Done()
			reply, err := client.Send([]byte(messages[i]))
			if err != nil {
				t.Errorf("Request %d failed: %v", i, err)
				return
			}
			if string(reply) != messages[i] {
				t.Errorf("Request %d expected reply %q, got %q", i, messages[i], string(reply))
			}
		}(i)
	}

	wg.Wait()
}

func TestClientRequestTimeout(t *testing.T) {
	// Server will delay response by 2 seconds.
	delay := 2 * time.Second
	timeout := 1 * time.Second

	handler := &delayHandler{delay: delay}
	clientOptions := []ClientOption{
		WithRequestTimeout(timeout),
	}

	// Handlers and functions for request ID assignment and extraction.
	var idCounter int32
	assignID := func(msg []byte) (*string, []byte, error) {
		id := atomic.AddInt32(&idCounter, 1)
		idStr := strconv.Itoa(int(id))
		msgWithID := append([]byte(idStr+":"), msg...)
		return &idStr, msgWithID, nil
	}
	extractID := func(msg []byte) (*string, []byte, error) {
		parts := bytes.SplitN(msg, []byte(":"), 2)
		if len(parts) != 2 {
			return nil, nil, errors.New("invalid message format")
		}
		id := string(parts[0])
		return &id, parts[1], nil
	}
	clientOptions = append(clientOptions, WithRequestIDFunctions(assignID, extractID))

	client, _, teardown := initClientServer(handler, clientOptions...)

	defer teardown()

	startTime := time.Now()
	_, err := client.Send([]byte("This will timeout"))
	elapsed := time.Since(startTime)

	if err == nil {
		t.Errorf("Expected timeout error but got none")
	} else {
		t.Logf("Received expected error: %v", err)
		if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "timed out") {
			t.Errorf("Expected a timeout error, got %v", err)
		}
	}

	if elapsed < timeout {
		t.Errorf("Timeout occurred too early: %v", elapsed)
	} else if elapsed > delay {
		t.Errorf("Timeout occurred too late: %v", elapsed)
	}
}

func TestServerShutdownDuringActiveConnections(t *testing.T) {
	handler := &delayHandler{delay: 2 * time.Second}
	client, server, teardown := initClientServer(handler)

	defer teardown()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		_, err := client.Send([]byte("Message before shutdown"))
		if err != nil {
			t.Logf("Client send error: %v", err)
		} else {
			t.Logf("Client received response before shutdown")
		}
	}()

	// Wait a moment and then shutdown the server.
	time.Sleep(1 * time.Second)
	err := server.Shutdown(t.Context())
	if err != nil {
		t.Errorf("Server shutdown error: %v", err)
	}

	wg.Wait()
}

func TestClientSendWithoutRequestIDFunctions(t *testing.T) {
	handler := &echoHandler{}
	client, _, teardown := initClientServer(handler)

	defer teardown()

	// Send a message without request ID functions (synchronous send).
	reply, err := client.Send([]byte("Hello, Server!"))
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if string(reply) != "Hello, Server!" {
		t.Errorf("Expected reply %q, got %q", "Hello, Server!", string(reply))
	}
}

func TestInvalidMessageFormat(t *testing.T) {
	// Handler sends back a message without request ID.
	handler := &errorHandler{}

	// Handlers and functions for request ID assignment and extraction.
	var idCounter int32
	assignID := func(msg []byte) (*string, []byte, error) {
		id := atomic.AddInt32(&idCounter, 1)
		idStr := strconv.Itoa(int(id))
		msgWithID := append([]byte(idStr+":"), msg...)
		return &idStr, msgWithID, nil
	}
	extractID := func(msg []byte) (*string, []byte, error) {
		_ = bytes.SplitN(msg, []byte(":"), 2)
		// Force an invalid format error.
		return nil, nil, errors.New("invalid message format")
	}
	clientOptions := []ClientOption{
		WithRequestIDFunctions(assignID, extractID),
		WithDeadLetterQueue(10),
		WithRequestTimeout(time.Second),
	}

	client, _, teardown := initClientServer(handler, clientOptions...)

	defer teardown()

	_, err := client.Send([]byte("Test message"))
	if err == nil {
		t.Errorf("Expected error due to invalid message format, but got none")
	}

	// Check the dead letter queue.
	item, err := client.PopDeadLetter()
	if err != nil {
		t.Errorf("Failed to pop dead letter: %v", err)
	} else {
		t.Logf("Dead letter item: response=%q, error=%v", item.Response, item.Err)
		if item.Err == nil {
			t.Errorf("Expected an error in dead letter item")
		}
	}
}

func TestErrorHandlingAndDeadLetterQueue(t *testing.T) {
	// Use errorHandler to send invalid responses.
	handler := &errorHandler{}
	clientOptions := []ClientOption{
		WithDeadLetterQueue(10),
	}

	// Handlers and functions for request ID assignment and extraction.
	var idCounter int32
	assignID := func(msg []byte) (*string, []byte, error) {
		id := atomic.AddInt32(&idCounter, 1)
		idStr := strconv.Itoa(int(id))
		msgWithID := append([]byte(idStr+":"), msg...)
		return &idStr, msgWithID, nil
	}
	extractID := func(msg []byte) (*string, []byte, error) {
		parts := bytes.SplitN(msg, []byte(":"), 2)
		if len(parts) != 2 {
			return nil, nil, errors.New("invalid message format")
		}
		id := string(parts[0])
		return &id, parts[1], nil
	}
	clientOptions = append(
		clientOptions,
		WithRequestIDFunctions(assignID, extractID),
		WithRequestTimeout(time.Millisecond*100),
	)

	client, _, teardown := initClientServer(handler, clientOptions...)

	defer teardown()

	_, err := client.Send([]byte("Test message"))
	if err == nil {
		t.Errorf("Expected error due to invalid response, but got none")
	}

	// Check the dead letter queue.
	item, err := client.PopDeadLetter()
	if err != nil {
		t.Errorf("Failed to pop dead letter: %v", err)
	} else {
		t.Logf("Dead letter item: response=%q, error=%v", item.Response, item.Err)
		if item.Err == nil {
			t.Errorf("Expected an error in dead letter item")
		}
	}
}
