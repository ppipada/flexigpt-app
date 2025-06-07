package net

import (
	"fmt"
	"io"
	"net"
	"slices"
	"sync"
	"time"
)

// StdioAddr represents the network address for StdioConn.
type StdioAddr struct {
	network string
	address string
}

// Network returns the address's network name, implementing net.Addr.
func (a StdioAddr) Network() string {
	return a.network
}

// String returns the string form of the address, implementing net.Addr.
func (a StdioAddr) String() string {
	return a.address
}

// StdioConn implements net.Conn over io.Reader and io.Writer.
type StdioConn struct {
	addr net.Addr
	// Channel to signal connection closed.
	closed chan struct{}
	// Ensures Close only runs once.
	closeOnce sync.Once
	// Client and Server will do buffering as needed.
	// This conn does simple reader/writer interfacing.
	reader       io.Reader
	writer       io.Writer
	readTimeout  time.Duration
	writeTimeout time.Duration

	readCh  chan readResult
	writeCh chan writeRequest
}

// readResult represents the result of a read operation.
type readResult struct {
	data []byte
	err  error
}

// writeRequest represents a write operation request.
type writeRequest struct {
	data  []byte
	resCh chan writeResult
}

// writeResult represents the result of a write operation.
type writeResult struct {
	n   int
	err error
}

// StdioConnOption defines a functional option for StdioConn.
type StdioConnOption func(*StdioConn)

// NewStdioConn creates a new StdioConn with provided options.
func NewStdioConn(r io.Reader, w io.Writer, options ...StdioConnOption) *StdioConn {
	c := &StdioConn{
		addr: StdioAddr{
			network: "stdio",
			address: "stdio",
		},
		closed:       make(chan struct{}),
		reader:       r,
		writer:       w,
		readTimeout:  0,
		writeTimeout: 0,
		readCh:       make(chan readResult),
		writeCh:      make(chan writeRequest),
	}

	// Apply options.
	for _, option := range options {
		option(c)
	}

	// Start the read and write loops.
	go c.readLoop()
	go c.writeLoop()

	return c
}

// WithReadTimeout sets the default read timeout.
func WithReadTimeout(d time.Duration) StdioConnOption {
	return func(c *StdioConn) {
		c.readTimeout = d
	}
}

// WithWriteTimeout sets the default write timeout.
func WithWriteTimeout(d time.Duration) StdioConnOption {
	return func(c *StdioConn) {
		c.writeTimeout = d
	}
}

// WithConnAddress sets the connection's address.
func WithConnAddress(network, address string) StdioConnOption {
	return func(c *StdioConn) {
		c.addr = StdioAddr{
			network: network,
			address: address,
		}
	}
}

// LocalAddr returns the local network address, implementing net.Conn.
func (c *StdioConn) LocalAddr() net.Addr {
	return c.addr
}

// RemoteAddr returns the remote network address, implementing net.Conn.
func (c *StdioConn) RemoteAddr() net.Addr {
	return c.addr
}

// SetDeadline sets the read and write timeouts, implementing net.Conn.
func (c *StdioConn) SetDeadline(t time.Time) error {
	_ = c.SetReadDeadline(t)
	_ = c.SetWriteDeadline(t)
	return nil
}

// SetReadDeadline sets the read timeout, implementing net.Conn.
func (c *StdioConn) SetReadDeadline(t time.Time) error {
	if t.IsZero() {
		c.readTimeout = 0
	} else {
		c.readTimeout = time.Until(t)
	}
	return nil
}

// SetWriteDeadline sets the write timeout, implementing net.Conn.
func (c *StdioConn) SetWriteDeadline(t time.Time) error {
	if t.IsZero() {
		c.writeTimeout = 0
	} else {
		c.writeTimeout = time.Until(t)
	}
	return nil
}

// readLoop continuously reads from the underlying reader and sends results over readCh.
func (c *StdioConn) readLoop() {
	defer close(c.readCh)
	for {
		// Read from the underlying reader.
		buf := make([]byte, 4096)
		n, err := c.reader.Read(buf)
		res := readResult{data: buf[:n], err: err}

		// Send the result or exit if closed.
		select {
		case <-c.closed:
			return
		case c.readCh <- res:
			// If an error occurred, exit the loop.
			if err != nil {
				return
			}
		}
	}
}

// Read reads data from the connection, implementing net.Conn.
func (c *StdioConn) Read(b []byte) (int, error) {
	// Check if the connection is closed.
	select {
	case <-c.closed:
		return 0, io.EOF
	default:
		// Continue.
	}

	// Determine timeout duration.
	timeout := c.readTimeout

	if timeout == 0 {
		// No timeout.
		select {
		case res, ok := <-c.readCh:
			if !ok {
				return 0, io.EOF
			}
			n := copy(b, res.data)
			return n, res.err
		case <-c.closed:
			return 0, io.EOF
		}
	}

	// Read with timeout.
	select {
	case res, ok := <-c.readCh:
		if !ok {
			return 0, io.EOF
		}
		n := copy(b, res.data)
		return n, res.err
	case <-time.After(timeout):
		return 0, &timeoutError{op: "read", timeout: timeout}
	case <-c.closed:
		return 0, io.EOF
	}
}

// writeLoop continuously handles write requests from writeCh.
func (c *StdioConn) writeLoop() {
	defer close(c.writeCh)
	for {
		select {
		case <-c.closed:
			return
		case req, ok := <-c.writeCh:
			if !ok {
				return
			}
			if req.resCh == nil {
				// Invalid write request, skip.
				continue
			}
			// Write to the underlying writer.
			n, err := c.writer.Write(req.data)
			req.resCh <- writeResult{n: n, err: err}
			if err != nil {
				return
			}
		}
	}
}

// Write writes data to the connection, implementing net.Conn.
func (c *StdioConn) Write(b []byte) (int, error) {
	// Check if the connection is closed.
	select {
	case <-c.closed:
		return 0, io.ErrClosedPipe
	default:
		// Continue.
	}

	// Prepare the write request.
	resCh := make(chan writeResult, 1)
	// Make a copy of b to avoid data races.
	dataCopy := slices.Clone(b)
	req := writeRequest{
		data:  dataCopy,
		resCh: resCh,
	}

	// Determine timeout duration.
	timeout := c.writeTimeout

	if timeout == 0 {
		// No timeout.
		select {
		case c.writeCh <- req:
			res := <-resCh
			return res.n, res.err
		case <-c.closed:
			return 0, io.ErrClosedPipe
		}
	}

	// Write with timeout.
	select {
	case c.writeCh <- req:
		select {
		case res := <-resCh:
			return res.n, res.err
		case <-time.After(timeout):
			return 0, &timeoutError{op: "write", timeout: timeout}
		case <-c.closed:
			return 0, io.ErrClosedPipe
		}
	case <-time.After(timeout):
		return 0, &timeoutError{op: "write", timeout: timeout}
	case <-c.closed:
		return 0, io.ErrClosedPipe
	}
}

// Close closes the connection, implementing net.Conn.
func (c *StdioConn) Close() error {
	c.closeOnce.Do(func() {
		close(c.closed)
	})
	return nil
}

// timeoutError represents a timeout error.
type timeoutError struct {
	op      string
	timeout time.Duration
}

func (e *timeoutError) Error() string {
	return fmt.Sprintf("%s timeout after %s", e.op, e.timeout)
}

func (e *timeoutError) Timeout() bool {
	return true
}

func (e *timeoutError) Temporary() bool {
	return true
}
