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
	"github.com/reducedb/encoding"
)

var (
	codec encoding.Integer
)

func init() {
	codec = NewIntegratedBP32()
}

func generateData(N int) []uint32 {
	data := encoding.GenerateClustered(N, N*2)

	fmt.Printf("variablebyte/generateData: len(data) = %d\n", len(data))

	return data
}

func runCompression(data []uint32, codec encoding.Integer) []uint32 {
	compressed := make([]uint32, len(data)*2)
	inpos := encoding.NewCursor()
	outpos := encoding.NewCursor()
	codec.Compress(data, inpos, len(data), compressed, outpos)
	compressed = compressed[:outpos.Get()]
	return compressed
}

func runDecompression(data []uint32, codec encoding.Integer) []uint32 {
	recovered := make([]uint32, len(data)*20)
	rinpos := encoding.NewCursor()
	routpos := encoding.NewCursor()
	codec.Uncompress(data, rinpos, len(data), recovered, routpos)
	recovered = recovered[:routpos.Get()]
	return recovered
}

func TestBasicExample(t *testing.T) {
	for _, k := range []int{128, 128*10, 128*100, 128*1000, 128*10000} {
		data := generateData(k)

		compressed := runCompression(data, codec)
		fmt.Printf("bp32/TestBasicExample: Compressed from %d bytes to %d bytes\n", len(data)*4, len(compressed)*4)

		recovered := runDecompression(compressed, codec)
		fmt.Printf("bp32/TestBasicExample: Decompressed from %d bytes to %d bytes\n", len(compressed)*4, len(recovered)*4)

		if !reflect.DeepEqual(data, recovered) {
			t.Fatalf("bp32/TestBasicExample: Problem recovering. Original length = %d, recovered length = %d\n     data = %v\nrecovered = %v\n", len(data), len(recovered), data, recovered)
		}
	}
}

func BenchmarkCompress(b *testing.B) {
	N := encoding.CeilBy(uint32(b.N), 128)
	data := generateData(int(N))

	b.ResetTimer()
	compressed := runCompression(data, codec)
	b.StopTimer()

	fmt.Printf("bp32/BenchmarkCompress: Compressed from %d bytes to %d bytes\n", len(data)*4, len(compressed)*4)
}

func BenchmarkDecompress(b *testing.B) {
	N := encoding.CeilBy(uint32(b.N), 128)
	data := generateData(int(N))
	compressed := runCompression(data, codec)

	b.ResetTimer()
	recovered := runDecompression(compressed, codec)
	b.StopTimer()

	fmt.Printf("bp32/BenchmarkDecompress: Decompressed from %d bytes to %d bytes\n", len(compressed)*4, len(recovered)*4)
}
