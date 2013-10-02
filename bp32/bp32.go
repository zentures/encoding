/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package bp32

import (
	"errors"
	"github.com/reducedb/encoding"
)

const (
	DefaultBlockSize uint32 = 128
	DefaultPageSize uint32  = 65536
)

type IntegratedBP32 struct {

}

var _ encoding.Integer = (*IntegratedBP32)(nil)

func NewIntegratedBP32() encoding.Integer {
	return &IntegratedBP32{}
}

func (this *IntegratedBP32) Compress(in []uint32, inpos *encoding.Cursor, inlength int, out []uint32, outpos *encoding.Cursor) error {
	//log.Printf("bp32/Compress: before inlength = %d\n", inlength)

	inlength = int(encoding.FloorBy(uint32(inlength), DefaultBlockSize))

	if inlength == 0 {
		return errors.New("BP32/Compress: block size less than 128. No work done.")
	}

	//log.Printf("bp32/Compress: after inlength = %d, len(in) = %d\n", inlength, len(in))

	out[outpos.Get()] = uint32(inlength)
	outpos.Increment()

	tmpoutpos := outpos.Get()
	initoffset := uint32(0)

	for s := inpos.Get(); s < inpos.Get() + inlength; s += 32*4 {
		mbits1 := encoding.MaxDiffBits(initoffset, in, s, 32)
		initoffset2 := in[s + 31]
		mbits2 := encoding.MaxDiffBits(initoffset2, in, s + 32, 32)
		initoffset3 := in[s + 32 + 31]
		mbits3 := encoding.MaxDiffBits(initoffset3, in, s + 2*32, 32)
		initoffset4 := in[s + 2*32 + 31]
		mbits4 := encoding.MaxDiffBits(initoffset4, in, s + 3*32, 32)
		
		//log.Printf("bp32/Compress: tmpoutpos = %d, s = %d\n", tmpoutpos, s)

		out[tmpoutpos] = (mbits1<<24) | (mbits2<<16) | (mbits3<<8) | mbits4
		tmpoutpos += 1

		//log.Printf("bp32/Compress: mbits1 = %d, mbits2 = %d, mbits3 = %d, mbits4 = %d, s = %d\n", mbits1, mbits2, mbits3, mbits4, out[tmpoutpos-1])

		encoding.IntegratedPack(initoffset, in, s, out, tmpoutpos, int(mbits1))
		//encoding.PrintUint32sInBits(in, s, 32)
		//encoding.PrintUint32sInBits(out, tmpoutpos, int(mbits1))
		tmpoutpos += int(mbits1)

		encoding.IntegratedPack(initoffset2, in, s + 32, out, tmpoutpos, int(mbits2))
		//encoding.PrintUint32sInBits(in, s+32, 32)
		//encoding.PrintUint32sInBits(out, tmpoutpos, int(mbits2))
		tmpoutpos += int(mbits2)

		encoding.IntegratedPack(initoffset3, in, s + 2*32, out, tmpoutpos, int(mbits3))
		//encoding.PrintUint32sInBits(in, s+2*32, 32)
		//encoding.PrintUint32sInBits(out, tmpoutpos, int(mbits3))
		tmpoutpos += int(mbits3)

		encoding.IntegratedPack(initoffset4, in, s + 3*32, out, tmpoutpos, int(mbits4))
		//encoding.PrintUint32sInBits(in, s+3*32, 32)
		//encoding.PrintUint32sInBits(out, tmpoutpos, int(mbits4))
		tmpoutpos += int(mbits4)

		initoffset = in[s + 3*32 + 31]
	}

	inpos.Add(inlength)
	outpos.Set(tmpoutpos)

	return nil
}

func (this *IntegratedBP32) Uncompress(in []uint32, inpos *encoding.Cursor, inlength int, out []uint32, outpos *encoding.Cursor) error {
	if inlength == 0 {
		return errors.New("BP32/Uncompress: Length is 0. No work done.")
	}

	outlength := in[inpos.Get()]
	inpos.Increment()

	tmpinpos := inpos.Get()
	initoffset := uint32(0)

	//log.Printf("bp32/Uncompress: outlength = %d, inpos = %d, outpos = %d\n", outlength, inpos.Get(), outpos.Get())
	for s := outpos.Get(); s < outpos.Get() + int(outlength); s += 32*4 {
		mbits1 := in[tmpinpos]>>24
		mbits2 := (in[tmpinpos]>>16) & 0xFF
		mbits3 := (in[tmpinpos]>>8) & 0xFF
		mbits4 := (in[tmpinpos]) & 0xFF

		//log.Printf("bp32/Uncopmress: mbits1 = %d, mbits2 = %d, mbits3 = %d, mbits4 = %d, s = %d\n", mbits1, mbits2, mbits3, mbits4, s)
		tmpinpos += 1

		encoding.IntegratedUnpack(initoffset, in, tmpinpos, out, s, int(mbits1))
		tmpinpos += int(mbits1)
		initoffset = out[s + 31]
		//log.Printf("bp32/Uncompress: out = %v\n", out)

		encoding.IntegratedUnpack(initoffset, in, tmpinpos, out, s + 32, int(mbits2))
		tmpinpos += int(mbits2)
		initoffset = out[s + 32 + 31]
		//log.Printf("bp32/Uncompress: out = %v\n", out)

		encoding.IntegratedUnpack(initoffset, in, tmpinpos, out, s + 2*32, int(mbits3))
		tmpinpos += int(mbits3)
		initoffset = out[s + 2*32 + 31]
		//log.Printf("bp32/Uncompress: out = %v\n", out)

		encoding.IntegratedUnpack(initoffset, in, tmpinpos, out, s + 3*32, int(mbits4))
		tmpinpos += int(mbits4)
		initoffset = out[s + 3*32 + 31]
		//log.Printf("bp32/Uncompress: out = %v\n", out)
	}

	outpos.Add(int(outlength))
	inpos.Set(tmpinpos)

	return nil
}
