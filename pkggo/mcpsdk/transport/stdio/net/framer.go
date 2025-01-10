package net

import (
	"bufio"
	"bytes"
)

// MessageFramer defines how messages are read from a stream
type MessageFramer interface {
	WriteMessage(w *bufio.Writer, msg []byte) error
	ReadMessage(r *bufio.Reader) ([]byte, error)
}

// LineFramer frames messages delimited by newline characters.
type LineFramer struct{}

// WriteMessage writes a message with a newline delimiter.
func (f *LineFramer) WriteMessage(w *bufio.Writer, msg []byte) error {
	if !bytes.HasSuffix(msg, []byte("\n")) {
		msg = append(msg, '\n')
	}
	_, err := w.Write(msg)
	return err
}

// ReadMessage reads a message up to the next newline.
func (f *LineFramer) ReadMessage(r *bufio.Reader) ([]byte, error) {
	b, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	b = bytes.TrimSuffix(b, []byte("\n"))
	return b, nil
}
