/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package variablebyte

import (
	"errors"
	"github.com/reducedb/encoding"
	"github.com/reducedb/encoding/buffers"
	"github.com/reducedb/encoding/cursor"
)

type VariableByte struct {
}

var _ encoding.Integer = (*VariableByte)(nil)

func New() encoding.Integer {
	return &VariableByte{}
}

func (this *VariableByte) Compress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {
	if inlength == 0 {
		return errors.New("VariableByte/Compress: inlength = 0. No work done.")
	}

	//fmt.Printf("VariableByte/Compress: after inlength = %d\n", inlength)

	buf := buffers.NewByteBuffer(inlength * 8)
	tmpinpos := inpos.Get()

	for _, v := range in[tmpinpos : tmpinpos+inlength] {
		val := uint32(v)

		for val >= 0x80 {
			buf.Put(byte(val) | 0x80)
			val >>= 7
		}
		buf.Put(byte(val))
	}

	for buf.Position()%4 != 0 {
		//fmt.Printf("VariableByte/Compress: putting 128\n")
		buf.Put(128)
	}

	length := buf.Position()
	buf.Flip()
	ibuf := buf.AsInt32Buffer()
	//fmt.Printf("VariableByte/Compress: l = %d, outpos = %d, ibuf = %v, buf = %v\n", length/4, outpos.Get(), ibuf, buf)
	err := ibuf.GetInt32s(out, outpos.Get(), length/4)
	if err != nil {
		//fmt.Printf("VariableByte/Compress: error with GetUint32s - %v\n", err)
		return err
	}
	outpos.Add(length / 4)
	inpos.Add(inlength)
	//fmt.Printf("VariableByte/Compress: out = %v\n", out)

	return nil
}

func (this *VariableByte) Uncompress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {
	if inlength == 0 {
		return errors.New("VariableByte/Uncompress: inlength = 0. No work done.")
	}

	//fmt.Printf("VariableByte/Uncompress: after inlength = %d\n", inlength)

	s := uint(0)
	p := inpos.Get()
	finalp := p + inlength
	tmpoutpos := outpos.Get()
	v := int32(0)
	shift := uint(0)

	for p < finalp {
		c := in[p] >> (24 - s)
		s += 8

		if s == 32 {
			s = 0
			p += 1
		}

		v += ((c & 127) << shift)
		if c&128 == 0 {
			out[tmpoutpos] = v
			tmpoutpos += 1
			v = 0
			shift = 0
		} else {
			shift += 7
		}
	}

	outpos.Set(tmpoutpos)
	inpos.Add(inlength)

	return nil
}
