/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package encoding

import (
	"testing"
	"fmt"
)

func TestCodec(codec Integer, data []uint32, sizes []int, t *testing.T) {
	for _, k := range sizes {
		if k > len(data) {
			continue
		}

		fmt.Printf("encoding/TestCodec: Testing with %d integers\n", k)

		compressed := Compress(codec, data, k)
		fmt.Printf("encoding/TestCodec: Compressed from %d bytes to %d bytes\n", k*4, len(compressed)*4)

		recovered := Uncompress(codec, compressed, k)
		fmt.Printf("encoding/TestCodec: Uncompressed from %d bytes to %d bytes\n", len(compressed)*4, len(recovered)*4)

		for i := 0; i < k; i++ {
			if data[i] != recovered[i] {
				t.Fatalf("encoding/TestCodec: Problem recovering. Original length = %d, recovered length = %d\n", k, len(recovered))
			}
		}
	}
}

func BenchmarkCompress(codec Integer, data []uint32, b *testing.B) {
	k := int(CeilBy(uint32(b.N), 128))

	b.ResetTimer()
	compressed := Compress(codec, data, k)
	b.StopTimer()

	fmt.Printf("encoding/BenchmarkCompress: Compressed from %d bytes to %d bytes\n", k*4, len(compressed)*4)
}

func BenchmarkUncompress(codec Integer, data []uint32, b *testing.B) {
	k := int(CeilBy(uint32(b.N), 128))
	compressed := Compress(codec, data, k)

	b.ResetTimer()
	recovered := Uncompress(codec, compressed, k)
	b.StopTimer()

	fmt.Printf("encoding/BenchmarkUncompress: Uncompressed from %d bytes to %d bytes\n", len(compressed)*4, len(recovered)*4)
}

func Compress(codec Integer, data []uint32, length int) []uint32 {
	compressed := make([]uint32, length*2)
	inpos := NewCursor()
	outpos := NewCursor()
	codec.Compress(data, inpos, length, compressed, outpos)
	compressed = compressed[:outpos.Get()]
	return compressed
}

func Uncompress(codec Integer, data []uint32, length int) []uint32 {
	recovered := make([]uint32, length)
	rinpos := NewCursor()
	routpos := NewCursor()
	codec.Uncompress(data, rinpos, len(data), recovered, routpos)
	recovered = recovered[:routpos.Get()]
	return recovered
}

