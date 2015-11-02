// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

// Package logger provides asynchronous logging for Go. It is build for
// customisation and speed. It uses a custom EventWriter so any custom backend
// can be used to store the logs. Logger provides multiple ways to log
// information with different levels of importance.
//
// Each loggingoperation makes a single call to the EventWriter's Write method,
// but not necessarily at the same time as Log operation is called, since the
// logging is done asynchronously. The logger package can used simultaneously
// from multiple goroutines.
//
// Because the logger package is asynchronous Close musted be called before the
// program exits, this way logger will make sure all log event will be written.
//
// By default there are six different event types (from lower to higher): debug,
// info, warn, error, fatal and thumb. But new event types can be created, to be
// used in the custom EventWriter.
package logger
