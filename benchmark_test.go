// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

import "testing"

// todo: benchmark with different size tags: 0, 1, 2, 3, 5, 10 (20?).

var (
	benchmarkResultTagString string
	benchmarkResultTagBytes  []byte
	benchmarkResultTagJSON   []byte
)

func BenchmarkTags_String(b *testing.B) {
	b.ReportAllocs()
	var str string
	for n := 0; n < b.N; n++ {
		str = Tags{"hi", "world"}.String()
	}
	benchmarkResultTagString = str
}

func BenchmarkTags_Bytes(b *testing.B) {
	b.ReportAllocs()
	var bb []byte
	for n := 0; n < b.N; n++ {
		bb = Tags{"hi", "world"}.Bytes()
	}
	benchmarkResultTagBytes = bb
}

func BenchmarkTags_MarshalJSON(b *testing.B) {
	b.ReportAllocs()
	var json []byte
	for n := 0; n < b.N; n++ {
		json, _ = Tags{"hi", "world"}.MarshalJSON()
	}
	benchmarkResultTagJSON = json
}
