/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package bp32

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
	log.Printf("bp32/init: generating %d int32s\n", size)
	data = encoding.GenerateClustered(size, size*2)
	log.Printf("bp32/init: generated %d integers for test", size)
}

func TestBP32(t *testing.T) {
	sizes := []int{128, 128*10, 128*100, 128*1000, 128*10000}
	encoding.TestCodec(NewBP32(), data, sizes, t)
}

func BenchmarkBP32Compress(b *testing.B) {
	encoding.BenchmarkCompress(NewBP32(), data, b)
}

func BenchmarkBP32Uncompress(b *testing.B) {
	encoding.BenchmarkUncompress(NewBP32(), data, b)
}
