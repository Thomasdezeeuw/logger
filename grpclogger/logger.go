// Copyright (C) 2015-2016 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

// Package grpclogger creates a logger interface to be used in grpc logger
// package (google.golang.org/grpc/grpclog). For more information on grpc see
// http://www.grpc.io, for grpc-go see https://github.com/grpc/grpc-go.
package grpclogger

import (
	"errors"
	"fmt"
	"os"

	"github.com/Thomasdezeeuw/logger"
	"github.com/Thomasdezeeuw/logger/internal/util"
	"google.golang.org/grpc/grpclog"
)

type log struct {
	tags    logger.Tags
	closeFn func()
}

func (log *log) Fatal(args ...interface{}) {
	msg := util.InterfacesToString(args)
	logger.Fatal(log.tags, msg)
	exit(log.closeFn)
}

func (log *log) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Fatal(msg)
}

func (log *log) Fatalln(args ...interface{}) {
	log.Fatal(args...)
}

func (log *log) Print(args ...interface{}) {
	msg := util.InterfacesToString(args)
	logger.Error(log.tags, errors.New(msg))
}

func (log *log) Printf(format string, args ...interface{}) {
	logger.Errorf(log.tags, format, args...)
}

func (log *log) Println(args ...interface{}) {
	log.Print(args...)
}

// Stubbed for testing.
var osExit = os.Exit

// Hate to do this, but it is what the default log package does.
// Stubbed for testing.
var exit = func(closeFn func()) {
	closeFn()
	osExit(1)
}

// CreateLogger creates a new logger that can be used in grpc/grpclog. It
// follows the default log package style of logging and therefore it doesn't
// have log levels (EventType). Because of this all calls to Print* will call
// logger.Error, hoping that grpclogger will only log errors in production.
//
// Another shortcoming of the Logger interface defined by grpclog is that uses
// an empty interface a lot. We combat this by trying to make this into a single
// string using the fmt package. As result of this the message might not look
// very pretty, but it will contain all the provided information.
//
// A third point is that a call to Fatal* in the builtin log package calls to
// os.Exit, which closes the application immediately without running deffered
// statements. To combat that we accept a close function which runs before the
// call to os.Exit. In this function logger.Close must be called by the user.
func CreateLogger(tags logger.Tags, closeFn func()) grpclog.Logger {
	return &log{tags, closeFn}
}
