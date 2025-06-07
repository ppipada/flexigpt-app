package mcpstdio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// RequestParams holds default parameters for HTTP requests.
type RequestParams struct {
	Method string
	URL    string
	Header http.Header
}

// ResponseWriter implements http.ResponseWriter.
type ResponseWriter struct {
	writer      io.Writer
	writeMu     *sync.Mutex
	header      http.Header
	statusCode  int
	wroteHeader bool
}

// Header returns the header map to be sent by WriteHeader.
func (w *ResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

// WriteHeader sends an HTTP response header with the provided status code.
func (w *ResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		// Avoid multiple WriteHeader calls.
		return
	}
	w.statusCode = statusCode
	w.wroteHeader = true
	// We don't output the status code or headers in this.
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *ResponseWriter) Write(b []byte) (int, error) {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()

	// Write to the connection.
	return w.writer.Write(b)
}

// HTTPMessageHandler uses an http.Handler to process messages.
type HTTPMessageHandler struct {
	Handler       http.Handler
	RequestParams RequestParams
}

// NewHTTPMessageHandler creates a new HTTPMessageHandler.
func NewHTTPMessageHandler(handler http.Handler, params RequestParams) *HTTPMessageHandler {
	// Set default values if not provided.
	if params.Method == "" {
		params.Method = "POST"
	}
	if params.URL == "" {
		params.URL = "/"
	}
	if params.Header == nil {
		params.Header = make(http.Header)
	}
	return &HTTPMessageHandler{
		Handler:       handler,
		RequestParams: params,
	}
}

// HandleMessage processes a single message.
func (h *HTTPMessageHandler) HandleMessage(writer io.Writer, msg []byte) {
	// Log.Printf("MSG: %s", string(msg))
	// Create a ResponseWriter for this handler.
	w := &ResponseWriter{
		writer:  writer,
		writeMu: &sync.Mutex{},
	}

	// Create Request with the message as the body.
	req, err := http.NewRequestWithContext(
		context.Background(),
		h.RequestParams.Method,
		h.RequestParams.URL,
		bytes.NewReader(msg),
	)
	if err != nil {
		// Log the error and return.
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		return
	}

	// Copy default headers.
	for k, v := range h.RequestParams.Header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	// Handle the request.
	h.Handler.ServeHTTP(w, req)
}
