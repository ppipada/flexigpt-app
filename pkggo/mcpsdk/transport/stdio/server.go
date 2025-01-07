package stdio

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// StdioConn implements net.Conn over os.Stdin and os.Stdout
type StdioConn struct {
	closed     chan struct{} // Channel to signal connection closed
	closeOnce  sync.Once     // Ensures Close only runs once
	reader     *bufio.Reader // Buffered reader wrapping os.Stdin
	writer     *bufio.Writer // Buffered writer wrapping os.Stdout
	readDead   time.Time     // Read deadline
	writeDead  time.Time     // Write deadline
	deadlineMu sync.Mutex    // Mutex to protect readDead and writeDead
	writeMutex sync.Mutex    // Mutex to synchronize writes
}

// NewStdioConn creates a new StdioConn wrapping os.Stdin and os.Stdout
func NewStdioConn() *StdioConn {
	return &StdioConn{
		closed: make(chan struct{}),
		reader: bufio.NewReader(os.Stdin),
		writer: bufio.NewWriter(os.Stdout),
	}
}

// Read reads data from the connection, implementing net.Conn
func (c *StdioConn) Read(b []byte) (int, error) {
	// Lock to safely read the read deadline
	c.deadlineMu.Lock()
	readDeadline := c.readDead
	c.deadlineMu.Unlock()

	// Check if read deadline has passed
	if !readDeadline.IsZero() {
		if time.Now().After(readDeadline) {
			return 0, os.ErrDeadlineExceeded
		}
	}

	// Check if the connection is closed
	select {
	case <-c.closed:
		return 0, io.EOF
	default:
		// Continue reading
	}

	// Read data into the buffer
	n, err := c.reader.Read(b)
	return n, err
}

// Write writes data to the connection, implementing net.Conn
func (c *StdioConn) Write(b []byte) (int, error) {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	// Lock to safely read the write deadline
	c.deadlineMu.Lock()
	writeDeadline := c.writeDead
	c.deadlineMu.Unlock()

	// Check if write deadline has passed
	if !writeDeadline.IsZero() {
		if time.Now().After(writeDeadline) {
			return 0, os.ErrDeadlineExceeded
		}
	}

	// Check if the connection is closed
	select {
	case <-c.closed:
		return 0, io.ErrClosedPipe
	default:
		// Continue writing
	}

	// Write data from the buffer
	n, err := c.writer.Write(b)
	if err != nil {
		return n, err
	}

	// Flush the buffered writer to ensure data is sent
	err = c.writer.Flush()
	return n, err
}

// Close closes the connection, implementing net.Conn
func (c *StdioConn) Close() error {
	// Ensure Close is only called once
	var err error
	c.closeOnce.Do(func() {
		close(c.closed)
		// Flush any remaining data
		err = c.writer.Flush()
	})
	return err
}

// LocalAddr returns the local network address, implementing net.Conn
func (c *StdioConn) LocalAddr() net.Addr {
	return StdioAddr{}
}

// RemoteAddr returns the remote network address, implementing net.Conn
func (c *StdioConn) RemoteAddr() net.Addr {
	return StdioAddr{}
}

// SetDeadline sets the read and write deadlines, implementing net.Conn
func (c *StdioConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

// SetReadDeadline sets the read deadline, implementing net.Conn
func (c *StdioConn) SetReadDeadline(t time.Time) error {
	c.deadlineMu.Lock()
	c.readDead = t
	c.deadlineMu.Unlock()
	return nil
}

// SetWriteDeadline sets the write deadline, implementing net.Conn
func (c *StdioConn) SetWriteDeadline(t time.Time) error {
	c.deadlineMu.Lock()
	c.writeDead = t
	c.deadlineMu.Unlock()
	return nil
}

// StdioAddr represents the network address for StdioConn
type StdioAddr struct{}

// Network returns the address's network name, implementing net.Addr
func (a StdioAddr) Network() string {
	return "stdio"
}

// String returns the string form of the address, implementing net.Addr
func (a StdioAddr) String() string {
	return "stdio"
}

// StdioListener implements net.Listener using StdioConn
type StdioListener struct {
	conn     net.Conn      // The connection to accept
	closed   chan struct{} // Channel to signal listener closed
	closeMu  sync.Mutex    // Protects close operations
	acceptMu sync.Mutex    // Ensures Accept is serialized
}

// NewStdioListener creates a new StdioListener
func NewStdioListener() *StdioListener {
	return &StdioListener{
		conn:   NewStdioConn(),
		closed: make(chan struct{}),
	}
}

// Accept accepts the connection, implementing net.Listener
func (l *StdioListener) Accept() (net.Conn, error) {
	l.acceptMu.Lock()
	defer l.acceptMu.Unlock()

	// Check if the listener is closed
	select {
	case <-l.closed:
		return nil, net.ErrClosed
	default:
		// Continue
	}

	l.closeMu.Lock()
	defer l.closeMu.Unlock()
	if l.conn == nil {
		// Reinitialize the connection for the next accept
		l.conn = NewStdioConn()
	}
	conn := l.conn
	l.conn = nil // Only accept one connection at a time
	return conn, nil
}

// Close closes the listener, implementing net.Listener
func (l *StdioListener) Close() error {
	l.closeMu.Lock()
	defer l.closeMu.Unlock()

	// Check if the listener is already closed
	select {
	case <-l.closed:
		// Already closed
		return net.ErrClosed
	default:
		close(l.closed)
	}

	// Close the connection if it exists
	if l.conn != nil {
		l.conn.Close()
		l.conn = nil
	}

	return nil
}

// Addr returns the listener's network address, implementing net.Listener
func (l *StdioListener) Addr() net.Addr {
	return StdioAddr{}
}

// RequestParams holds default parameters for HTTP requests
type RequestParams struct {
	Method string
	URL    string
	Header http.Header
}

// ResponseWriter implements http.ResponseWriter
type ResponseWriter struct {
	conn        net.Conn
	writeMu     *sync.Mutex
	header      http.Header
	statusCode  int
	wroteHeader bool
	buffer      bytes.Buffer // Buffer to store written data
}

// Header returns the header map to be sent by WriteHeader
func (w *ResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

// WriteHeader sends an HTTP response header with the provided status code
func (w *ResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		// Avoid multiple WriteHeader calls
		return
	}
	w.statusCode = statusCode
	w.wroteHeader = true
	// We don't output the status code or headers
}

// Write writes the data to the connection as part of an HTTP reply
func (w *ResponseWriter) Write(b []byte) (int, error) {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()

	// Buffer the data
	n, err := w.buffer.Write(b)
	if err != nil {
		return n, err
	}

	// Write to the connection
	_, err = w.conn.Write(b)
	return n, err
}

// Server handles incoming messages over the stdio transport
type Server struct {
	Handler       http.Handler
	listener      net.Listener
	writeMu       sync.Mutex
	wg            sync.WaitGroup
	done          chan struct{}
	requestParams RequestParams
}

// NewServer creates a new Server with optional configurations
func NewServer(handler http.Handler, opts ...ServerOption) *Server {
	s := &Server{
		Handler: handler,
		done:    make(chan struct{}),
	}
	// Set default request parameters
	s.requestParams.Method = "POST"
	s.requestParams.URL = "/"
	s.requestParams.Header = make(http.Header)
	// Apply options
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ServerOption represents an option for configuring the server
type ServerOption func(*Server)

// WithRequestParams sets the default request parameters
func WithRequestParams(params RequestParams) ServerOption {
	return func(s *Server) {
		s.requestParams = params
	}
}

// ListenAndServe starts the server and listens on stdio
func (s *Server) ListenAndServe() error {
	l := NewStdioListener()
	s.listener = l
	return s.Serve(l)
}

// Serve accepts connections and handles messages
func (s *Server) Serve(l net.Listener) error {
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	lines := make(chan []byte)
	errCh := make(chan error, 1)

	// Start a goroutine to read from conn and send lines over the channel
	go func() {
		reader := bufio.NewReader(conn)
		for {
			// Read until a newline
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					// Send error to errCh if it's not EOF
					errCh <- err
				}
				// Close the lines channel to signal EOF or error
				close(lines)
				return
			}

			// Send the line to the lines channel
			lines <- line
		}
	}()

	for {
		select {
		case <-s.done:
			// Shutdown signal received
			return nil
		case err := <-errCh:
			// An error occurred while reading
			return err
		case line, ok := <-lines:
			if !ok {
				// Lines channel closed, EOF reached
				return nil
			}

			// Process the line
			if len(bytes.TrimSpace(line)) == 0 {
				// Ignore empty messages (consecutive newlines)
				continue
			}

			s.processLine(conn, line)
		}
	}
}

// processLine processes a single line read from the input
func (s *Server) processLine(conn net.Conn, line []byte) {
	// Make a copy of the message to avoid data races
	msg := make([]byte, len(line))
	copy(msg, line)

	// Create a ResponseWriter for this handler
	w := &ResponseWriter{
		conn:    conn,
		writeMu: &s.writeMu,
	}

	// Create Request with the message as the body
	req, err := http.NewRequest(
		s.requestParams.Method,
		s.requestParams.URL,
		bytes.NewReader(msg),
	)
	if err != nil {
		// Log the error and return
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		return
	}

	// Copy default headers
	for k, v := range s.requestParams.Header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	// Set RemoteAddr
	req.RemoteAddr = conn.RemoteAddr().String()

	// Handle the request in a new goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.Handler.ServeHTTP(w, req)
	}()
}

// Shutdown gracefully shuts down the server without interrupting active handlers
func (s *Server) Shutdown(ctx context.Context) error {
	close(s.done)
	// Close the listener to unblock Accept
	if s.listener != nil {
		s.listener.Close()
	}
	// Wait for handlers to finish or context to timeout
	doneChan := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(doneChan)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-doneChan:
		return nil
	}
}
