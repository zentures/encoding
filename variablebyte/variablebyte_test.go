/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package variablebyte

import (
	"testing"
	"reflect"
	"fmt"
	"log"
	"github.com/reducedb/encoding"
)

var (
	codec encoding.Integer
	data []uint32
	size int = 100000000
)

func init() {
	codec = NewIntegratedVariableByte()
	log.Printf("variablebyte_test/init: generating %d uint32s\n", size)
	data = encoding.GenerateClustered(size, size*2)
	log.Printf("variablebyte_test/init: generated %d integers for test", size)
}

func runCompression(data []uint32, length int, codec encoding.Integer) []uint32 {
	compressed := make([]uint32, length)
	inpos := encoding.NewCursor()
	outpos := encoding.NewCursor()
	codec.Compress(data, inpos, length, compressed, outpos)
	compressed = compressed[:outpos.Get()]
	return compressed
}

func runDecompression(data []uint32, length int, codec encoding.Integer) []uint32 {
	recovered := make([]uint32, length)
	rinpos := encoding.NewCursor()
	routpos := encoding.NewCursor()
	codec.Decompress(data, rinpos, len(data), recovered, routpos)
	recovered = recovered[:routpos.Get()]
	return recovered
}

func TestBasicExample(t *testing.T) {
	for _, k := range []int{1, 13, 133, 1333, 133333, 13333333} {
		fmt.Printf("variablebyte/TestBasicExample: Testing with %d integers\n", k)

		compressed := runCompression(data, k, codec)
		fmt.Printf("variablebyte/TestBasicExample: Compressed from %d bytes to %d bytes\n", k*4, len(compressed)*4)

		recovered := runDecompression(compressed, k, codec)
		fmt.Printf("variablebyte/TestBasicExample: Decompressed from %d bytes to %d bytes\n", len(compressed)*4, len(recovered)*4)

		if !reflect.DeepEqual(data, recovered) {
			t.Fatalf("variablebyte/TestBasicExample: Problem recovering. Original length = %d, recovered length = %d\n", k, len(recovered))
		}
	}
}

func BenchmarkCompress(b *testing.B) {
	b.ResetTimer()
	compressed := runCompression(data, b.N, codec)
	b.StopTimer()

	fmt.Printf("variablebyte/BenchmarkCompress: Compressed from %d bytes to %d bytes\n", b.N*4, len(compressed)*4)
}

func BenchmarkDecompress(b *testing.B) {
	compressed := runCompression(data, b.N, codec)

	b.ResetTimer()
	recovered := runDecompression(compressed, b.N, codec)
	b.StopTimer()

	fmt.Printf("variablebyte/BenchmarkDecompress: Decompressed from %d bytes to %d bytes\n", len(compressed)*4, len(recovered)*4)
}
