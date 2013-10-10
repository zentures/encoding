/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package encoding

import (
	"testing"
	"io"
	"fmt"
	"compress/lzw"
	"compress/gzip"
	"encoding/binary"
	"bytes"
)

func generateDataInBytes(N int) *bytes.Buffer {
	data := GenerateClustered(N, N*2)
	b := make([]byte, N*4)
	for i := 0; i < N; i++ {
		binary.LittleEndian.PutUint32(b[i*4:], uint32(data[i]))
	}

	fmt.Printf("encoding/generateData: len(data) = %d bytes\n", len(b))

	return bytes.NewBuffer(b)
}

func runLZWCompress(data *bytes.Buffer) *bytes.Buffer {
	var compressed bytes.Buffer
	w := lzw.NewWriter(&compressed, lzw.MSB, 8)
	defer w.Close()
	w.Write(data.Bytes())
	return &compressed
}

func runLZWDecompress(data *bytes.Buffer, length int) *bytes.Buffer {
	recovered := make([]byte, length*4)
	r := lzw.NewReader(data, lzw.MSB, 8)
	defer r.Close()

	total := 0
	n := 0
	var err error = nil
	for err != io.EOF {
		n, err = r.Read(recovered[total:])
		total += n
	}

	if total != len(recovered) {
		fmt.Printf("encoding/runLZWDecompress: something is wrong. Read %d bytes, expecting %d bytes\n", total, len(recovered))
	}

	return bytes.NewBuffer(recovered)
}

func TestLZW(t *testing.T) {
	for _, k := range []int{1, 13, 133, 1333, 133333, 13333333} {
		data := generateDataInBytes(k)
        RunTestLZW(data.Bytes(), t)
	}
}

func BenchmarkLZWCompress(b *testing.B) {
	data := generateDataInBytes(b.N)

	b.ResetTimer()
	compressed := runLZWCompress(data)
	b.StopTimer()

	fmt.Printf("encoding/BenchmarkLZWCompress: Compressed from %d bytes to %d bytes\n", data.Len(), compressed.Len())
}

func BenchmarkLZWDecompress(b *testing.B) {
	data := generateDataInBytes(b.N)
	compressed := runLZWCompress(data)
	cl := compressed.Len()

	b.ResetTimer()
	recovered := runLZWDecompress(compressed, b.N)
	b.StopTimer()

	fmt.Printf("encoding/BenchmarkLZWDecompress: Decompressed from %d bytes to %d bytes\n", cl, recovered.Len())
}

func runGzipCompress(data *bytes.Buffer) *bytes.Buffer {
	var compressed bytes.Buffer
	w := gzip.NewWriter(&compressed)
	defer w.Close()
	w.Write(data.Bytes())
	return &compressed
}

func runGzipDecompress(data *bytes.Buffer, length int) *bytes.Buffer {
	recovered := make([]byte, length*4)
	r, _ := gzip.NewReader(data)
	defer r.Close()

	total := 0
	n := 100
	var err error = nil
	for err != io.EOF && n != 0 {
		n, err = r.Read(recovered[total:])
		total += n
	}

	if total != len(recovered) {
		fmt.Printf("encoding/runGzipDecompress: something is wrong. Read %d bytes, expecting %d bytes\n", total, len(recovered))
	}

	return bytes.NewBuffer(recovered)
}

func TestGzip(t *testing.T) {
	for _, k := range []int{1, 13, 133, 1333, 133333, 13333333} {
		data := generateDataInBytes(k)
        RunTestGzip(data.Bytes(), t)
	}
}

func BenchmarkGzipCompress(b *testing.B) {
	data := generateDataInBytes(b.N)

	b.ResetTimer()
	compressed := runGzipCompress(data)
	b.StopTimer()

	fmt.Printf("encoding/BenchmarkGzipCompress: Compressed from %d bytes to %d bytes\n", data.Len(), compressed.Len())
}

func BenchmarkGzipDecompress(b *testing.B) {
	data := generateDataInBytes(b.N)
	compressed := runGzipCompress(data)
	cl := compressed.Len()

	b.ResetTimer()
	recovered := runGzipDecompress(compressed, b.N)
	b.StopTimer()

	fmt.Printf("encoding/BenchmarkGzipDecompress: Decompressed from %d bytes to %d bytes\n", cl, recovered.Len())
}
