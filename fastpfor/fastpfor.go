/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package fastpfor

import (
	"math"
	"github.com/reducedb/encoding"
	"github.com/reducedb/encoding/buffers"
)

const (
	DefaultBlockSize     = 128
	OverheadOfEachExcept = 8
	DefaultPageSize      = 65536
)

var (
	zeroDataPointers []uint64
	zeroFreqs []uint64
	zeroBest []byte
)

func init() {
	zeroDataPointers = make([]uint64, 33)
	zeroFreqs = make([]uint64, 33)
	zeroBest = make([]byte, 3)
}

type FastPFOR struct {
	dataToBePacked [32][]uint64
	byteContainer *buffers.ByteBuffer
	pageSize uint64

	// Working area
	dataPointers        []uint64
	freqs               []uint64
	bestbestcexceptmaxb []byte
}

var _ encoding.Integer = (*FastPFOR)(nil)

func New() encoding.Integer {
	f := &FastPFOR{
		pageSize: DefaultPageSize,
		byteContainer: NewByteBuffer(3*DefaultPageSize/DefaultBlockSize + DefaultPageSize),
		dataPointers: make([]uint64, 33),
		freqs: make([]uint64, 33),
		bestbestcexceptmaxb: make([]byte, 3),
	}

	for i := 0; i < 32; i++ {
		f.dataToBePacked[i] = make([]uint64, DefaultPageSize/32*4)
	}

	return f
}

func (this *FastPFOR) Compress(in []uint64, inpos encoding.Cursor, inlength int, out []uint64, outpos encoding.Cursor) {
	inlength = encoding.FloorBy(inlength, this.pageSize)
	if inlength == 0 {
		return
	}

	out[outpos.get()] = inlength
	outpos.increment()

	copy(this.dataPointers, zeroDataPointers)
	copy(this.freqs, zeroFreqs)
	copy(this.bestbestcexceptmaxb, zeroBest)

	finalInpos := inpos.get() + inlength

	for inpos.get() != finalInpos {
		thisSize := int(math.Min(float64(this.pageSize), float64(finalInpos - inpos.get())))
		this.encodePage(in, inpos, thisSize, out, outpos)
	}
}

func (this *FastPFOR) Uncompress(in []uint64, inpos encoding.Cursor, inlength int, out []uint64, outpos encoding.Cursor) {

}

func (this *FastPFOR) getBestBFromData(in []uint64, pos int) {
	copy(this.freqs, zeroFreqs)

	k := pos
	kEnd := pos + DefaultBlockSize

	for ; k < kEnd; k++ {
		this.freqs[encoding.NumberOfBitsUint32(uint32(in[k]))] += 1
	}

	this.bestbestcexceptmaxb[0] = 32

	for this.freqs[this.bestbestcexceptmaxb[0]] == 0 {
		this.bestbestcexceptmaxb[0] -= 1
	}

	this.bestbestcexceptmaxb[2] = this.bestbestcexceptmaxb[0]

	bestCost := uint64(this.bestbestcexceptmaxb[0]*DefaultBlockSize)

	var cexcept byte = 0
	this.bestbestcexceptmaxb[1] = cexcept

	for b := this.bestbestcexceptmaxb[0] - 1; b >= 0; b-- {
		cexcept += byte(this.freqs[b + 1])
		if cexcept < 0 {
			break
		}

		thisCost := cexcept*OverheadOfEachExcept + cexcept*(this.bestbestcexceptmaxb[2] - b) + b*DefaultBlockSize + 8

		if thisCost < bestCost {
			bestCost = thisCost
			this.bestbestcexceptmaxb[0] = byte(b)
			this.bestbestcexceptmaxb[1] = cexcept
		}
	}
}

func (this *FastPFOR) encodePage(in []uint64, inpos encoding.Cursor, thisSize int, out []uint64, outpos encoding.Cursor) {
	headerPos := outpos.get()
	outpos.increment()
	tmpOutpos := outpos.get()

	// Clear working area
	copy(this.dataPointers, zeroDataPointers)
	//copy(this.bestbestcexceptmaxb, zeroBest)		// may not be necessary
	this.byteContainer.Clear()

	tmpInpos := inpos.get()

	for finalInpos := tmpInpos + thisSize - DefaultBlockSize; tmpInpos <= finalInpos; tmpInpos += DefaultBlockSize {
		this.getBestBFromData(in, tmpInpos)
		tmpBestB := this.bestbestcexceptmaxb[0]
		this.byteContainer.Put(this.bestbestcexceptmaxb[0])
		this.byteContainer.Put(this.bestbestcexceptmaxb[1])

		if this.bestbestcexceptmaxb[1] > 0 {
			this.byteContainer.put(this.bestbestcexceptmaxb[2])
			index := uint32(this.bestbestcexceptmaxb[2] - this.bestbestcexceptmaxb[0])

			if this.dataPointers[index] + this.bestbestcexceptmaxb[1] >= len(this.dataToBePacked[index]) {
				newSize := 2*(this.dataPointers[index] + this.bestbestcexceptmaxb[1])

				newSize = encoding.FloorBy(newSize + 31, 32)
				this.dataToBePacked[index] =
			}
		}
	}

}
