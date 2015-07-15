package logger

import (
	"io"
	"os"
)

type ioWriterMsgWriter struct {
	w io.Writer
}

func (iw *ioWriterMsgWriter) Write(msg Msg) error {
	// todo: check if length of the write mathces the length of the message bytes.
	bytes := append(msg.Bytes(), '\n')
	_, err := iw.w.Write(bytes)
	return err
}

func (iw *ioWriterMsgWriter) Close() error {
	return nil
}

// NewWriter creates a new logger that writes to the given io.Writer.
func NewWriter(name string, w io.Writer) (*Logger, error) {
	mw := &ioWriterMsgWriter{w}
	return New(name, mw)
}

// Error ouput, usefull for testing.
var stderr io.Writer = os.Stderr

// NewConsole creates a new logger that writes to error output (os.Stderr).
func NewConsole(name string) (*Logger, error) {
	return NewWriter(name, stderr)
}
