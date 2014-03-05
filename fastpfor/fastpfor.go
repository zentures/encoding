/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package fastpfor

import (
	"errors"
	"github.com/reducedb/bytebuffer"
	"github.com/reducedb/encoding"
	"github.com/reducedb/encoding/bitpacking"
	"github.com/reducedb/encoding/cursor"
	"math"
)

const (
	DefaultBlockSize     = 128
	OverheadOfEachExcept = 8
	DefaultPageSize      = 65536
)

var (
	zeroDataPointers []int32
	zeroFreqs        []int32
)

func init() {
	zeroDataPointers = make([]int32, 33)
	zeroFreqs = make([]int32, 33)
}

type FastPFOR struct {
	dataToBePacked [33][]int32
	byteContainer  *bytebuffer.ByteBuffer
	pageSize       int32

	// Working area
	dataPointers []int32
	freqs        []int32
}

var _ encoding.Integer = (*FastPFOR)(nil)

func New() encoding.Integer {
	f := &FastPFOR{
		pageSize:      DefaultPageSize,
		byteContainer: bytebuffer.NewByteBuffer(3*DefaultPageSize/DefaultBlockSize + DefaultPageSize),
		dataPointers:  make([]int32, 33),
		freqs:         make([]int32, 33),
	}

	for i := 1; i < 33; i++ {
		f.dataToBePacked[i] = make([]int32, DefaultPageSize/32*4)
	}

	return f
}

func (this *FastPFOR) Compress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {
	inlength = encoding.FloorBy(inlength, DefaultBlockSize)

	//log.Printf("fastpfor/Compress: inlength = %d, pageSize = %d\n", inlength, this.pageSize)
	if inlength == 0 {
		return errors.New("fastpfor/Compress: inlength = 0. No work done.")
	}

	//originpos := inpos.Get()
	//origoutpos := outpos.Get()
	out[outpos.Get()] = int32(inlength)
	outpos.Increment()

	copy(this.dataPointers, zeroDataPointers)
	copy(this.freqs, zeroFreqs)

	finalInpos := inpos.Get() + inlength
	//log.Printf("fastpfor: finalInpos = %d\n", finalInpos)

	for inpos.Get() != finalInpos {
		thissize := int(math.Min(float64(this.pageSize), float64(finalInpos-inpos.Get())))
		//log.Printf("fastpfor/Compress: thissize = %d\n", thissize)
		if err := this.encodePage(in, inpos, thissize, out, outpos); err != nil {
			//log.Printf("fastpfor/Compress: error with encodePage - %v\n", err)
			return errors.New("fastpfor/Compress: " + err.Error())
		}
	}

	//log.Printf("fastpfor/Compress: inpos[%d:%d] = %v\n", originpos, finalInpos, in[originpos:finalInpos])
	//encoding.PrintInt32sInBits(out[origoutpos:outpos.Get()][0:outpos.Get()-origoutpos])
	return nil
}

func (this *FastPFOR) Uncompress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {
	if inlength == 0 {
		return errors.New("fastpfor/Uncompress: inlength = 0. No work done.")
	}

	mynvalue := in[inpos.Get()]
	inpos.Increment()

	copy(this.dataPointers, zeroDataPointers)

	finalout := outpos.Get() + int(mynvalue)
	for outpos.Get() != finalout {
		thissize := int(math.Min(float64(this.pageSize), float64(finalout-outpos.Get())))
		if err := this.decodePage(in, inpos, out, outpos, thissize); err != nil {
			//log.Printf("fastpfor/Uncompress: error with decodePage - %v\n", err)
			return errors.New("fastpfor/Uncompress: " + err.Error())
		}
		//log.Printf("fastpfor/Uncompress: thissize = %d, inpos = %d, outpos = %d, finalout = %d\n", thissize, inpos.Get(), outpos.Get(), finalout)
	}
	return nil
}

// getBestBFromData determins the best bit position with the best cost of exceptions,
// and the max bit position of the array of int32s
func (this *FastPFOR) getBestBFromData(in []int32) (bestb int32, bestc int32, maxb int32) {
	copy(this.freqs, zeroFreqs)

	//k := pos
	//kEnd := pos + DefaultBlockSize

	// Get the count of all the leading bit positionsfor the slice
	// Mainly to figure out what's the best (most popular) bit position
	//for _, v := range in[k:kEnd] {
	for _, v := range in {
		this.freqs[encoding.LeadingBitPosition(uint32(v))]++
	}
	//encoding.FastLeadingBitFrequency128(in, this.freqs)

	bestb = 32

	for this.freqs[bestb] == 0 {
		bestb--
	}

	maxb = bestb
	bestCost := bestb * DefaultBlockSize
	var cexcept int32
	bestc = cexcept

	// Find the cost of storing exceptions for each bit position
	for b := bestb - 1; b >= 0; b-- {
		cexcept += this.freqs[b+1]
		//log.Printf("fastpfor/getBestBFromData: this.freqs[%d] = %d, b = %d, cexcept = %d\n", (b+1), this.freqs[(b+1)], b, cexcept)
		if cexcept < 0 {
			break
		}

		// the extra 8 is the cost of storing maxbits
		thisCost := cexcept*OverheadOfEachExcept + cexcept*(maxb-b) + b*DefaultBlockSize + 8
		//log.Printf("fastpfor/bestBFromData: thisCost = %d, bestCost = %d\n", thisCost, bestCost)

		if thisCost < bestCost {
			bestCost = thisCost
			bestb = b
			bestc = cexcept
		}
	}

	return
}

func (this *FastPFOR) encodePage(in []int32, inpos *cursor.Cursor, thissize int, out []int32, outpos *cursor.Cursor) error {
	//log.Printf("fastpfor/encodePage: encoding %d integers\n", thissize)
	headerpos := int32(outpos.Get())
	outpos.Increment()
	tmpoutpos := int32(outpos.Get())

	// Clear working area
	copy(this.dataPointers, zeroDataPointers)
	this.byteContainer.Clear()

	tmpinpos := int32(inpos.Get())
	//log.Printf("fastpfor/encodePage: tmpinpos = %d\n", tmpinpos)

	for finalInpos := tmpinpos + int32(thissize) - DefaultBlockSize; tmpinpos <= finalInpos; tmpinpos += DefaultBlockSize {
		//log.Printf("fastpfor/encodePage: finalinpos = %d, tmpinpos = %d\n", finalInpos, tmpinpos)
		bestb, bestc, maxb := this.getBestBFromData(in[tmpinpos : tmpinpos+DefaultBlockSize])
		//log.Printf("fastpfor/encodePage: bestb = %d, bestc = %d, maxb = %d\n", bestb, bestc, maxb)
		tmpbestb := bestb
		this.byteContainer.Put(byte(bestb))
		this.byteContainer.Put(byte(bestc))

		if bestc > 0 {
			this.byteContainer.Put(byte(maxb))
			index := maxb - bestb
			//log.Printf("maxb = %d, bestb = %d, bestc = %d, index = %d\n", maxb, bestb, bestc, index)

			if int(this.dataPointers[index]+bestc) >= len(this.dataToBePacked[index]) {
				newSize := int(2 * (this.dataPointers[index] + bestc))

				// make sure it is a multiple of 32.
				// there might be a better way to do this
				newSize = encoding.CeilBy(newSize, 32)
				newSlice := make([]int32, newSize)
				copy(newSlice, this.dataToBePacked[index])
				this.dataToBePacked[index] = newSlice
				//this.dataToBePacked[index] = append(make([]int32, 0, newSize), this.dataToBePacked[index]...)
			}

			for k := int32(0); k < DefaultBlockSize; k++ {
				if uint32(in[k+tmpinpos])>>uint(bestb) != 0 {
					// we have an exception
					this.byteContainer.Put(byte(k))
					this.dataToBePacked[index][this.dataPointers[index]] = int32(uint32(in[k+tmpinpos]) >> uint(tmpbestb))
					//log.Printf("fastpfor/encodePage: k = %d, index = %d, v = %d, %032b\n", k, index, this.dataToBePacked[index][this.dataPointers[index]], this.dataToBePacked[index][this.dataPointers[index]])
					this.dataPointers[index] += 1
				}
			}
		}

		for k := int32(0); k < 128; k += 32 {
			bitpacking.FastPack(in, int(tmpinpos+k), out, int(tmpoutpos), int(tmpbestb))
			tmpoutpos += tmpbestb
		}
	}

	inpos.Set(int(tmpinpos))
	out[headerpos] = tmpoutpos - headerpos
	//log.Printf("fastpfor/encodePage: headerpos = %d, tmpoutpos = %d, out[headerpos] = %d\n", headerpos, tmpoutpos, out[headerpos])

	for this.byteContainer.Position()&3 != 0 {
		//log.Printf("fastpfor/encodePage: byteContainer.Position() = %d\n", this.byteContainer.Position())
		this.byteContainer.Put(0)
	}

	bytesize := int32(this.byteContainer.Position())
	out[tmpoutpos] = bytesize
	tmpoutpos += 1
	howmanyints := bytesize / 4
	//log.Printf("fastpfor/encodePage: bytesize = %d, howmanyints = %d\n", bytesize, howmanyints)

	this.byteContainer.Flip()
	this.byteContainer.AsInt32Buffer().GetInt32s(out, int(tmpoutpos), int(howmanyints))
	//log.Printf("fastpfor/encodePage: byteContainer[%d:%d] = %v\n", tmpoutpos, tmpoutpos+howmanyints, out[tmpoutpos:tmpoutpos+howmanyints])
	tmpoutpos += howmanyints

	bitmap := int32(0)
	for k := 1; k <= 32; k++ {
		v := this.dataPointers[k]
		if v != 0 {
			bitmap |= (1 << uint(k-1))
		}
	}

	out[tmpoutpos] = bitmap
	tmpoutpos += 1
	//log.Printf("fastpfor/encodePage: bitmap = %d, tmpoutpos = %d\n", bitmap, tmpoutpos)

	for k := 1; k < 33; k++ {
		v := this.dataPointers[k]
		if v != 0 {
			out[tmpoutpos] = v // size
			//log.Printf("fastpfor/encodePage: tmpoutpos = %d, size = %d, k = %d, out = %032b\n", tmpoutpos, v, k, out[tmpoutpos])
			tmpoutpos += 1
			for j := 0; j < int(v); j += 32 {
				bitpacking.FastPack(this.dataToBePacked[k], j, out, int(tmpoutpos), k)
				tmpoutpos += int32(k)
				//log.Printf("fastpfor/encodePage: tmpoutpos = %d\n", tmpoutpos)
			}
		}
	}

	outpos.Set(int(tmpoutpos))
	//encoding.PrintInt32sInBits(out[:tmpoutpos])

	return nil
}

func grapByte(in []int32, index uint) byte {
	return byte(in[index/4] >> (24 - (index%4)*8))
}

func (this *FastPFOR) decodePage(in []int32, inpos *cursor.Cursor, out []int32, outpos *cursor.Cursor, thissize int) error {
	initpos := int32(inpos.Get())
	wheremeta := in[initpos]
	inpos.Increment()

	inexcept := initpos + wheremeta
	bytesize := in[inexcept]
	inexcept += 1
	mybytearray := in[inexcept:]
	mybp := uint(0)

	inexcept += bytesize / 4
	bitmap := in[inexcept]
	inexcept += 1

	for k := int32(1); k < 33; k++ {
		if bitmap&(1<<uint32(k-1)) != 0 {
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

	run := 0
	run_end := thissize / DefaultBlockSize
	for run < run_end {
		bestb := int32(grapByte(mybytearray, mybp))
		mybp++
		cexcept := int32(grapByte(mybytearray, mybp))
		mybp++
                for k := int32(0); k < 128; k += 32 {
			bitpacking.FastUnpack(in, int(tmpinpos), out, int(tmpoutpos+k), int(bestb))
			tmpinpos += bestb
		}

		if cexcept > 0 {
			maxbits := int32(grapByte(mybytearray, mybp))
			mybp++
			index := maxbits - bestb

			for k := int32(0); k < cexcept; k++ {
				pos := int32(grapByte(mybytearray, mybp))
				mybp++
				exceptvalue := this.dataToBePacked[index][this.dataPointers[index]]
				this.dataPointers[index] += 1
				out[pos+tmpoutpos] |= exceptvalue << uint(bestb)
			}
		}

		run += 1
		tmpoutpos += DefaultBlockSize
	}

	outpos.Set(int(tmpoutpos))
	inpos.Set(int(inexcept))

	return nil
}
