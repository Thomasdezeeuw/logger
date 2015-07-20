package main

import (
	"io"
	"os"

	"github.com/Thomasdezeeuw/logger"
)

// todo: add example with SetMinLogLevel, custom LogLevel and custom Msg.Data
// to display how to use it. Converting the custom data based on LogLevel.

var requestLevel = logger.NewLogLevel("Request")

type msgWriter struct {
	w io.Writer
}

func (mw *msgWriter) Write(msg logger.Msg) error {
	if msg.Level == requestLevel {
		// Bases on our custom LogLevel we can does something else with the custom
		// data.
		httpRequest, ok := msg.Data.(HTTPRequest)
		if !ok {
			// This should never happen.
			panic("Can't convert Request log data to a HTTPRequest")
		}

		str := "Got request " + httpRequest.Method + " for " + httpRequest.URL + "\n"
		return mw.write([]byte(str))
	}

	bytes := append(msg.Bytes(), '\n')
	return mw.write(bytes)
}

func (mw *msgWriter) write(bytes []byte) error {
	n, err := mw.w.Write(bytes)
	if err != nil {
		return err
	} else if n != len(bytes) {
		return io.ErrShortWrite
	}
	return nil
}

func (mw *msgWriter) Close() error {
	return nil
}

type HTTPRequest struct {
	URL    string
	Method string
}

func main() {
	mw := msgWriter{os.Stdout}
	log, err := logger.New("App", &mw)
	if err != nil {
		panic(err)
	}
	defer log.Close()

	msg := logger.Msg{
		Level: requestLevel,
		Tags:  logger.Tags{"custom.go"},
		// The timestamp gets set by log.Message
		Data: HTTPRequest{"/url", "GET"},
	}

	log.Message(msg)
	log.Info(logger.Tags{}, "Hello world")
}
