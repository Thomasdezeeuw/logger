package logger

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
)

const (
	defaultFileFlag       = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	defaultFilePermission = 0644
)

type fileMsgWriter struct {
	w *bufio.Writer
	f *os.File
}

func (fw *fileMsgWriter) Write(msg Msg) error {
	bytes := append(msg.Bytes(), '\n')
	n, err := fw.w.Write(bytes)
	if err != nil {
		return err
	} else if n != len(bytes) {
		return io.ErrShortWrite
	}
	return nil
}

func (fw *fileMsgWriter) Close() error {
	flushErr := fw.w.Flush()
	err := fw.f.Close()
	if err == nil {
		err = flushErr
	}
	return err
}

// NewFile creates a new logger that writes to the given file.
func NewFile(name, path string) (*Logger, error) {
	f, err := os.OpenFile(path, defaultFileFlag, defaultFilePermission)
	if err != nil {
		return nil, err
	}

	mw := &fileMsgWriter{bufio.NewWriter(f), f}
	return New(name, mw)
}

type ioWriterMsgWriter struct {
	w io.Writer
}

func (iw *ioWriterMsgWriter) Write(msg Msg) error {
	bytes := append(msg.Bytes(), '\n')
	n, err := iw.w.Write(bytes)
	if err != nil {
		return err
	} else if n != len(bytes) {
		return io.ErrShortWrite
	}
	return nil
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

type jsonWriterMsgWriter struct {
	enc *json.Encoder
	bw  *bufio.Writer
}

func (jw *jsonWriterMsgWriter) Write(msg Msg) error {
	return jw.enc.Encode(msg)
}

func (jw *jsonWriterMsgWriter) Close() error {
	return jw.bw.Flush()
}

// NewJSON creates a new logger that writes logs in a JSON format to the given
// io.Writer.
func NewJSON(name string, w io.Writer) (*Logger, error) {
	bw := bufio.NewWriter(w)
	enc := json.NewEncoder(bw)
	mw := &jsonWriterMsgWriter{enc, bw}
	return New(name, mw)
}
