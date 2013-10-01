/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package bp32

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
)

func init() {
	codec = NewIntegratedBP32()
	log.Printf("bp32/init: generating %d uint32s\n", size)
	data = encoding.GenerateClustered(size, size*2)
	log.Printf("bp32/init: generated %d integers for test", size)
}

func runCompression(data []uint32, length int, codec encoding.Integer) []uint32 {
	compressed := make([]uint32, length*2)
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
	codec.Uncompress(data, rinpos, len(data), recovered, routpos)
	recovered = recovered[:routpos.Get()]
	return recovered
}

func TestBasicExample(t *testing.T) {
	for _, k := range []int{128, 128*10, 128*100, 128*1000, 128*10000} {
		fmt.Printf("bp32/TestBasicExample: Testing with %d integers\n", k)

		compressed := runCompression(data, k, codec)
		fmt.Printf("bp32/TestBasicExample: Compressed from %d bytes to %d bytes\n", k*4, len(compressed)*4)

		recovered := runDecompression(compressed, k, codec)
		fmt.Printf("bp32/TestBasicExample: Decompressed from %d bytes to %d bytes\n", len(compressed)*4, len(recovered)*4)

		if !reflect.DeepEqual(data[:k], recovered) {
			t.Fatalf("bp32/TestBasicExample: Problem recovering. Original length = %d, recovered length = %d\n", k, len(recovered))
		}
	}
}

func BenchmarkCompress(b *testing.B) {
	k := int(encoding.CeilBy(uint32(b.N), 128))
	//data := generateData(int(N))

	b.ResetTimer()
	compressed := runCompression(data, k, codec)
	b.StopTimer()

	fmt.Printf("bp32/BenchmarkCompress: Compressed from %d bytes to %d bytes\n", k*4, len(compressed)*4)
}

func BenchmarkDecompress(b *testing.B) {
	k := int(encoding.CeilBy(uint32(b.N), 128))
	//data := generateData(int(N))
	compressed := runCompression(data, k, codec)

	b.ResetTimer()
	recovered := runDecompression(compressed, len(compressed), codec)
	b.StopTimer()

	fmt.Printf("bp32/BenchmarkDecompress: Decompressed from %d bytes to %d bytes\n", len(compressed)*4, len(recovered)*4)
}
