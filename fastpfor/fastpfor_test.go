/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package fastpfor

import (
	"testing"
	"log"
	"github.com/reducedb/encoding"
)

var (
	data []int32
	size int = 128000000
)

func init() {
	log.Printf("fastpfor_test/init: generating %d uint32s\n", size)
	data = encoding.GenerateClustered(size, size*2)
	log.Printf("fastpfor_test/init: generated %d integers for test", size)
}

func TestFastPFOR(t *testing.T) {
	//sizes := []int{128, 128*10, 128*100, 128*1000, 128*10000}
	//log.Printf("Testing DeltaFastPFOR\n")
	//encoding.TestCodecPprof(NewDeltaFastPFOR(), data, sizes, t)
	//log.Printf("\n\nTesting FastPFOR\n")
	//encoding.TestCodec(NewFastPFOR(), data, sizes, t)
	encoding.TestCodecPprof(NewDeltaFastPFOR(), data, []int{128000000}, t)
}

func BenchmarkBP32Compress(b *testing.B) {
	encoding.BenchmarkCompress(NewFastPFOR(), data, b)
}

func BenchmarkBP32Uncompress(b *testing.B) {
	encoding.BenchmarkUncompress(NewFastPFOR(), data, b)
}
