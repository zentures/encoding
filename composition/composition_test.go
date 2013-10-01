/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package composition

import (
	"testing"
	"log"
	"github.com/reducedb/encoding"
	"github.com/reducedb/encoding/bp32"
	"github.com/reducedb/encoding/variablebyte"
)

var (
	codec encoding.Integer
	data []uint32
	size int = 1000000000
)

func init() {
	codec = NewIntegratedComposition(bp32.NewIntegratedBP32(), variablebyte.NewIntegratedVariableByte())
	log.Printf("composition_test/init: generating %d uint32s\n", size)
	data = encoding.GenerateClustered(size, size*2)
	log.Printf("composition_test/init: generated %d integers for test", size)
}

func TestCodec(t *testing.T) {
	sizes := []int{100, 100*10, 100*100, 100*1000, 100*10000}
	encoding.TestCodec(codec, data, sizes, t)
}

func BenchmarkCompress(b *testing.B) {
	encoding.BenchmarkCompress(codec, data, b)
}

func BenchmarkUncompress(b *testing.B) {
	encoding.BenchmarkUncompress(codec, data, b)
}
