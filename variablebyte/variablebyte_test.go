/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package variablebyte

import (
	"testing"
	"log"
	"github.com/reducedb/encoding"
)

var (
	data []int32
	size int = 1000000000
)

func init() {
	log.Printf("variablebyte_test/init: generating %d uint32s\n", size)
	data = encoding.GenerateClustered(size, size*2)
	log.Printf("variablebyte_test/init: generated %d integers for test", size)
}

func TestVariableByte(t *testing.T) {
	sizes := []int{100, 100*10, 100*100, 100*1000, 100*10000}
	encoding.TestCodec(NewVariableByte(), data, sizes, t)
}

func BenchmarkVariableByteCompress(b *testing.B) {
	encoding.BenchmarkCompress(NewVariableByte(), data, b)
}

func BenchmarkVariableByteUncompress(b *testing.B) {
	encoding.BenchmarkUncompress(NewVariableByte(), data, b)
}
