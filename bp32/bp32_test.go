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
	codec encoding.Integer
	data []uint32
	size int = 1000000000
)

func init() {
	codec = NewIntegratedBP32()
	log.Printf("bp32/init: generating %d uint32s\n", size)
	data = encoding.GenerateClustered(size, size*2)
	log.Printf("bp32/init: generated %d integers for test", size)
}

func TestCodec(t *testing.T) {
	sizes := []int{128, 128*10, 128*100, 128*1000, 128*10000}
	encoding.TestCodec(codec, data, sizes, t)
}

func BenchmarkCompress(b *testing.B) {
	encoding.BenchmarkCompress(codec, data, b)
}

func BenchmarkUncompress(b *testing.B) {
	encoding.BenchmarkUncompress(codec, data, b)
}
