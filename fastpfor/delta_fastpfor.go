/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package fastpfor

import (
	"math"
	"errors"
	"github.com/reducedb/encoding"
	"github.com/reducedb/encoding/buffers"
	"github.com/reducedb/encoding/bitpacking"
)

type DeltaFastPFOR struct {
	dataToBePacked [32][]int32
	byteContainer *buffers.ByteBuffer
	pageSize int32

	// Working area
	dataPointers        []int32
	freqs               []int32
}

var _ encoding.Integer = (*DeltaFastPFOR)(nil)

func NewDeltaFastPFOR() encoding.Integer {
	f := &DeltaFastPFOR{
		pageSize: DefaultPageSize,
		byteContainer: buffers.NewByteBuffer(3*DefaultPageSize/DefaultBlockSize + DefaultPageSize),
		dataPointers: make([]int32, 33),
		freqs: make([]int32, 33),
	}

	for i := 0; i < 32; i++ {
		f.dataToBePacked[i] = make([]int32, DefaultPageSize/32*4)
	}

	return f
}

func (this *DeltaFastPFOR) Compress(in []int32, inpos *encoding.Cursor, inlength int, out []int32, outpos *encoding.Cursor) error {
	inlength = encoding.FloorBy(inlength, DefaultBlockSize)

	if inlength == 0 {
		return errors.New("fastpfor/Compress: inlength = 0. No work done.")
	}

	out[outpos.Get()] = int32(inlength)
	outpos.Increment()

	initoffset := encoding.NewCursor()

	copy(this.dataPointers, zeroDataPointers)
	copy(this.freqs, zeroFreqs)

	finalInpos := inpos.Get() + inlength

	for inpos.Get() != finalInpos {
		thissize := int(math.Min(float64(this.pageSize), float64(finalInpos - inpos.Get())))

		if err := this.encodePage(in, inpos, thissize, out, outpos, initoffset); err != nil {
			return errors.New("fastpfor/Compress: " + err.Error())
		}
	}

	return nil
}

func (this *DeltaFastPFOR) Uncompress(in []int32, inpos *encoding.Cursor, inlength int, out []int32, outpos *encoding.Cursor) error {
	if inlength == 0 {
		return errors.New("fastpfor/Uncompress: inlength = 0. No work done.")
	}

	mynvalue := in[inpos.Get()]
	inpos.Increment()

	initoffset := encoding.NewCursor()

	copy(this.dataPointers, zeroDataPointers)

	finalout := outpos.Get() + int(mynvalue)
	for outpos.Get() != finalout {
		thissize := int(math.Min(float64(this.pageSize), float64(finalout - outpos.Get())))

		if err := this.decodePage(in, inpos, out, outpos, thissize, initoffset); err != nil {
			return errors.New("fastpfor/Uncompress: " + err.Error())
		}
	}
	return nil
}

// getBestBFromData determins the best bit position with the best cost of exceptions,
// and the max bit position of the array of int32s
func (this *DeltaFastPFOR) getBestBFromData(in []int32) (bestb int32, bestc int32, maxb int32) {
	copy(this.freqs, zeroFreqs)

	// Get the count of all the leading bit positions for the slice
	// Mainly to figure out what's the best (most popular) bit position
	for _, v := range in {
		l := encoding.LeadingBitPosition(uint32(v))
		this.freqs[l] += 1
		if l > bestb {
			bestb = l
		}
	}

	maxb = bestb
	bestCost := bestb*DefaultBlockSize
	var cexcept int32
	bestc = cexcept

	// Find the cost of storing exceptions for each bit position
	for b := bestb - 1; b >= 0; b-- {
		cexcept += this.freqs[b+1]
		if cexcept < 0 {
			break
		}

		// the extra 8 is the cost of storing maxbits
		thisCost := cexcept*OverheadOfEachExcept + cexcept*(maxb - b) + b*DefaultBlockSize + 8

		if thisCost < bestCost {
			bestCost = thisCost
			bestb = b
			bestc = cexcept
		}
	}

	return
}

func (this *DeltaFastPFOR) encodePage(in []int32, inpos *encoding.Cursor, thissize int, out []int32, outpos *encoding.Cursor, initoffset *encoding.Cursor) error {
	headerpos := int32(outpos.Get())
	outpos.Increment()
	tmpoutpos := int32(outpos.Get())

	// Clear working area
	copy(this.dataPointers, zeroDataPointers)
	this.byteContainer.Clear()

	tmpinpos := int32(inpos.Get())
	var delta [DefaultBlockSize]int32

	for finalInpos := tmpinpos + int32(thissize) - DefaultBlockSize; tmpinpos <= finalInpos; tmpinpos += DefaultBlockSize {
		offset := int32(initoffset.Get())
		for i, v := range in[tmpinpos:tmpinpos+DefaultBlockSize] {
			delta[i] = v - offset
			offset = v
		}
		//encoding.Delta(in[tmpinpos:tmpinpos+DefaultBlockSize], delta, int32(initoffset.Get()))

		/*
		copy(delta, in[tmpinpos:tmpinpos+DefaultBlockSize])
        for i := 0; i < DefaultBlockSize; i += 4 {
			tmpoffset := delta[i+3]
			delta[i+3] -= delta[i+2]
			delta[i+2] -= delta[i+1]
			delta[i+1] -= delta[i]
			delta[i] -= offset
			offset = tmpoffset
        }
        */

		/*
		for i := 0; i < DefaultBlockSize; i += 4 {
			delta[i] = in[int(tmpinpos)+i] - offset
			delta[i+1] = in[int(tmpinpos)+i+1] - in[int(tmpinpos)+i]
			delta[i+2] = in[int(tmpinpos)+i+2] - in[int(tmpinpos)+i+1]
			delta[i+3] = in[int(tmpinpos)+i+3] - in[int(tmpinpos)+i+2]
			offset = in[int(tmpinpos)+i+3]
		}
		*/

		initoffset.Set(int(in[tmpinpos+DefaultBlockSize-1]))

		//bestb, bestc, maxb := this.getBestBFromData(in[tmpinpos:tmpinpos+DefaultBlockSize])
		bestb, bestc, maxb := this.getBestBFromData(delta[:])
		tmpbestb := bestb
		this.byteContainer.Put(byte(bestb))
		this.byteContainer.Put(byte(bestc))

		if bestc > 0 {
			this.byteContainer.Put(byte(maxb))
			index := maxb - bestb

			if int(this.dataPointers[index] + bestc) >= len(this.dataToBePacked[index]) {
				newSize := int(2*(this.dataPointers[index] + bestc))

				// make sure it is a multiple of 32.
				// there might be a better way to do this
				newSize = encoding.CeilBy(newSize, 32)
				newSlice := make([]int32, newSize)
				copy(newSlice, this.dataToBePacked[index])
				this.dataToBePacked[index] = newSlice
			}

			for k := int32(0); k < DefaultBlockSize; k++ {
				if uint32(delta[k]) >> uint(bestb) != 0 {
					// we have an exception
					this.byteContainer.Put(byte(k))
					this.dataToBePacked[index][this.dataPointers[index]] = int32(uint32(delta[k]) >> uint(tmpbestb))
					this.dataPointers[index] += 1
				}
			}
		}

		for k := int32(0); k < 128; k += 32 {
			bitpacking.FastPack(delta[:], int(k), out, int(tmpoutpos), int(tmpbestb))
			tmpoutpos += tmpbestb
		}
	}

	inpos.Set(int(tmpinpos))
	out[headerpos] = tmpoutpos - headerpos

	for this.byteContainer.Position() & 3 != 0 {
		this.byteContainer.Put(0)
	}

	bytesize := int32(this.byteContainer.Position())
	out[tmpoutpos] = bytesize
	tmpoutpos += 1
	howmanyints := bytesize / 4

	this.byteContainer.Flip()
	this.byteContainer.AsInt32Buffer().GetInt32s(out, int(tmpoutpos), int(howmanyints))
	tmpoutpos += howmanyints

	bitmap := int32(0)
	for k := 1; k <=32; k++ {
		v := this.dataPointers[k]
		if v != 0 {
			bitmap |= (1 << uint(k - 1))
		}
	}

	out[tmpoutpos] = bitmap
	tmpoutpos += 1

	for k := 1; k <= 31; k++ {
		v := this.dataPointers[k]
		if v != 0 {
			out[tmpoutpos] = v // size
			tmpoutpos += 1
			for j := 0; j < int(v); j += 32 {
				bitpacking.FastPack(this.dataToBePacked[k], j, out, int(tmpoutpos), k)
				tmpoutpos += int32(k)
			}
		}
	}

	outpos.Set(int(tmpoutpos))

	return nil
}

func (this *DeltaFastPFOR) decodePage(in []int32, inpos *encoding.Cursor, out []int32, outpos *encoding.Cursor, thissize int, initoffset *encoding.Cursor) error {
	initpos := int32(inpos.Get())
	wheremeta := in[initpos]
	inpos.Increment()

	inexcept := initpos + wheremeta
	bytesize := in[inexcept]
	inexcept += 1

	this.byteContainer.Clear()
	if err := this.byteContainer.AsInt32Buffer().PutInt32s(in, int(inexcept), int(bytesize/4)); err != nil {
		return err
	}

	inexcept += bytesize / 4
	bitmap := in[inexcept]
	inexcept += 1

	for k := int32(1); k <= 31; k++ {
		if bitmap & (1 << uint32(k - 1)) != 0 {
			size := in[inexcept]
			inexcept += 1

			if int32(len(this.dataToBePacked[k])) < size {
				this.dataToBePacked[k] = make([]int32, encoding.CeilBy(int(size), 32))
			}

			for j := int32(0); j < size; j += 32 {
				bitpacking.FastUnpack(in, int(inexcept), this.dataToBePacked[k], int(j), int(k))
				inexcept += k
			}
		}
	}

	copy(this.dataPointers, zeroDataPointers)
	tmpoutpos := int32(outpos.Get())
	tmpinpos := int32(inpos.Get())

	delta := make([]int32, DefaultBlockSize)

	run := 0
	run_end := thissize / DefaultBlockSize
	for run < run_end {
		var err error
		var bestb int32
		if bestb, err = this.byteContainer.GetAsInt32(); err != nil {
			return err
		}

		var cexcept int32
		if cexcept, err = this.byteContainer.GetAsInt32(); err != nil {
			return err
		}

		for k := int32(0); k < 128; k += 32 {
			//bitpacking.FastUnpack(in, int(tmpinpos), out, int(tmpoutpos+k), int(bestb))
			bitpacking.FastUnpack(in, int(tmpinpos), delta, int(k), int(bestb))
			tmpinpos += bestb
		}

		if cexcept > 0 {
			var maxbits int32
			if maxbits, err = this.byteContainer.GetAsInt32(); err != nil {
				return err
			}

			index := maxbits - bestb

			for k := int32(0); k < cexcept; k++ {
				var pos int32
				if pos, err = this.byteContainer.GetAsInt32(); err != nil {
					return err
				}

				exceptvalue := this.dataToBePacked[index][this.dataPointers[index]]
				this.dataPointers[index] += 1
				//out[pos + tmpoutpos] |= exceptvalue << uint(bestb)
				delta[pos] |= exceptvalue << uint(bestb)
			}
		}

		//encoding.InverseDelta(delta, out[tmpoutpos:tmpoutpos+DefaultBlockSize], int32(initoffset.Get()))
		offset := int32(initoffset.Get())
		for i, v := range delta {
			out[int(tmpoutpos)+i], offset = v + offset, v + offset
			//offset += v
		}
		initoffset.Set(int(out[tmpoutpos+DefaultBlockSize-1]))

		run += 1
		tmpoutpos += DefaultBlockSize
	}

	outpos.Set(int(tmpoutpos))
	inpos.Set(int(inexcept))

	return nil
}
