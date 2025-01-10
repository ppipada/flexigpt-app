package net

import (
	"errors"
	"net"
	"sync"
	"time"
)

// AssignID assigns a request ID to the message and returns the updated message.
// It returns a pointer to a string as the request ID, which can be nil.
type AssignID func([]byte) (*string, []byte, error)

// ExtractID extracts the request ID from the message.
// It returns a pointer to a string as the request ID, which can be nil.
type ExtractID func([]byte) (*string, []byte, error)

// DeadLetterItem represents an item in the dead letter (error) queue.
type DeadLetterItem struct {
	Response []byte
	Err      error
}

// ClientOption configures the client
type ClientOption func(*Client)

// WithRequestIDFunctions allows assigning custom ID functions.
// If assignID returns a nil request ID, the request is sent but not tracked.
// If extractID returns a nil request ID, the response is dropped.
// If extractID returns an error, the response is added to the dead letter queue.
func WithRequestIDFunctions(
	assignID AssignID,
	extractID ExtractID,
) ClientOption {
	return func(c *Client) {
		c.assignID = assignID
		c.extractID = extractID
		c.concurrencyEnabled = true
	}
}

// WithDeadLetterQueue configures the dead letter queue with a given capacity.
// The default capacity is 4096. You can only increase the capacity.
func WithDeadLetterQueue(capacity int) ClientOption {
	return func(c *Client) {
		if capacity < 4096 {
			capacity = 4096 // Ensure minimum capacity of 4096
		}
		c.deadLetters = make(chan DeadLetterItem, capacity)
	}
}

// WithRequestTimeout sets the request timeout duration.
// The default timeout is 1 minute.
func WithRequestTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.requestTimeout = timeout
	}
}

// Client represents the client structure.
type Client struct {
	conn               net.Conn
	framer             MessageFramer
	pending            map[string]chan []byte
	pendingMu          sync.Mutex
	assignID           AssignID
	extractID          ExtractID
	concurrencyEnabled bool
	deadLetters        chan DeadLetterItem
	done               chan struct{}
	wg                 sync.WaitGroup
	requestTimeout     time.Duration
}

// NewClient creates a new client with the provided net.Conn and MessageFramer.
// You can configure the client using ClientOptions.
func NewClient(conn net.Conn, framer MessageFramer, options ...ClientOption) *Client {
	client := &Client{
		conn:           conn,
		framer:         framer,
		pending:        make(map[string]chan []byte),
		deadLetters:    make(chan DeadLetterItem, 4096), // Default capacity 4096
		requestTimeout: time.Minute,                     // Default timeout of 1 minute
		done:           make(chan struct{}),
	}
	// Apply options
	for _, opt := range options {
		opt(client)
	}
	// Start the receiver goroutine
	client.wg.Add(1)
	go client.receiver()
	return client
}

// Send sends a message and waits for a response.
// If the assignID function returns a nil request ID, the request is sent but the response is not tracked.
// If concurrency is disabled, Send operates synchronously.
func (c *Client) Send(msg []byte) ([]byte, error) {
	if !c.concurrencyEnabled {
		// No request ID functions provided, do synchronous send
		err := c.framer.WriteMessage(c.conn, msg)
		if err != nil {
			return nil, err
		}
		// Read response
		resp, err := c.framer.ReadMessage(c.conn)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	// Assign a request ID
	reqIDPtr, msgWithID, err := c.assignID(msg)
	if err != nil {
		return nil, err
	}

	// Send the message
	err = c.framer.WriteMessage(c.conn, msgWithID)
	if err != nil {
		return nil, err
	}

	if reqIDPtr == nil {
		// No request ID, don't track responses
		return nil, nil
	}

	reqID := *reqIDPtr
	responseCh := make(chan []byte, 1)
	c.pendingMu.Lock()
	c.pending[reqID] = responseCh
	c.pendingMu.Unlock()

	// Wait for response or timeout
	select {
	case resp := <-responseCh:
		return resp, nil
	case <-c.done:
		return nil, errors.New("client closed")
	case <-time.After(c.requestTimeout):
		// Timeout
		// Clean up pending
		c.pendingMu.Lock()
		delete(c.pending, reqID)
		c.pendingMu.Unlock()
		return nil, errors.New("request timed out")
	}
}

// receiver reads messages from the connection and dispatches them.
func (c *Client) receiver() {
	defer c.wg.Done()
	for {
		select {
		case <-c.done:
			return
		default:
			resp, err := c.framer.ReadMessage(c.conn)
			if err != nil {
				// Handle error, add to dead letter queue
				c.addToDeadLetter(DeadLetterItem{Response: resp, Err: err})
				continue
			}
			if !c.concurrencyEnabled {
				// No request ID functions, drop the response (or handle as needed)
				continue
			}
			// Extract request ID
			reqIDPtr, respWithoutID, err := c.extractID(resp)
			if err != nil {
				// Handle error, add to dead letter queue
				c.addToDeadLetter(DeadLetterItem{Response: resp, Err: err})
				continue
			}
			if reqIDPtr == nil {
				// Request ID is nil, drop the response
				continue
			}
			reqID := *reqIDPtr

			// Deliver response to waiting request
			c.pendingMu.Lock()
			ch, ok := c.pending[reqID]
			if ok {
				delete(c.pending, reqID)
				c.pendingMu.Unlock()
				ch <- respWithoutID
			} else {
				c.pendingMu.Unlock()
				// No pending request, add to dead letter queue
				c.addToDeadLetter(DeadLetterItem{Response: resp, Err: errors.New("no pending request for response")})
			}
		}
	}
}

// addToDeadLetter adds an item to the dead letter queue without blocking.
func (c *Client) addToDeadLetter(item DeadLetterItem) {
	select {
	case c.deadLetters <- item:
		// Added to queue
	default:
		// Dead letter queue is full, drop the item
	}
}

// Close closes the client and cleans up resources.
func (c *Client) Close() error {
	close(c.done)
	c.conn.Close()
	c.wg.Wait()
	return nil
}

// PopDeadLetter pops a message from the dead letter (error) queue.
func (c *Client) PopDeadLetter() (DeadLetterItem, error) {
	select {
	case item := <-c.deadLetters:
		return item, nil
	case <-time.After(time.Second):
		return DeadLetterItem{}, errors.New("no messages in dead letter queue")
	}
}
