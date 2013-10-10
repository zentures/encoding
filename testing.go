/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package encoding

import (
	"testing"
	"os"
	"io"
	"bytes"
	"log"
	"time"
	"compress/gzip"
	"compress/lzw"
	"runtime/pprof"
)

func TestCodec(codec Integer, data []int32, sizes []int, t *testing.T) {
	for _, k := range sizes {
		if k > len(data) {
			continue
		}

		log.Printf("encoding/TestCodec: Testing with %d integers\n", k)

		now := time.Now()
		compressed := Compress(codec, data, k)
		log.Printf("encoding/TestCodec: Compressed %d integers from %d bytes to %d bytes in %d ns\n", k, k*4, len(compressed)*4, time.Since(now).Nanoseconds())

		recovered := Uncompress(codec, compressed, k)
		log.Printf("encoding/TestCodec: Uncompressed from %d bytes to %d bytes in %d ns\n", len(compressed)*4, len(recovered)*4, time.Since(now).Nanoseconds())

		for i := 0; i < k; i++ {
			if data[i] != recovered[i] {
				t.Fatalf("encoding/TestCodec: Problem recovering. index = %d, data = %d, recovered = %d, original length = %d, recovered length = %d\n", i, data[i], recovered[i], k, len(recovered))
			}
		}
	}
}

func TestCodecPprof(codec Integer, data []int32, sizes []int, t *testing.T) {
	for _, k := range sizes {
		if k > len(data) {
			continue
		}

		f, e := os.Create("cpu.compress.prof")
		if e != nil {
			log.Fatal(e)
		}
		defer f.Close()

		now := time.Now()
		pprof.StartCPUProfile(f)
		compressed := Compress(codec, data, k)
		pprof.StopCPUProfile()

		log.Printf("encoding/TestCodecPprof: Compressed %d integers from %d bytes to %d bytes in %d ns\n", k, k*4, len(compressed)*4, time.Since(now).Nanoseconds())

		f2, e2 := os.Create("cpu.uncompress.prof")
		if e2 != nil {
			log.Fatal(e2)
		}
		defer f2.Close()

		log.Printf("encoding/TestCodecPprof: Testing decompression\n")
		now = time.Now()
		pprof.StartCPUProfile(f2)
		recovered := Uncompress(codec, compressed, k)
		pprof.StopCPUProfile()
		log.Printf("encoding/TestCodecPprof: Uncompressed from %d bytes to %d bytes in %d ns\n", len(compressed)*4, len(recovered)*4, time.Since(now).Nanoseconds())

		for i := 0; i < k; i++ {
			if data[i] != recovered[i] {
				t.Fatalf("encoding/TestCodecPprof: Problem recovering. index = %d, data = %d, recovered = %d, original length = %d, recovered length = %d\n", i, data[i], recovered[i], k, len(recovered))
			}
		}
	}
}

func BenchmarkCompress(codec Integer, data []int32, b *testing.B) {
	k := CeilBy(b.N, 128)

	b.ResetTimer()
	now := time.Now()
	compressed := Compress(codec, data, k)
	b.StopTimer()

	log.Printf("encoding/BenchmarkCompress: Compressed %d integers from %d bytes to %d bytes in %d ns\n", k, k*4, len(compressed)*4, time.Since(now).Nanoseconds())
}

func BenchmarkUncompress(codec Integer, data []int32, b *testing.B) {
	k := CeilBy(b.N, 128)
	compressed := Compress(codec, data, k)

	b.ResetTimer()
	recovered := Uncompress(codec, compressed, k)
	b.StopTimer()

	log.Printf("encoding/BenchmarkUncompress: Uncompressed from %d bytes to %d bytes\n", len(compressed)*4, len(recovered)*4)
}

func Compress(codec Integer, data []int32, length int) []int32 {
	compressed := make([]int32, length*2)
	inpos := NewCursor()
	outpos := NewCursor()
	codec.Compress(data, inpos, length, compressed, outpos)
	compressed = compressed[:outpos.Get()]
	return compressed
}

func Uncompress(codec Integer, data []int32, length int) []int32 {
	recovered := make([]int32, length)
	rinpos := NewCursor()
	routpos := NewCursor()
	codec.Uncompress(data, rinpos, len(data), recovered, routpos)
	recovered = recovered[:routpos.Get()]
	return recovered
}

func TestGzip(data []byte, t *testing.T) {
	log.Printf("encoding/TestGzip: Testing comprssion Gzip\n")

	var compressed bytes.Buffer
	w := gzip.NewWriter(&compressed)
	defer w.Close()
	now := time.Now()
	w.Write(data)

	cl := compressed.Len()
	log.Printf("encoding/TestGzip: Compressed from %d bytes to %d bytes in %d ns\n", len(data), cl, time.Since(now).Nanoseconds())

	recovered := make([]byte, len(data))
	r, _ := gzip.NewReader(&compressed)
	defer r.Close()

	total := 0
	n := 100
	var err error = nil
	for err != io.EOF && n != 0 {
		n, err = r.Read(recovered[total:])
		total += n
	}
	log.Printf("encoding/TestGzip: Uncompressed from %d bytes to %d bytes in %d ns\n", cl, len(recovered), time.Since(now).Nanoseconds())
}

func TestLZW(data []byte, t *testing.T) {
	log.Printf("encoding/TestLZW: Testing comprssion LZW\n")

	var compressed bytes.Buffer
	w := lzw.NewWriter(&compressed, lzw.MSB, 8)
	defer w.Close()
	now := time.Now()
	w.Write(data)

	cl := compressed.Len()
	log.Printf("encoding/TestLZW: Compressed from %d bytes to %d bytes in %d ns\n", len(data), cl, time.Since(now).Nanoseconds())

	recovered := make([]byte, len(data))
	r := lzw.NewReader(&compressed, lzw.MSB, 8)
	defer r.Close()

	total := 0
	n := 100
	var err error = nil
	for err != io.EOF && n != 0 {
		n, err = r.Read(recovered[total:])
		total += n
	}
	log.Printf("encoding/TestLZW: Uncompressed from %d bytes to %d bytes in %d ns\n", cl, len(recovered), time.Since(now).Nanoseconds())
}
