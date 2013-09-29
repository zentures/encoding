/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package encoding

import (
	"testing"
	"fmt"
	"github.com/reducedb/encoding"
)

func TestBasic(t *testing.T) {
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

		c.Uncompress(dataout, inpos, thiscompsize - 1, buffer, outpos)

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
