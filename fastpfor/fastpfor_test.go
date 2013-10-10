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
	size int = 10000000
)

func init() {
	log.Printf("fastpfor_test/init: generating %d uint32s\n", size)
	data = encoding.GenerateClustered(size, size*2)
	log.Printf("fastpfor_test/init: generated %d integers for test", size)
}

func TestFastPFOR(t *testing.T) {
	sizes := []int{128, 128*10, 128*100, 128*1000, 128*10000}
	encoding.TestCodec(New(), data, sizes, t)
}

func BenchmarkBP32Compress(b *testing.B) {
	encoding.BenchmarkCompress(New(), data, b)
}

func BenchmarkBP32Uncompress(b *testing.B) {
	encoding.BenchmarkUncompress(New(), data, b)
}
