/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package variablebyte

import (
	"errors"

	"github.com/dataence/bytebuffer"
	"github.com/dataence/encoding"
	"github.com/dataence/encoding/cursor"
)

type VariableByte struct {
}

var _ encoding.Integer = (*VariableByte)(nil)

func New() encoding.Integer {
	return &VariableByte{}
}

func (this *VariableByte) Compress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {
	if inlength == 0 {
		return errors.New("variablebyte/Compress: inlength = 0. No work done.")
	}

	//fmt.Printf("variablebyte/Compress: after inlength = %d\n", inlength)

	buf := bytebuffer.NewByteBuffer(inlength * 8)
	initoffset := int32(0)

	tmpinpos := inpos.Get()
	for _, v := range in[tmpinpos : tmpinpos+inlength] {
		val := uint32(v - initoffset)
		initoffset = v

		for val >= 0x80 {
			buf.Put(byte(val) | 0x80)
			val >>= 7
		}
		buf.Put(byte(val))
	}

	for buf.Position()%4 != 0 {
		//fmt.Printf("variablebyte/Compress: putting 128\n")
		buf.Put(128)
	}

	length := buf.Position()
	buf.Flip()
	ibuf := buf.AsInt32Buffer()
	//fmt.Printf("variablebyte/Compress: l = %d, outpos = %d, ibuf = %v, buf = %v\n", length/4, outpos.Get(), ibuf, buf)
	err := ibuf.GetInt32s(out, outpos.Get(), length/4)
	if err != nil {
		//fmt.Printf("variablebyte/Compress: error with GetUint32s - %v\n", err)
		return err
	}
	outpos.Add(length / 4)
	inpos.Add(inlength)
	//fmt.Printf("variablebyte/Compress: out = %v\n", out)

	return nil
}

func (this *VariableByte) Uncompress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {
	if inlength == 0 {
		return errors.New("variablebyte/Uncompress: inlength = 0. No work done.")
	}

	//fmt.Printf("variablebyte/Uncompress: after inlength = %d\n", inlength)

	s := uint(0)
	p := inpos.Get()
	finalp := inpos.Get() + inlength
	tmpoutpos := outpos.Get()
	initoffset := int32(0)
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
			out[tmpoutpos] = v + initoffset
			initoffset = out[tmpoutpos]
			tmpoutpos += 1
			v = 0
			shift = 0
		} else {
			shift += 7
		}

		outpos.Set(tmpoutpos)
		inpos.Add(inlength)
	}

	return nil
}
