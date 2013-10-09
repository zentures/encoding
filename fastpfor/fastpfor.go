/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package fastpfor

import (
	"math"
	"errors"
	"log"
	"github.com/reducedb/encoding"
	"github.com/reducedb/encoding/buffers"
	"github.com/reducedb/encoding/bitpacking"
)

const (
	DefaultBlockSize     = 128
	OverheadOfEachExcept = 8
	DefaultPageSize      = 65536
)

var (
	zeroDataPointers []int32
	zeroFreqs []int32
)

func init() {
	zeroDataPointers = make([]int32, 33)
	zeroFreqs = make([]int32, 33)
}

type FastPFOR struct {
	dataToBePacked [32][]int32
	byteContainer *buffers.ByteBuffer
	pageSize uint32

	// Working area
	dataPointers        []int32
	freqs               []int32
}

var _ encoding.Integer = (*FastPFOR)(nil)

func New() encoding.Integer {
	f := &FastPFOR{
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

func (this *FastPFOR) Compress(in []int32, inpos *encoding.Cursor, inlength int, out []int32, outpos *encoding.Cursor) error {
	inlength = int(encoding.FloorBy(uint32(inlength), DefaultBlockSize))

	log.Printf("fastpfor/Compress: inlength = %d, pageSize = %d\n", inlength, this.pageSize)
	if inlength == 0 {
		return errors.New("fastpfor/Compress: inlength = 0. No work done.")
	}

	origInpos := inpos.Get()
	origOutpos := outpos.Get()
	out[outpos.Get()] = uint32(inlength)
	outpos.Increment()

	copy(this.dataPointers, zeroDataPointers)
	copy(this.freqs, zeroFreqs)

	finalInpos := inpos.Get() + inlength
	log.Printf("fastpfor: finalInpos = %d\n", finalInpos)

	for inpos.Get() != finalInpos {
		thissize := int(math.Min(float64(this.pageSize), float64(finalInpos - inpos.Get())))
		log.Printf("fastpfor/Compress: thissize = %d\n", thissize)
		if err := this.encodePage(in, inpos, thissize, out, outpos); err != nil {
			log.Printf("fastpfor/Compress: error with encodePage - %v\n", err)
			return errors.New("fastpfor/Compress: " + err.Error())
		}
	}

	log.Printf("fastpfor/Compress: inpos[%d:%d] = %v\n", origInpos, finalInpos, in[origInpos:finalInpos])
	encoding.PrintUint32sInBits(out[origOutpos:outpos.Get()], 0, outpos.Get()-origOutpos)
	return nil
}

func (this *FastPFOR) Uncompress(in []int32, inpos *encoding.Cursor, inlength int, out []int32, outpos *encoding.Cursor) error {
	if inlength == 0 {
		return errors.New("fastpfor/Uncompress: inlength = 0. No work done.")
	}
	
	mynvalue := in[inpos.Get()]
	inpos.Increment()
	
	copy(this.dataPointers, zeroDataPointers)

	finalout := outpos.Get() + int(mynvalue)
	for outpos.Get() != finalout {
		thissize := int(math.Min(float64(this.pageSize), float64(finalout - outpos.Get())))
		if err := this.decodePage(in, inpos, out, outpos, thissize); err != nil {
			log.Printf("fastpfor/Uncompress: error with decodePage - %v\n", err)
			return errors.New("fastpfor/Uncompress: " + err.Error())
		}
		log.Printf("fastpfor/Uncompress: thissize = %d, outpos = %d, finalout = %d\n", thissize, outpos.Get(), finalout)
	}
	return nil
}

// getBestBFromData determins the best bit position with the best cost of exceptions,
// and the max bit position of the array of uint32s
func (this *FastPFOR) getBestBFromData(in []int32, pos int) (uint32, uint32, uint32) {
	copy(this.freqs, zeroFreqs)
	var bestb, bestc, maxb uint32;

	k := pos
	kEnd := pos + DefaultBlockSize

	// Get the count of all the leading bit positionsfor the slice
	// Mainly to figure out what's the best (most popular) bit position
	for _, v := range in[k:kEnd] {
		l := encoding.LeadingBitPosition(v)
		this.freqs[l] += 1
		if l > bestb {
			bestb = l
		}
		log.Printf("fastpfor/bestBFromData: l = %d, bestb = %d\n", l, bestb)
	}

	maxb = bestb
	bestCost := bestb*DefaultBlockSize
	var cexcept uint32
	bestc = cexcept

	// Find the cost of storing exceptions for each bit position
	for b := int(bestb - 1); b >= 0; b-- {
		cexcept += this.freqs[b+1]
		log.Printf("fastpfor/getBestBFromData: this.freqs[%d] = %d, b = %d, cexcept = %d\n", (b+1), this.freqs[(b+1)], b, cexcept)
		if cexcept < 0 {
			break
		}

		// the extra 8 is the cost of storing maxbits
		thisCost := cexcept*OverheadOfEachExcept + cexcept*uint32(int(maxb) - b) + uint32(b)*DefaultBlockSize + 8
		log.Printf("fastpfor/bestBFromData: thisCost = %d, bestCost = %d\n", thisCost, bestCost)

		if thisCost < bestCost {
			bestCost = thisCost
			bestb = uint32(b)
			bestc = cexcept
		}
	}

	return bestb, bestc, maxb
}

func (this *FastPFOR) encodePage(in []int32, inpos *encoding.Cursor, thissize int, out []int32, outpos *encoding.Cursor) error {
	log.Printf("fastpfor/encodePage: encoding %d integers\n", thissize)
	headerpos := uint32(outpos.Get())
	outpos.Increment()
	tmpoutpos := uint32(outpos.Get())

	// Clear working area
	copy(this.dataPointers, zeroDataPointers)
	this.byteContainer.Clear()

	tmpinpos := uint32(inpos.Get())
	log.Printf("fastpfor/encodePage: tmpinpos = %d\n", tmpinpos)

	for finalInpos := tmpinpos + uint32(thissize) - DefaultBlockSize; tmpinpos <= finalInpos; tmpinpos += DefaultBlockSize {
		log.Printf("fastpfor/encodePage: finalinpos = %d, tmpinpos = %d\n", finalInpos, tmpinpos)
		bestb, bestc, maxb := this.getBestBFromData(in, int(tmpinpos))
		log.Printf("fastpfor/encodePage: bestb = %d, bestc = %d, maxb = %d\n", bestb, bestc, maxb)
		tmpbestb := bestb
		this.byteContainer.Put(byte(bestb))
		this.byteContainer.Put(byte(bestc))

		if bestc > 0 {
			this.byteContainer.Put(byte(maxb))
			index := uint32(maxb - bestb)

			if int(this.dataPointers[index] + uint32(bestc)) >= len(this.dataToBePacked[index]) {
				newSize := 2*(this.dataPointers[index] + uint32(bestc))

				// make sure it is a multiple of 32.
				// there might be a better way to do this
				newSize = encoding.CeilBy(newSize, 32)
				this.dataToBePacked[index] = append(make([]int32, 0, newSize), this.dataToBePacked[index]...)
			}

			for k := uint32(0); k < DefaultBlockSize; k++ {
				if in[k+tmpinpos] >> bestb != 0 {
					// we have an exception
					this.byteContainer.Put(byte(k))
					this.dataToBePacked[index][this.dataPointers[index]] = in[k+tmpinpos] >> tmpbestb
					this.dataPointers[index] += 1
				}
			}
		}

		for k := uint32(0); k < 128; k += 32 {
			bitpacking.FastPack(in, int(tmpinpos + k), out, int(tmpoutpos), int(tmpbestb))
			tmpoutpos += uint32(tmpbestb)
		}
	}

	inpos.Set(int(tmpinpos))
	out[headerpos] = tmpoutpos - headerpos
	log.Printf("fastpfor/encodePage: headerpos = %d, tmpoutpos = %d, out[headerpos] = %d\n", headerpos, tmpoutpos, out[headerpos])

	for this.byteContainer.Position() & 3 != 0 {
		log.Printf("fastpfor/encodePage: byteContainer.Position() = %d\n", this.byteContainer.Position())
		this.byteContainer.Put(0)
	}

	bytesize := uint32(this.byteContainer.Position())
	out[tmpoutpos] = bytesize
	tmpoutpos += 1
	howmanyints := bytesize / 4
	log.Printf("fastpfor/encodePage: bytesize = %d, howmanyints = %d\n", bytesize, howmanyints)

	this.byteContainer.Flip()
	this.byteContainer.AsUint32Buffer().GetUint32s(out, int(tmpoutpos), int(howmanyints))
	log.Printf("fastpfor/encodePage: byteContainer[%d:%d] = %v\n", tmpoutpos, tmpoutpos+howmanyints, out[tmpoutpos:tmpoutpos+howmanyints])
	tmpoutpos += howmanyints

	bitmap := uint32(0)
	for k, v := range this.dataPointers[1:33] {
		if v != 0 {
			bitmap |= (1 << uint32(k) - 1)
		}
	}

	out[tmpoutpos] = bitmap
	tmpoutpos += 1
	log.Printf("fastpfor/encodePage: bitmap = %d, tmpoutpos = %d\n", bitmap, tmpoutpos)

	for k, v := range this.dataPointers[1:32] {
		if v != 0 {
			out[tmpoutpos] = v // size
			for j := 0; j < int(v); j += 32 {
				bitpacking.FastPack(this.dataToBePacked[k], j, out, int(tmpoutpos), k)
				tmpoutpos += uint32(k)
			}
		}
	}

	outpos.Set(int(tmpoutpos))

	return nil
}

func (this *FastPFOR) decodePage(in []int32, inpos *encoding.Cursor, out []int32, outpos *encoding.Cursor, thissize int) error {
	log.Printf("fastpfor/decodePage: in[%d:%d] = %v\n", inpos.Get(), inpos.Get()+thissize, in[inpos.Get():inpos.Get()+thissize])
	initpos := uint32(inpos.Get())
	wheremeta := in[initpos]
	inpos.Increment()

	inexcept := initpos + wheremeta
	bytesize := in[inexcept]
	inexcept += 1

	log.Printf("fastpfor/decodePage: initpos = %d, wheremeta = %d, inexcept = %d, bytesize = %d\n", initpos, wheremeta, inexcept, bytesize)
	this.byteContainer.Clear()
	if err := this.byteContainer.AsUint32Buffer().PutUint32s(in, int(inexcept), int(bytesize/4)); err != nil {
		log.Printf("fastpfor/decodePage: error with PutUint32s, %v\n", err)
		return err
	}

	inexcept += bytesize / 4
	bitmap := in[inexcept]
	inexcept += 1

	for k := uint32(1); k <= 31; k++ {
		if bitmap & (1 << k - 1) != 0 {
			size := in[inexcept]
			log.Printf("fastpfor/decodePage: size = %d\n", size)

			if uint32(len(this.dataToBePacked[k])) < size {
				this.dataToBePacked[k] = make([]int32, encoding.CeilBy(size, 32))
			}

			for j := uint32(0); j < size; j += 32 {
				bitpacking.FastUnpack(in, int(inexcept), this.dataToBePacked[k], int(j), int(k))
				inexcept += k
			}
		}
	}

	copy(this.dataPointers, zeroDataPointers)
	tmpoutpos := uint32(outpos.Get())
	tmpinpos := uint32(inpos.Get())

	log.Printf("fastpfor/decodePage: tmpoutpos = %d, tmpinpos = %d\n", tmpoutpos, tmpinpos)

	run := 0
	run_end := thissize / DefaultBlockSize
	for run < run_end {
		log.Printf("fastpfor/decodePage: run = %d, run_end = %d, byteContainer.Position() = %d\n", run, run_end, this.byteContainer.Position())

		var err error
		var bestb uint32
		if bestb, err = this.byteContainer.GetAsUint32(); err != nil {
			return err
		}

		var cexcept uint32
		if cexcept, err = this.byteContainer.GetAsUint32(); err != nil {
			return err
		}

		log.Printf("fastpfor/decodePage: bestb = %d, cexcept = %d\n", bestb, cexcept)

		origInpos := tmpinpos

		for k := uint32(0); k < 128; k += 32 {
			bitpacking.FastUnpack(in, int(tmpinpos), out, int(tmpoutpos+k), int(bestb))
			tmpinpos += bestb
		}

		log.Printf("fastpfor/decodePage: in[%d:%d] = %v\n", origInpos, tmpinpos, in[origInpos:tmpinpos])

		if cexcept > 0 {
			var maxbits uint32
			if maxbits, err = this.byteContainer.GetAsUint32(); err != nil {
				return err
			}

			index := maxbits - bestb

			for k := uint32(0); k < cexcept; k++ {
				var pos uint32
				if pos, err = this.byteContainer.GetAsUint32(); err != nil {
					return err
				}

				exceptvalue := this.dataToBePacked[index][this.dataPointers[index]]
				this.dataPointers[index] += 1
				out[pos + tmpoutpos] |= exceptvalue << bestb
			}
		}

		outpos.Set(int(tmpoutpos))
		inpos.Set(int(inexcept))

		run += 1
		tmpoutpos += DefaultBlockSize
	}

	return nil
}
