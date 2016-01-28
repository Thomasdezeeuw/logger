// Copyright (C) 2015-2016 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

// Package logger provides asynchronous logging. It is build for customisation
// and speed. It uses a custom EventWriter so any custom backend can be used to
// store the logs. Logger provides multiple ways to log information with
// different levels of importance.
//
// Each logging operation makes a single call to the EventWriter's Write method,
// but not necessarily at the same time as Log operation is called, since the
// logging is done asynchronously. The logger package can used simultaneously
// from multiple goroutines.
//
// Because the logger package is asynchronous Close must be called before the
// program exits, this way logger will make sure all log event will be written.
// After Close is called all calls to any log operation will panic. This is
// because internally the logger package uses a channel to make the logging
// asynchronous and sending to a closed channel will panic.
//
// By default there are six different event types (from lower to higher): debug,
// info, warn, error, fatal and thumb. But new event types can be created using
// NewEventType. These can then be used in a custom EventWriter to extract data
// from Event.Data.
package logger
