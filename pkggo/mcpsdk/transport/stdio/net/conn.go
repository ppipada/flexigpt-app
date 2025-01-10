package net

import (
	"io"
	"net"
	"sync"
	"time"
)

// StdioConn implements net.Conn over io.Reader and io.Writer
type StdioConn struct {
	closed     chan struct{} // Channel to signal connection closed
	closeOnce  sync.Once     // Ensures Close only runs once
	reader     io.Reader     // Underlying reader
	writer     io.Writer     // Underlying writer
	readDead   time.Time     // Read deadline
	writeDead  time.Time     // Write deadline
	deadlineMu sync.Mutex    // Mutex to protect readDead and writeDead
	writeMutex sync.Mutex    // Mutex to synchronize writes
}

// NewStdioConn creates a new StdioConn wrapping provided io.Reader and io.Writer
func NewStdioConn(r io.Reader, w io.Writer) *StdioConn {
	return &StdioConn{
		closed: make(chan struct{}),
		reader: r,
		writer: w,
	}
}

// Read reads data from the connection, implementing net.Conn
func (c *StdioConn) Read(b []byte) (int, error) {
	// Lock to safely read the read deadline
	c.deadlineMu.Lock()
	if !c.readDead.IsZero() && time.Now().After(c.readDead) {
		c.deadlineMu.Unlock()
		return 0, io.EOF
	}
	c.deadlineMu.Unlock()

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
	// Lock to safely write the write deadline and check if it's passed
	c.deadlineMu.Lock()
	if !c.writeDead.IsZero() && time.Now().After(c.writeDead) {
		c.deadlineMu.Unlock()
		return 0, io.EOF
	}
	c.deadlineMu.Unlock()

	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	// Check if the connection is closed
	select {
	case <-c.closed:
		return 0, io.ErrClosedPipe
	default:
		// Continue writing
	}

	// Write data from the buffer
	n, err := c.writer.Write(b)
	return n, err
}

// Close closes the connection, implementing net.Conn
func (c *StdioConn) Close() error {
	// Ensure Close is only called once
	var err error
	c.closeOnce.Do(func() {
		close(c.closed)
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
	connCreator func() net.Conn // A function to create new connection
	conn        net.Conn        // The connection to accept
	closed      chan struct{}   // Channel to signal listener closed
	closeMu     sync.Mutex      // Protects close operations
	acceptMu    sync.Mutex      // Ensures Accept is serialized
}

// NewStdioListener creates a new StdioListener
func NewStdioListener(connCreator func() net.Conn) *StdioListener {
	return &StdioListener{
		connCreator: connCreator,
		conn:        connCreator(),
		closed:      make(chan struct{}),
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
		l.conn = l.connCreator()
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
