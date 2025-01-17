package net

import (
	"bufio"
	"bytes"
	"errors"
)

// MessageFramer defines how messages are read from a stream.
type MessageFramer interface {
	WriteMessage(w *bufio.Writer, msg []byte) error
	ReadMessage(r *bufio.Reader) ([]byte, error)
}

// LineFramer frames messages delimited by newline characters.
type LineFramer struct{}

// WriteMessage writes a message with a newline delimiter.
func (f *LineFramer) WriteMessage(w *bufio.Writer, msg []byte) error {
	c := bytes.TrimSuffix(msg, []byte("\n"))
	if bytes.Contains(c, []byte("\n")) {
		return errors.New("invalid character newline in the middle")
	}
	if !bytes.HasSuffix(msg, []byte("\n")) {
		msg = append(msg, '\n')
	}
	totalWritten := 0
	for totalWritten < len(msg) {
		n, err := w.Write(msg[totalWritten:])
		if err != nil {
			return err
		}
		totalWritten += n
	}
	return nil
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
