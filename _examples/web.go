// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package main

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/Thomasdezeeuw/logger"
)

var log *logger.Logger

func init() {
	// create a new logger
	l, err := logger.NewConsole("App")
	if err != nil {
		panic(err)
	}
	log = l
}

func main() {
	// IMPORTANT, otherwise not all logs will be written!
	defer log.Close()

	// Create a new mux add our endpoints to the mux.
	mux := http.NewServeMux()
	mux.Handle("/", identifier(http.HandlerFunc(home)))
	mux.Handle("/other", identifier(http.HandlerFunc(other)))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}

// A simple endpoint displaying our home page.
func home(w http.ResponseWriter, r *http.Request) {
	// Log a message with the request id.
	log.Info(logger.Tags{"home", r.Header.Get("X-REQUEST-ID")}, "Request to home page")

	w.Write([]byte("Welcome to the homepage"))
}

// A simple endpoint displaying our home page.
func other(w http.ResponseWriter, r *http.Request) {
	// Log a message with the request id.
	log.Info(logger.Tags{"other", r.Header.Get("X-REQUEST-ID")}, "Request to other page")

	w.Write([]byte("Welcome to our other page"))
}

// Create a new id for every request and add it to the request header.
func identifier(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a new request id based on the URL and a random number.
		// With these request id we can track all the logs from a single request.
		idString := r.URL.Path + strconv.Itoa(rand.Int())
		requestID := fmt.Sprintf("%x", sha1.Sum([]byte(idString)))
		r.Header.Set("X-REQUEST-ID", requestID)

		// Log a message with the newly created id.
		log.Info(logger.Tags{"identifier", requestID}, "Created a request id")

		next.ServeHTTP(w, r)
	})
}
