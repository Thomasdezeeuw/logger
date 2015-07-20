// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Thomasdezeeuw/logger"
	_ "github.com/go-sql-driver/mysql"
)

// SQL Logs table:
//	CREATE TABLE IF NOT EXISTS `logs` (
//		`id` int(10) unsigned NOT NULL AUTO_INCREMENT,
//		`date` datetime NOT NULL,
//		`level` enum('FATAL','ERROR','INFO','DEBUG','THUMB') NOT NULL,
//		`tags` varchar(100) NOT NULL,
//		`msg` varchar(200) NOT NULL,
//		`data` varchar(500) DEFAULT NULL,
//		PRIMARY KEY (`id`)
//	);

var queryStr = "INSERT INTO Logs (date, level, tags, msg, data) VALUES (?, ?, ?, ?, ?)"

type sqlMsgWriter struct {
	query *sql.Stmt
}

func (sql *sqlMsgWriter) Write(msg logger.Msg) error {
	var dataStr string
	if msg.Level == logger.Fatal {
		dataBytes, ok := msg.Data.([]byte)
		if ok {
			dataStr = string(dataBytes)
		}
	}

	result, err := sql.query.Exec(msg.Timestamp.UTC(), msg.Level.String(),
		msg.Tags.String(), msg.Msg, dataStr)
	if err != nil {
		// It might be usefull to have some sort of backup log, like a file or
		// stdout to catch these kind of errors (altough they should be caught with
		// good testing and a dependable database).
		log.Error(logger.Tags{"sql.go", "sqlMsgWriter.Write"}, err)
	} else if n, err := result.RowsAffected(); err == nil && n != 1 {
		err := fmt.Errorf("Wanted to create a single log entry, but created %d", n)
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

	unusedFunction()
}

func doSomething(str string) error {
	// Log an debug message.
	log.Debug(logger.Tags{"sql.go", "doSomething"}, "doSomething(%q)", str)

	return errors.New("oops")
}

func unusedFunction() {
	// Log thumbstone, to see if the function is used in production.
	log.Thumbstone(logger.Tags{"sql.go"}, "unusedFunction")

	panic(errors.New("Oh no!"))
}
