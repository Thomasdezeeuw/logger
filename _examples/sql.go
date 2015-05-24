// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"errors"

	"github.com/Thomasdezeeuw/logger"
	_ "github.com/go-sql-driver/mysql"
)

// SQL Logs table:
// id 		INT AUTO_INCREMENT
// date		DATETIME
// level	ENUM("FATAL", "ERROR", "INFO", "DEBUG", "THUMB")
// tags		VARCHAR
// msg		VARCHAR

var queryStr = "INSERT INTO Logs (date, level, tags, msg) VALUES (?, ?, ?, ?)"

type sqlMsgWriter struct {
	query *sql.Stmt
}

func (sql *sqlMsgWriter) Write(msg logger.Msg) error {
	_, err := sql.query.Exec(msg.Timestamp, msg.Level, msg.Tags.String(), msg.Msg)
	if err != nil {
		// It might be usefull to have some sort of backup log, like a file or
		// stdout to catch these kind of errors (altough they should be caught with
		// good testing and a dependable database).
		log.Error(logger.Tags{"sql.go", "sqlMsgWriter.Write"}, err)
	}
	return err
}

func (sql *sqlMsgWriter) Close() error {
	return sql.query.Close()
}

var log *logger.Logger

func init() {
	// Connect to the database.
	db, err := sql.Open("mysql", "test:7Eg13uve@tcp(127.0.0.1:3306)/test")
	if err != nil {
		panic(err)
	}

	// Create an prepared query aswell as our log writer.
	query, err := db.Prepare(queryStr)
	if err != nil {
		panic(err)
	}
	mw := &sqlMsgWriter{query: query}

	// Create a new logger like normal.
	log, err = logger.New("AppDB", mw)
	if err != nil {
		panic(err)
	}

	// Show debug messages.
	log.ShowDebug = true
}

func main() {
	// IMPORTANT! Otherwise the query will never be executed!
	defer log.Close()

	defer func() {
		// Log an recoverd error (panic).
		if recv := recover(); recv != nil {
			log.Fatal(logger.Tags{"sql.go", "main"}, recv)
		}
	}()

	// Log an error.
	err := doSomething("stuff")
	if err != nil {
		log.Error(logger.Tags{"sql.go", "main"}, err)
	}

	// Log an informational message.
	address := "localhost:8080"
	log.Info(logger.Tags{"sql.go", "main"}, "Listening on address %s", address)

	panic(errors.New("Oh no!"))
}

func doSomething(str string) error {
	// Log an debug message.
	log.Debug(logger.Tags{"sql.go", "doSomething"}, "doSomething(%q)", str)

	return errors.New("oops")
}

func unusedFunction() {
	// Log thumbstone, to see if the function is used in production.
	log.Thumbstone("unusedFunction in _examples/sql.go")
}
