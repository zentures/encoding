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
)

type IntegratedVariableByte struct {

}

var _ encoding.Integer = (*IntegratedVariableByte)(nil)

func NewIntegratedVariableByte() encoding.Integer {
	return &IntegratedVariableByte{}
}

func (this *IntegratedVariableByte) Compress(in []uint32, inpos *encoding.Cursor, inlength int, out []uint32, outpos *encoding.Cursor) error {
	if inlength == 0 {
		return errors.New("variablebyte/Compress: inlength = 0. No work done.")
	}

	//fmt.Printf("variablebyte/Compress: after inlength = %d\n", inlength)

	buf := buffers.NewByteBuffer(inlength*8)
	initoffset := uint32(0)

	for k := inpos.Get(); k < inpos.Get() + inlength; k++ {
		val := in[k] - initoffset
		//fmt.Printf("variablebyte/Compress: val = %d, initoffset = %d\n", val, initoffset)
		initoffset = in[k]

		// This section emulates a do..while loop
		b := val & 127
		//fmt.Printf("variablebyte/Compress: before val = %d, b = %d\n", val, b)
		val = val>>7

		if val != 0 {
			b |= 128
		}
		//fmt.Printf("variablebyte/Compress: after val = %d, b = %d\n", val, b)

		buf.Put(byte(b))

		for val != 0 {
			b = val & 127
			//fmt.Printf("variablebyte/Compress: before val = %d, b = %d\n", val, b)
			val = val>>7

			if val != 0 {
				b |= 128
			}
			//fmt.Printf("variablebyte/Compress: after val = %d, b = %d\n", val, b)

			buf.Put(byte(b))
		}
	}

	for buf.Position()%4 != 0 {
		//fmt.Printf("variablebyte/Compress: putting 128\n")
		buf.Put(128)
	}

	length := buf.Position()
	buf.Flip()
	ibuf := buf.AsUint32Buffer()
	//fmt.Printf("variablebyte/Compress: l = %d, outpos = %d, ibuf = %v, buf = %v\n", length/4, outpos.Get(), ibuf, buf)
	err := ibuf.GetUint32s(out, outpos.Get(), length/4)
	if err != nil {
		//fmt.Printf("variablebyte/Compress: error with GetUint32s - %v\n", err)
		return err
	}
	outpos.Add(length/4)
	inpos.Add(inlength)
	//fmt.Printf("variablebyte/Compress: out = %v\n", out)

	return nil
}

func (this *IntegratedVariableByte) Uncompress(in []uint32, inpos *encoding.Cursor, inlength int, out []uint32, outpos *encoding.Cursor) error {
	if inlength == 0 {
		return errors.New("variablebyte/Uncompress: inlength = 0. No work done.")
	}

	//fmt.Printf("variablebyte/Uncompress: after inlength = %d\n", inlength)

	s := 0
	p := inpos.Get()
	finalp := inpos.Get() + inlength
	tmpoutpos := outpos.Get()
	initoffset := uint32(0)
	v := uint32(0)
	shift := uint(0)

	for p < finalp {
		c := in[p]>>uint(24 - s)
		s += 8

		if s == 32 {
			s = 0
			p += 1
		}

		v += ((c & 127)<<shift)
		if c & 128 == 0 {
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

