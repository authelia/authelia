package utils

import (
	"errors"
	"io"
)

// NewWriteCloser creates a new io.WriteCloser from an io.Writer.
func NewWriteCloser(wr io.Writer) io.WriteCloser {
	return &WriteCloser{wr: wr}
}

// WriteCloser is a io.Writer with an io.Closer.
type WriteCloser struct {
	wr io.Writer

	closed bool
}

// Write to the io.Writer.
func (w *WriteCloser) Write(p []byte) (n int, err error) {
	if w.closed {
		return -1, errors.New("already closed")
	}

	return w.wr.Write(p)
}

// Close the io.Closer.
func (w *WriteCloser) Close() error {
	w.closed = true

	return nil
}
