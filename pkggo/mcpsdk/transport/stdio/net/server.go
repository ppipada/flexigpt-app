package net

import (
	"context"
	"io"
	"net"
	"sync"
)

// MessageHandler defines how messages are handled
type MessageHandler interface {
	HandleMessage(conn net.Conn, msg []byte)
}

// Server orchestrates the transport, framing, and message handling
type Server struct {
	listener net.Listener
	framer   MessageFramer
	handler  MessageHandler
	done     chan struct{}
	wg       sync.WaitGroup // Tracks connection handler goroutines
}

// NewServer creates a new Server with provided components
func NewServer(listener net.Listener, framer MessageFramer, handler MessageHandler) *Server {
	return &Server{
		listener: listener,
		framer:   framer,
		handler:  handler,
		done:     make(chan struct{}),
	}
}

// Serve starts the server and listens for connections
func (s *Server) Serve() error {
	defer s.listener.Close()

	for {
		select {
		case <-s.done:
			return nil
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					continue
				}
				return err
			}

			s.wg.Add(1) // Increment WaitGroup for the connection handler
			go func() {
				defer s.wg.Done() // Decrement when connection handler is done
				s.handleConnection(conn)
			}()
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	// WaitGroup to ensure all message handlers complete before closing the connection
	var handlerWG sync.WaitGroup
	// Close the connection after all message handlers are done
	defer func() {
		handlerWG.Wait()
		conn.Close()
	}()

	for {
		select {
		case <-s.done:
			return
		default:
			msg, err := s.framer.ReadMessage(conn)
			if err != nil {
				if err == io.EOF {
					return // Client closed the connection
				}
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					continue // Temporary error, try reading again
				}
				// Unrecoverable error, exit the connection handler
				return
			}

			// Ignore empty messages
			if len(msg) == 0 {
				continue
			}

			// Start a new goroutine to handle the message
			handlerWG.Add(1)
			go func(msgCopy []byte) {
				defer handlerWG.Done() // Decrement when message handler is done
				// Handle the message (assuming HandleMessage handles its own panic recovery)
				s.handler.HandleMessage(conn, msgCopy)
			}(msg)
		}
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Signal to stop accepting new connections and stop processing
	close(s.done)
	s.listener.Close()

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
