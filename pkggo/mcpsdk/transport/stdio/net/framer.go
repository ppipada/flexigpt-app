package net

import (
	"bufio"
	"bytes"
	"io"
)

// MessageFramer defines how messages are read from a stream
type MessageFramer interface {
	WriteMessage(w io.Writer, msg []byte) error
	ReadMessage(r io.Reader) ([]byte, error)
}

// LineFramer frames messages delimited by newline characters.
type LineFramer struct{}

// WriteMessage writes a message with a newline delimiter.
func (f *LineFramer) WriteMessage(w io.Writer, msg []byte) error {
	if !bytes.HasSuffix(msg, []byte("\n")) {
		msg = append(msg, '\n')
	}
	_, err := w.Write(msg)
	return err
}

// ReadMessage reads a message up to the next newline.
func (f *LineFramer) ReadMessage(r io.Reader) ([]byte, error) {
	reader := bufio.NewReader(r)
	return reader.ReadBytes('\n')
}
