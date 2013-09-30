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
	"reflect"
	"compress/lzw"
	"compress/gzip"
	"encoding/binary"
	"bytes"
)

func generateDataInBytes(N int) *bytes.Buffer {
	data := GenerateClustered(N, N*2)
	b := make([]byte, N*4)
	for i := 0; i < N; i++ {
		binary.LittleEndian.PutUint32(b[i*4:], data[i])
	}

	fmt.Printf("encoding/generateData: len(data) = %d bytes\n", len(b))

	return bytes.NewBuffer(b)
}

func runLZWCompress(data *bytes.Buffer) *bytes.Buffer {
	var compressed bytes.Buffer
	w := lzw.NewWriter(&compressed, lzw.MSB, 8)
	w.Write(data.Bytes())
	w.Close()
	return &compressed
}

func runLZWDecompress(data *bytes.Buffer, length int) *bytes.Buffer {
	recovered := make([]byte, length*4)
	r := lzw.NewReader(data, lzw.MSB, 8)

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

		compressed := runLZWCompress(data)
		cl := compressed.Len()
		fmt.Printf("encoding/TestBasicExample: Compressed from %d bytes to %d bytes\n", len(data.Bytes()), cl)

		recovered := runLZWDecompress(compressed, k)
		fmt.Printf("encoding/TestBasicExample: Decompressed from %d bytes to %d bytes\n", cl, len(recovered.Bytes()))

		if !reflect.DeepEqual(data, recovered) {
			t.Fatalf("encoding/TestBasicExample: Problem recovering. Original length = %d, recovered length = %d\n", len(data.Bytes()), len(recovered.Bytes()))
		}
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
	w.Write(data.Bytes())
	w.Close()
	return &compressed
}

func runGzipDecompress(data *bytes.Buffer, length int) *bytes.Buffer {
	recovered := make([]byte, length*4)
	r, _ := gzip.NewReader(data)

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

		compressed := runGzipCompress(data)
		cl := compressed.Len()
		fmt.Printf("encoding/TestGzip: Compressed from %d bytes to %d bytes\n", len(data.Bytes()), cl)

		recovered := runGzipDecompress(compressed, k)
		fmt.Printf("encoding/TestGzip: Decompressed from %d bytes to %d bytes\n", cl, len(recovered.Bytes()))

		if !reflect.DeepEqual(data, recovered) {
			t.Fatalf("encoding/TestGzip: Problem recovering. Original length = %d, recovered length = %d\n", len(data.Bytes()), len(recovered.Bytes()))
		}
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

/*
func testBasic(t *testing.T) {
	N := 5
	nbr := 10

	for sparsity := 1; sparsity < 31 - nbr; sparsity += 4 {
		fmt.Println("Testing sparsity", sparsity)
		data := make([][]uint32, N)
		max := 1<<uint(nbr + sparsity)

		for k := 0; k < N; k++ {
			data[k] = encoding.GenerateClustered(1<<uint(nbr), max)
		}

		codec := NewIntegratedBP32()
		testCodec(t, codec, data, max)

	}
}

func testCodec(t *testing.T, c encoding.Integer, data [][]uint32, max int) {
	N := len(data)
	maxlength := 0

	for k := 0; k < N; k++ {
		if len(data[k]) > maxlength {
			maxlength = len(data[k])
		}
	}

	buffer := make([]uint32, maxlength + 1024)
	dataout := make([]uint32, 4*maxlength + 1024)

	inpos := encoding.NewCursor()
	outpos := encoding.NewCursor()

	for k := 0; k < N; k++ {
		backupdata := append(make([]uint32, 0), data[k]...)
		fmt.Printf("bp32_test/testCodec: len(backupdata) = %d\n", len(backupdata))
		inpos.Set(1)
		outpos.Set(0)

		c.Compress(backupdata, inpos, len(backupdata) - inpos.Get(), dataout, outpos)

		fmt.Printf("bp32_test/testCodec: inpos = %d, outpos = %d\n", inpos.Get(), outpos.Get())
		thiscompsize := outpos.Get() + 1
		inpos.Set(0)
		outpos.Set(1)
		buffer[0] = backupdata[0]

		c.Decompress(dataout, inpos, thiscompsize - 1, buffer, outpos)

		if outpos.Get() != len(data[k]) {
			t.Fatalf("We have a bug (diff length): %d expected, got %d\n", len(data[k]), outpos.Get())
		}

		for m := 0; m < outpos.Get(); m++ {
			if buffer[m] != data[k][m] {
				t.Fatalf("We have a bug (actual difference), %d expected, found %d at $d\n", data[k][m], buffer[m], m)
			}
		}
	}
}
*/
