package main

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"

	"github.com/Thomasdezeeuw/logger"
	_ "github.com/go-sql-driver/mysql"
)

// SQL Logs table:
// id 		INT AUTO_INCREMENT
// date		DATETIME
// level	ENUM("FATAL", "ERROR", "INFO", "DEBUG")
// tags		VARCHAR
// msg		VARCHAR

var queryStr = "INSERT INTO Logs (date, level, tags, msg) VALUES (?, ?, ?, ?)"

type sqlWriter struct {
	query *sql.Stmt
}

func (sql *sqlWriter) Write(b []byte) (int, error) {
	// The message will always be in the following format:
	// YYYY-MM-DD HH:MM:SS [LEVEL] tag1, tag2...: msg
	// The date, time and level will always be of the same length.
	// The tags and message however are not.

	date := string(b[:19])
	lvl := strings.TrimSpace(string(b[21:26])) // INFO has an extra space.

	bb := bytes.SplitN(b[27:], []byte(":"), 2)
	tags := strings.TrimSpace(string(bb[0]))
	msg := strings.TrimSpace(string(bb[1]))

	_, err := sql.query.Exec(date, lvl, tags, msg)
	if err != nil {
		// It might be usefull to have some sort of backup log, like a file or
		// stdout to catch these kind of errors (altough they should be caught with
		// good testing).
		log.Error(logger.Tags{"sql.go", "sqlWriter.Write"}, err)
	}
	return 0, err
}

var log *logger.Logger

func init() {
	// Connect to the database.
	db, err := sql.Open("mysql", "test:password@tcp(127.0.0.1:3306)/test")
	if err != nil {
		panic(err)
	}

	// Show debug messages.
	log.ShowDebug = true

	// Create an prepared query aswell as our log writer.
	query, err := db.Prepare(queryStr)
	if err != nil {
		panic(err)
	}
	w := &sqlWriter{query: query}

	// Create a new logger like normal.
	log, err = logger.New("AppDB", 1024, w)
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

	panic(errors.New("Oh no!"))
}

func doSomething(str string) error {
	// Log an debug message.
	log.Debug(logger.Tags{"file.go", "doSomething"}, "doSomething(%q)", str)

	return errors.New("oops")
}
