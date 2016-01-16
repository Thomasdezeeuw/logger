// Copyright (C) 2015-2016 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

import "testing"

// go test -run none -bench . -benchmem -benchtime 5s -timeout 15m

var (
	benchmarkResultTagString string
	benchmarkResultTagBytes  []byte
	benchmarkResultTagJSON   []byte
)

var (
	tag0  = Tags{}
	tag1  = Tags{"tag1"}
	tag2  = Tags{"tag1", "tag2"}
	tag3  = Tags{"tag1", "tag2", "tag3"}
	tag4  = Tags{"tag1", "tag2", "tag3", "tag4"}
	tag5  = Tags{"tag1", "tag2", "tag3", "tag4", "tag5"}
	tag10 = Tags{"tag1", "tag2", "tag3", "tag4", "tag5",
		"tag6", "tag7", "tag8", "tag9", "tag10"}

	taglong1  = Tags{"very-long-tag1"}
	taglong2  = Tags{"very-long-tag1", "very-long-tag2"}
	taglong3  = Tags{"very-long-tag1", "very-long-tag2", "very-long-tag3"}
	taglong4  = Tags{"very-long-tag1", "very-long-tag2", "very-long-tag3", "very-long-tag4"}
	taglong5  = Tags{"very-long-tag1", "very-long-tag2", "very-long-tag3", "very-long-tag4", "very-long-tag5"}
	taglong10 = Tags{"very-long-tag1", "very-long-tag2", "very-long-tag3", "very-long-tag4", "very-long-tag5",
		"very-long-tag6", "very-long-tag7", "very-long-tag8", "very-long-tag9", "very-long-tag10"}
)

func BenchmarkTags_String0Tag(b *testing.B)  { benchmarkTagsString(b, tag0) }
func BenchmarkTags_String1Tag(b *testing.B)  { benchmarkTagsString(b, tag1) }
func BenchmarkTags_String2Tag(b *testing.B)  { benchmarkTagsString(b, tag2) }
func BenchmarkTags_String3Tag(b *testing.B)  { benchmarkTagsString(b, tag3) }
func BenchmarkTags_String4Tag(b *testing.B)  { benchmarkTagsString(b, tag4) }
func BenchmarkTags_String5Tag(b *testing.B)  { benchmarkTagsString(b, tag5) }
func BenchmarkTags_String10Tag(b *testing.B) { benchmarkTagsString(b, tag10) }

func BenchmarkTags_String1TagLong(b *testing.B)  { benchmarkTagsString(b, taglong1) }
func BenchmarkTags_String2TagLong(b *testing.B)  { benchmarkTagsString(b, taglong2) }
func BenchmarkTags_String3TagLong(b *testing.B)  { benchmarkTagsString(b, taglong3) }
func BenchmarkTags_String4TagLong(b *testing.B)  { benchmarkTagsString(b, taglong4) }
func BenchmarkTags_String5TagLong(b *testing.B)  { benchmarkTagsString(b, taglong5) }
func BenchmarkTags_String10TagLong(b *testing.B) { benchmarkTagsString(b, taglong10) }

func benchmarkTagsString(b *testing.B, tags Tags) {
	var str string
	for n := 0; n < b.N; n++ {
		str = tags.String()
	}
	benchmarkResultTagString = str
}

func BenchmarkTags_Bytes0Tag(b *testing.B)  { benchmarkTagsBytes(b, tag0) }
func BenchmarkTags_Bytes1Tag(b *testing.B)  { benchmarkTagsBytes(b, tag1) }
func BenchmarkTags_Bytes2Tag(b *testing.B)  { benchmarkTagsBytes(b, tag2) }
func BenchmarkTags_Bytes3Tag(b *testing.B)  { benchmarkTagsBytes(b, tag3) }
func BenchmarkTags_Bytes4Tag(b *testing.B)  { benchmarkTagsBytes(b, tag4) }
func BenchmarkTags_Bytes5Tag(b *testing.B)  { benchmarkTagsBytes(b, tag5) }
func BenchmarkTags_Bytes10Tag(b *testing.B) { benchmarkTagsBytes(b, tag10) }

func BenchmarkTags_Bytes1TagLong(b *testing.B)  { benchmarkTagsBytes(b, taglong1) }
func BenchmarkTags_Bytes2TagLong(b *testing.B)  { benchmarkTagsBytes(b, taglong2) }
func BenchmarkTags_Bytes3TagLong(b *testing.B)  { benchmarkTagsBytes(b, taglong3) }
func BenchmarkTags_Bytes4TagLong(b *testing.B)  { benchmarkTagsBytes(b, taglong4) }
func BenchmarkTags_Bytes5TagLong(b *testing.B)  { benchmarkTagsBytes(b, taglong5) }
func BenchmarkTags_Bytes10TagLong(b *testing.B) { benchmarkTagsBytes(b, taglong10) }

func benchmarkTagsBytes(b *testing.B, tags Tags) {
	var bytes []byte
	for n := 0; n < b.N; n++ {
		bytes = tags.Bytes()
	}
	benchmarkResultTagBytes = bytes
}

func BenchmarkTags_MarshalJSON0Tag(b *testing.B)  { benchmarkTagsMarshalJSON(b, tag0) }
func BenchmarkTags_MarshalJSON1Tag(b *testing.B)  { benchmarkTagsMarshalJSON(b, tag1) }
func BenchmarkTags_MarshalJSON2Tag(b *testing.B)  { benchmarkTagsMarshalJSON(b, tag2) }
func BenchmarkTags_MarshalJSON3Tag(b *testing.B)  { benchmarkTagsMarshalJSON(b, tag3) }
func BenchmarkTags_MarshalJSON4Tag(b *testing.B)  { benchmarkTagsMarshalJSON(b, tag4) }
func BenchmarkTags_MarshalJSON5Tag(b *testing.B)  { benchmarkTagsMarshalJSON(b, tag5) }
func BenchmarkTags_MarshalJSON10Tag(b *testing.B) { benchmarkTagsMarshalJSON(b, tag10) }

func BenchmarkTags_MarshalJSON1TagLong(b *testing.B)  { benchmarkTagsMarshalJSON(b, taglong1) }
func BenchmarkTags_MarshalJSON2TagLong(b *testing.B)  { benchmarkTagsMarshalJSON(b, taglong2) }
func BenchmarkTags_MarshalJSON3TagLong(b *testing.B)  { benchmarkTagsMarshalJSON(b, taglong3) }
func BenchmarkTags_MarshalJSON4TagLong(b *testing.B)  { benchmarkTagsMarshalJSON(b, taglong4) }
func BenchmarkTags_MarshalJSON5TagLong(b *testing.B)  { benchmarkTagsMarshalJSON(b, taglong5) }
func BenchmarkTags_MarshalJSON10TagLong(b *testing.B) { benchmarkTagsMarshalJSON(b, taglong10) }

func benchmarkTagsMarshalJSON(b *testing.B, tags Tags) {
	var json []byte
	for n := 0; n < b.N; n++ {
		json, _ = tags.MarshalJSON()
	}
	benchmarkResultTagJSON = json
}
