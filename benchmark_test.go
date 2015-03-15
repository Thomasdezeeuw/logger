// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

var result string

func BenchmarkTagString(b *testing.B) {
	b.ReportAllocs()
	var str string
	tags := Tags{"test", "test2"}
	for i := 0; i < b.N; i++ {
		str = tags.String()
	}
	result = str
}

func BenchmarkTagBytes(b *testing.B) {
	b.ReportAllocs()
	var buf []byte
	tags := Tags{"test", "test2"}
	for i := 0; i < b.N; i++ {
		buf = tags.Bytes()
	}
	result = string(buf)
}

func BenchmarkItoa2(b *testing.B) {
	b.ReportAllocs()
	var buf []byte
	for i := 0; i < b.N; i++ {
		itoa(&buf, 12, 2)
		buf = []byte{}
	}
	result = string(buf)
}

func BenchmarkItoa4(b *testing.B) {
	b.ReportAllocs()
	var buf []byte
	for i := 0; i < b.N; i++ {
		itoa(&buf, 2015, 4)
		buf = []byte{}
	}
	result = string(buf)
}

func BenchmarkFormatMsg(b *testing.B) {
	b.ReportAllocs()
	var str string
	t := time.Now()
	var lvl, msg = "ERROR", "Message"
	tags := Tags{"test", "test2"}
	for i := 0; i < b.N; i++ {
		str = formatMsg(t, lvl, tags, msg)
	}
	result = str
}

func BenchmarkFatal(b *testing.B) {
	b.ReportAllocs()
	reset()
	var buf bytes.Buffer
	log, err := New("benchmark", 1024, &buf)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	msg := "Message"
	tags := Tags{"test", "test2"}
	for i := 0; i < b.N; i++ {
		log.Fatal(tags, msg)
	}

	log.Close()
	result = buf.String()
}

func BenchmarkError(b *testing.B) {
	b.ReportAllocs()
	reset()
	var buf bytes.Buffer
	log, err := New("benchmark", 1024, &buf)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	err = errors.New("some error")
	tags := Tags{"test", "test2"}
	for i := 0; i < b.N; i++ {
		log.Error(tags, err)
	}

	log.Close()
	result = buf.String()
}

func BenchmarkInfo(b *testing.B) {
	b.ReportAllocs()
	reset()
	var buf bytes.Buffer
	log, err := New("benchmark", 1024, &buf)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	msg := "Message"
	tags := Tags{"test", "test2"}
	for i := 0; i < b.N; i++ {
		log.Info(tags, msg)
	}

	log.Close()
	result = buf.String()
}

func BenchmarkDebug(b *testing.B) {
	b.ReportAllocs()
	reset()
	var buf bytes.Buffer
	log, err := New("benchmark", 1024, &buf)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	msg := "Message"
	tags := Tags{"test", "test2"}
	for i := 0; i < b.N; i++ {
		log.Debug(tags, msg)
	}

	log.Close()
	result = buf.String()
}

func BenchmarkThumbstone(b *testing.B) {
	b.ReportAllocs()
	reset()
	var buf bytes.Buffer
	log, err := New("benchmark", 1024, &buf)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	msg := "Message"
	for i := 0; i < b.N; i++ {
		log.Thumbstone(msg)
	}

	log.Close()
	result = buf.String()
}
