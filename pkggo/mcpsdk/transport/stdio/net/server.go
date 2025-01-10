package net

import (
	"bufio"
	"context"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// MessageHandler defines how messages are handled
type MessageHandler interface {
	HandleMessage(writer *bufio.Writer, msg []byte)
}

// Server orchestrates the transport, framing, and message handling
type Server struct {
	listener net.Listener
	framer   MessageFramer
	handler  MessageHandler
	done     chan struct{}
	doneOnce sync.Once
	wg       sync.WaitGroup // Tracks connection handler goroutines
	connMu   sync.Mutex
	conns    map[net.Conn]struct{}
}

// NewServer creates a new Server with provided components
func NewServer(listener net.Listener, framer MessageFramer, handler MessageHandler) *Server {
	return &Server{
		listener: listener,
		framer:   framer,
		handler:  handler,
		done:     make(chan struct{}),
		conns:    make(map[net.Conn]struct{}),
	}
}

// Serve starts the server and listens for connections
func (s *Server) Serve() error {
	defer s.listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				// Temporary error, continue accepting
				continue
			}
			// Check if listener is closed due to shutdown
			select {
			case <-s.done:
				return nil // Shutdown initiated
			default:
			}
			return err
		}

		s.connMu.Lock()
		s.conns[conn] = struct{}{}
		s.connMu.Unlock()

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConnection(conn)
		}()
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	var handlerWG sync.WaitGroup
	defer func() {
		handlerWG.Wait()
		conn.Close()
		s.connMu.Lock()
		delete(s.conns, conn)
		s.connMu.Unlock()
	}()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	var writeMutex sync.Mutex
	for {
		msg, err := s.framer.ReadMessage(reader)
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "EOF") {
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
			// log.Print("Started")
			defer handlerWG.Done()
			writeMutex.Lock()
			defer writeMutex.Unlock()
			s.handler.HandleMessage(writer, msgCopy)
			writer.Flush()
			// log.Print("return")
		}(msg)
	}

}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Signal to stop accepting new connections and stop processing
	s.doneOnce.Do(func() {
		close(s.done)
		s.listener.Close()
	})

	// Close all active connections
	s.connMu.Lock()
	for conn := range s.conns {
		conn.Close()
	}
	s.connMu.Unlock()

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
