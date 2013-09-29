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
	"github.com/reducedb/encoding"
)

func generateData(N int) []uint32 {
	data := encoding.GenerateClustered(N, N*2)

	fmt.Printf("variablebyte/generateData: len(data) = %d\n", len(data))

	return data
}

func runCompression(data []uint32, codec encoding.Integer) []uint32 {
	compressed := make([]uint32, len(data))
	inpos := encoding.NewCursor()
	outpos := encoding.NewCursor()
	codec.Compress(data, inpos, len(data), compressed, outpos)
	compressed = compressed[:outpos.Get()]
	return compressed
}

func runUncompression(data []uint32, codec encoding.Integer) []uint32 {
	recovered := make([]uint32, len(data)*10)
	rinpos := encoding.NewCursor()
	routpos := encoding.NewCursor()
	codec.Uncompress(data, rinpos, len(data), recovered, routpos)
	recovered = recovered[:routpos.Get()]
	return recovered
}

func TestBasicExample(t *testing.T) {
	codec := NewIntegratedVariableByte()

	for _, k := range []int{1, 13, 133, 1333, 133333, 13333333} {
		data := generateData(k)

		compressed := runCompression(data, codec)
		fmt.Printf("variablebyte/TestBasicExample: Compressed from %d bytes to %d bytes\n", len(data)*4, len(compressed)*4)

		recovered := runUncompression(compressed, codec)
		fmt.Printf("variablebyte/TestBasicExample: Uncompressed from %d bytes to %d bytes\n", len(compressed)*4, len(recovered)*4)

		if !reflect.DeepEqual(data, recovered) {
			t.Fatalf("variablebyte/TestBasicExample: Problem recovering. Original length = %d, recovered length = %d\n", len(data), len(recovered))
		}
	}
}

func BenchmarkCompress(b *testing.B) {
	data := generateData(b.N)
	codec := NewIntegratedVariableByte()

	b.ResetTimer()
	compressed := runCompression(data, codec)
	b.StopTimer()

	fmt.Printf("variablebyte/BenchmarkCompress: Compressed from %d bytes to %d bytes\n", len(data)*4, len(compressed)*4)
}

func BenchmarkUncompress(b *testing.B) {
	data := generateData(b.N)
	codec := NewIntegratedVariableByte()
	compressed := runCompression(data, codec)

	b.ResetTimer()
	recovered := runUncompression(compressed, codec)
	b.StopTimer()

	fmt.Printf("variablebyte/BenchmarkUncompress: Uncompressed from %d bytes to %d bytes\n", len(compressed)*4, len(recovered)*4)
}
