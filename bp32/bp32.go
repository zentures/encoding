/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

// Package bp32 is an implementation of the binary packing integer compression
// algorithm in in Go (also known as PackedBinary) using 32-integer blocks.
// It is mostly suitable for arrays containing small positive integers.
// Given a list of sorted integers, you should first compute the successive
// differences prior to compression.
// For details, please see
// Daniel Lemire and Leonid Boytsov, Decoding billions of integers per second
// through vectorization Software: Practice & Experience
// http://onlinelibrary.wiley.com/doi/10.1002/spe.2203/abstract or
//	http://arxiv.org/abs/1209.2137
package bp32

import (
	"errors"

	"github.com/dataence/encoding"
	"github.com/dataence/encoding/bitpacking"
	"github.com/dataence/encoding/cursor"
)

const (
	DefaultBlockSize = 128
)

type BP32 struct {
}

var _ encoding.Integer = (*BP32)(nil)

func New() encoding.Integer {
	return &BP32{}
}

func (this *BP32) Compress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {

	inlength = encoding.FloorBy(inlength, DefaultBlockSize)

	if inlength == 0 {
		return errors.New("BP32/Compress: block size less than 128. No work done.")
	}

	out[outpos.Get()] = int32(inlength)
	outpos.Increment()

	tmpoutpos := outpos.Get()
	s := inpos.Get()
	finalinpos := s + inlength

	for ; s < finalinpos; s += DefaultBlockSize {
		mbits1 := encoding.MaxBits(in[s : s+32])
		mbits2 := encoding.MaxBits(in[s+32 : s+2*32])
		mbits3 := encoding.MaxBits(in[s+2*32 : s+3*32])
		mbits4 := encoding.MaxBits(in[s+3*32 : s+4*32])

		out[tmpoutpos] = (mbits1 << 24) | (mbits2 << 16) | (mbits3 << 8) | mbits4
		tmpoutpos += 1
		bitpacking.FastPackWithoutMask(in, s, out, tmpoutpos, int(mbits1))
		tmpoutpos += int(mbits1)
		bitpacking.FastPackWithoutMask(in, s+32, out, tmpoutpos, int(mbits2))
		tmpoutpos += int(mbits2)
		bitpacking.FastPackWithoutMask(in, s+2*32, out, tmpoutpos, int(mbits3))
		tmpoutpos += int(mbits3)
		bitpacking.FastPackWithoutMask(in, s+3*32, out, tmpoutpos, int(mbits4))
		tmpoutpos += int(mbits4)
	}

	inpos.Add(inlength)
	outpos.Set(tmpoutpos)

	return nil
}

func (this *BP32) Uncompress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {
	if inlength == 0 {
		return errors.New("BP32/Uncompress: Length is 0. No work done.")
	}

	outlength := int(in[inpos.Get()])
	inpos.Increment()

	tmpinpos := inpos.Get()

	for s := outpos.Get(); s < outpos.Get()+outlength; s += 32 * 4 {
		tmp := in[tmpinpos]
		mbits1 := tmp >> 24
		mbits2 := (tmp >> 16) & 0xFF
		mbits3 := (tmp >> 8) & 0xFF
		mbits4 := (tmp) & 0xFF

		tmpinpos += 1

		bitpacking.FastUnpack(in, tmpinpos, out, s, int(mbits1))
		tmpinpos += int(mbits1)

		bitpacking.FastUnpack(in, tmpinpos, out, s+32, int(mbits2))
		tmpinpos += int(mbits2)

		bitpacking.FastUnpack(in, tmpinpos, out, s+2*32, int(mbits3))
		tmpinpos += int(mbits3)

		bitpacking.FastUnpack(in, tmpinpos, out, s+3*32, int(mbits4))
		tmpinpos += int(mbits4)
	}

	outpos.Add(outlength)
	inpos.Set(tmpinpos)

	return nil
}
