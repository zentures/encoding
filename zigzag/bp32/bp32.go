/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package bp32

import (
	"errors"
	"github.com/reducedb/encoding"
	"github.com/reducedb/encoding/cursor"
	"github.com/reducedb/encoding/bitpacking"
)

const (
	DefaultBlockSize = 128
	DefaultPageSize = 65536
)

type BP32 struct {

}

var _ encoding.Integer = (*BP32)(nil)

func New() encoding.Integer {
	return &BP32{}
}

func (this *BP32) Compress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {
	//log.Printf("zigzag_bp32/Compress: before inlength = %d\n", inlength)

	inlength = encoding.FloorBy(inlength, DefaultBlockSize)

	if inlength == 0 {
		return errors.New("zigzag_bp32/Compress: block size less than 128. No work done.")
	}

	//log.Printf("zigzag_bp32/Compress: after inlength = %d, len(in) = %d\n", inlength, len(in))

	out[outpos.Get()] = int32(inlength)
	outpos.Increment()

	tmpoutpos := outpos.Get()
	s := inpos.Get()
	finalinpos := s + inlength
	delta := make([]int32, DefaultBlockSize)

	for ; s < finalinpos; s += DefaultBlockSize {
		encoding.ZigZagDelta(in[s:s+DefaultBlockSize], delta)
		//log.Printf("zigzag_bp32/Compress: in = %v\n", in[s:s+DefaultBlockSize])
		//log.Printf("zigzag_bp32/Compress: delta = %v\n", delta)

		mbits1 := encoding.MaxBits(delta[0:32])
		mbits2 := encoding.MaxBits(delta[32:64])
		mbits3 := encoding.MaxBits(delta[64:96])
		mbits4 := encoding.MaxBits(delta[96:128])

		//log.Printf("zigzag_bp32/Compress: tmpoutpos = %d, s = %d\n", tmpoutpos, s)

		out[tmpoutpos] = (mbits1<<24) | (mbits2<<16) | (mbits3<<8) | mbits4
		tmpoutpos += 1

		//log.Printf("zigzag_bp32/Compress: mbits1 = %d, mbits2 = %d, mbits3 = %d, mbits4 = %d, s = %d\n", mbits1, mbits2, mbits3, mbits4, out[tmpoutpos-1])

		bitpacking.FastPackWithoutMask(delta, 0, out, tmpoutpos, int(mbits1))
		//encoding.PrintUint32sInBits(in[s:s+32])
		//encoding.PrintUint32sInBits(out[tmpoutpos:tmpoutpos+int(mbits1]))
		tmpoutpos += int(mbits1)

		bitpacking.FastPackWithoutMask(delta, 32, out, tmpoutpos, int(mbits2))
		//encoding.PrintUint32sInBits(in, s+32, 32)
		//encoding.PrintUint32sInBits(out, tmpoutpos, int(mbits2))
		tmpoutpos += int(mbits2)

		bitpacking.FastPackWithoutMask(delta, 64, out, tmpoutpos, int(mbits3))
		//encoding.PrintUint32sInBits(in, s+2*32, 32)
		//encoding.PrintUint32sInBits(out, tmpoutpos, int(mbits3))
		tmpoutpos += int(mbits3)

		bitpacking.FastPackWithoutMask(delta, 96, out, tmpoutpos, int(mbits4))
		//encoding.PrintUint32sInBits(in, s+3*32, 32)
		//encoding.PrintUint32sInBits(out, tmpoutpos, int(mbits4))
		tmpoutpos += int(mbits4)

		//log.Printf("zigzag_bp32/Compress: out = %v\n", out[s:s+DefaultBlockSize])
	}

	inpos.Add(inlength)
	outpos.Set(tmpoutpos)

	return nil
}

func (this *BP32) Uncompress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {
	if inlength == 0 {
		return errors.New("zigzag_bp32/Uncompress: Length is 0. No work done.")
	}

	outlength := int(in[inpos.Get()])
	inpos.Increment()

	tmpinpos := inpos.Get()
	s := outpos.Get()
	finalinpos := s + outlength
	delta := make([]int32, DefaultBlockSize)

	//log.Printf("zigzag_bp32/Uncompress: outlength = %d, inpos = %d, outpos = %d\n", outlength, inpos.Get(), outpos.Get())
	for ; s < finalinpos; s += DefaultBlockSize {
		tmp := in[tmpinpos]
		mbits1 := tmp>>24
		mbits2 := (tmp>>16) & 0xFF
		mbits3 := (tmp>>8) & 0xFF
		mbits4 := (tmp) & 0xFF

		//log.Printf("zigzag_bp32/Uncopmress: mbits1 = %d, mbits2 = %d, mbits3 = %d, mbits4 = %d, s = %d\n", mbits1, mbits2, mbits3, mbits4, s)
		tmpinpos += 1

		bitpacking.FastUnpack(in, tmpinpos, delta, 0, int(mbits1))
		tmpinpos += int(mbits1)
		//log.Printf("zigzag_bp32/Uncompress: delta = %v\n", out)

		bitpacking.FastUnpack(in, tmpinpos, delta, 32, int(mbits2))
		tmpinpos += int(mbits2)
		//log.Printf("zigzag_bp32/Uncompress: delta = %v\n", out)

		bitpacking.FastUnpack(in, tmpinpos, delta, 64, int(mbits3))
		tmpinpos += int(mbits3)
		//log.Printf("zigzag_bp32/Uncompress: delta = %v\n", out)

		bitpacking.FastUnpack(in, tmpinpos, delta, 96, int(mbits4))
		tmpinpos += int(mbits4)

		encoding.InverseZigZagDelta(delta, out[s:s+DefaultBlockSize])

		//log.Printf("zigzag_bp32/Uncompress: delta = %v\n", delta)
		//log.Printf("zigzag_bp32/Uncompress: out = %v\n", out[s:s+DefaultBlockSize])

	}

	outpos.Add(outlength)
	inpos.Set(tmpinpos)

	return nil
}
