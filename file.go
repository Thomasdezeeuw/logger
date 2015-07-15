package logger

import (
	"bufio"
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
