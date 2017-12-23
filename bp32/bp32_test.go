/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package bp32

import (
	"log"
	"testing"

	"github.com/dataence/encoding/benchtools"
	"github.com/dataence/encoding/cursor"
	"github.com/dataence/encoding/generators"
)

var (
	data []int32
	size int = 128000
)

func init() {
	log.Printf("bp32/init: generating %d int32s\n", size)
	data = generators.GenerateClustered(size, size*2)
	log.Printf("bp32/init: generated %d integers for test", size)
}

func TestCodec(t *testing.T) {
	sizes := []int{128, 128 * 10, 128 * 100, 128 * 1000}
	benchtools.TestCodec(New(), data, sizes)
}

// go test -bench=Decode
func BenchmarkDecode(b *testing.B) {
	b.StopTimer()
	length := 128 * 1024
	data := generators.GenerateClustered(length, 1<<24)
	compdata := make([]int32, 2*length)
	recov := make([]int32, length)
	inpos := cursor.New()
	outpos := cursor.New()
	codec := New()
	codec.Compress(data, inpos, len(data), compdata, outpos)
	b.StartTimer()
	for j := 0; j < b.N; j++ {
		newinpos := cursor.New()
		newoutpos := cursor.New()
		codec.Uncompress(compdata, newinpos, outpos.Get()-newinpos.Get(), recov, newoutpos)
	}
}
