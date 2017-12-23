/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package bp32

import (
	"errors"

	"github.com/dataence/encoding"
	"github.com/dataence/encoding/bitpacking"
	"github.com/dataence/encoding/cursor"
)

const (
	DefaultBlockSize = 128
	DefaultPageSize  = 65536
)

type BP32 struct {
}

var _ encoding.Integer = (*BP32)(nil)

func New() encoding.Integer {
	return &BP32{}
}

func (this *BP32) Compress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {
	//log.Printf("bp32/Compress: before inlength = %d\n", inlength)

	inlength = encoding.FloorBy(inlength, DefaultBlockSize)

	if inlength == 0 {
		return errors.New("BP32/Compress: block size less than 128. No work done.")
	}

	//log.Printf("bp32/Compress: after inlength = %d, len(in) = %d\n", inlength, len(in))

	out[outpos.Get()] = int32(inlength)
	outpos.Increment()

	tmpoutpos := outpos.Get()
	initoffset := int32(0)
	s := inpos.Get()
	finalinpos := s + inlength

	for ; s < finalinpos; s += DefaultBlockSize {
		mbits1 := encoding.DeltaMaxBits(initoffset, in[s:s+32])
		initoffset2 := in[s+31]
		mbits2 := encoding.DeltaMaxBits(initoffset2, in[s+32:s+2*32])
		initoffset3 := in[s+32+31]
		mbits3 := encoding.DeltaMaxBits(initoffset3, in[s+2*32:s+3*32])
		initoffset4 := in[s+2*32+31]
		mbits4 := encoding.DeltaMaxBits(initoffset4, in[s+3*32:s+4*32])

		//log.Printf("bp32/Compress: tmpoutpos = %d, s = %d\n", tmpoutpos, s)

		out[tmpoutpos] = (mbits1 << 24) | (mbits2 << 16) | (mbits3 << 8) | mbits4
		tmpoutpos += 1

		//log.Printf("bp32/Compress: mbits1 = %d, mbits2 = %d, mbits3 = %d, mbits4 = %d, s = %d\n", mbits1, mbits2, mbits3, mbits4, out[tmpoutpos-1])

		bitpacking.DeltaPack(initoffset, in, s, out, tmpoutpos, int(mbits1))
		//encoding.PrintUint32sInBits(in, s, 32)
		//encoding.PrintUint32sInBits(out, tmpoutpos, int(mbits1))
		tmpoutpos += int(mbits1)

		bitpacking.DeltaPack(initoffset2, in, s+32, out, tmpoutpos, int(mbits2))
		//encoding.PrintUint32sInBits(in, s+32, 32)
		//encoding.PrintUint32sInBits(out, tmpoutpos, int(mbits2))
		tmpoutpos += int(mbits2)

		bitpacking.DeltaPack(initoffset3, in, s+2*32, out, tmpoutpos, int(mbits3))
		//encoding.PrintUint32sInBits(in, s+2*32, 32)
		//encoding.PrintUint32sInBits(out, tmpoutpos, int(mbits3))
		tmpoutpos += int(mbits3)

		bitpacking.DeltaPack(initoffset4, in, s+3*32, out, tmpoutpos, int(mbits4))
		//encoding.PrintUint32sInBits(in, s+3*32, 32)
		//encoding.PrintUint32sInBits(out, tmpoutpos, int(mbits4))
		tmpoutpos += int(mbits4)

		initoffset = in[s+3*32+31]
	}

	inpos.Add(inlength)
	outpos.Set(tmpoutpos)

	return nil
}

func (this *BP32) Uncompress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {
	if inlength == 0 {
		return errors.New("BP32/Uncompress: Length is 0. No work done.")
	}

	outlength := in[inpos.Get()]
	inpos.Increment()

	tmpinpos := inpos.Get()
	initoffset := int32(0)

	//log.Printf("bp32/Uncompress: outlength = %d, inpos = %d, outpos = %d\n", outlength, inpos.Get(), outpos.Get())
	for s := outpos.Get(); s < outpos.Get()+int(outlength); s += 32 * 4 {
		tmp := in[tmpinpos]
		mbits1 := tmp >> 24
		mbits2 := (tmp >> 16) & 0xFF
		mbits3 := (tmp >> 8) & 0xFF
		mbits4 := (tmp) & 0xFF

		//log.Printf("bp32/Uncopmress: mbits1 = %d, mbits2 = %d, mbits3 = %d, mbits4 = %d, s = %d\n", mbits1, mbits2, mbits3, mbits4, s)
		tmpinpos += 1

		bitpacking.DeltaUnpack(initoffset, in, tmpinpos, out, s, int(mbits1))
		tmpinpos += int(mbits1)
		initoffset = out[s+31]
		//log.Printf("bp32/Uncompress: out = %v\n", out)

		bitpacking.DeltaUnpack(initoffset, in, tmpinpos, out, s+32, int(mbits2))
		tmpinpos += int(mbits2)
		initoffset = out[s+32+31]
		//log.Printf("bp32/Uncompress: out = %v\n", out)

		bitpacking.DeltaUnpack(initoffset, in, tmpinpos, out, s+2*32, int(mbits3))
		tmpinpos += int(mbits3)
		initoffset = out[s+2*32+31]
		//log.Printf("bp32/Uncompress: out = %v\n", out)

		bitpacking.DeltaUnpack(initoffset, in, tmpinpos, out, s+3*32, int(mbits4))
		tmpinpos += int(mbits4)
		initoffset = out[s+3*32+31]
		//log.Printf("bp32/Uncompress: out = %v\n", out)
	}

	outpos.Add(int(outlength))
	inpos.Set(tmpinpos)

	return nil
}
