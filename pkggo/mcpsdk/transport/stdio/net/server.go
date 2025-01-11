// server.go
package net

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// MessageHandler defines how messages are handled
type MessageHandler interface {
	HandleMessage(writer io.Writer, msg []byte)
}

// Server orchestrates the transport, framing, and message handling
type Server struct {
	conn     net.Conn
	framer   MessageFramer
	handler  MessageHandler
	done     chan struct{}
	doneOnce sync.Once
	wg       sync.WaitGroup // Tracks connection handler goroutines
}

// NewServer creates a new Server with provided components
func NewServer(conn net.Conn, framer MessageFramer, handler MessageHandler) *Server {
	return &Server{
		conn:    conn,
		framer:  framer,
		handler: handler,
		done:    make(chan struct{}),
	}
}

// Serve starts the server and listens for connections
func (s *Server) Serve() error {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.handleConnection(s.conn)
	}()
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	var handlerWG sync.WaitGroup
	defer func() {
		handlerWG.Wait()
		conn.Close()
	}()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	var writeMutex sync.Mutex

	for {
		select {
		case <-s.done:
			return // Exit if shutdown signal is received
		default:
		}

		msg, err := s.framer.ReadMessage(reader)
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "EOF") ||
				strings.Contains(err.Error(), "closed") {
				return // Client closed the connection
			}
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				time.Sleep(10 * time.Microsecond)
				continue // Temporary error, try reading again
			}
			// Unrecoverable error, exit the connection handler
			return
		}

		// Start a new goroutine to handle the message
		handlerWG.Add(1)
		go func(msgCopy []byte) {
			defer handlerWG.Done()
			// Create a buffer to collect the handler's output
			var responseBuffer bytes.Buffer
			// Provide an io.Writer to the handler
			s.handler.HandleMessage(&responseBuffer, msgCopy)
			// Write the framed message to the underlying writer
			writeMutex.Lock()
			defer writeMutex.Unlock()
			// Frame the message
			err := s.framer.WriteMessage(writer, responseBuffer.Bytes())
			if err != nil {
				// Handle write error
				return
			}
			err = writer.Flush()
			if err != nil {
				// Handle flush error
				return
			}
		}(msg)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Signal to stop accepting new connections and stop processing
	s.doneOnce.Do(func() {
		close(s.done)
		s.conn.Close() // Close the connection to unblock ReadMessage
	})

	doneChan := make(chan struct{})
	go func() {
		s.wg.Wait() // Wait for all connection handlers to finish
		close(doneChan)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err() // Context canceled or timed out
	case <-doneChan:
		return nil // Shutdown complete
	}
}
