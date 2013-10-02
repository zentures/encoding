/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package composition

import (
	"errors"
	"github.com/reducedb/encoding"
)

type IntegratedComposition struct {
	f1 encoding.Integer
	f2 encoding.Integer
}

var _ encoding.Integer = (*IntegratedComposition)(nil)

func NewIntegratedComposition(f1 encoding.Integer, f2 encoding.Integer) encoding.Integer {
	return &IntegratedComposition{
		f1: f1,
		f2: f2,
	}
}

func (this *IntegratedComposition) Compress(in []uint32, inpos *encoding.Cursor, inlength int, out []uint32, outpos *encoding.Cursor) error {
	if inlength == 0 {
		return errors.New("composition/Compress: inlength = 0. No work done.")
	}

	init := inpos.Get()
	this.f1.Compress(in, inpos, inlength, out, outpos)
	if outpos.Get() == 0 {
		out[0] = 0
		outpos.Increment()
	}
	//fmt.Printf("composition/Compress: f1 inpos = %d, outpos = %d, inlength = %d\n", inpos.Get(), outpos.Get(), inlength)

	inlength -= inpos.Get() - init
	this.f2.Compress(in, inpos, inlength, out, outpos)
	//fmt.Printf("composition/Compress: f2 inpos = %d, outpos = %d, inlength = %d\n", inpos.Get(), outpos.Get(), inlength)

	return nil
}

func (this *IntegratedComposition) Uncompress(in []uint32, inpos *encoding.Cursor, inlength int, out []uint32, outpos *encoding.Cursor) error {
	if inlength == 0 {
		return errors.New("composition/Uncompress: inlength = 0. No work done.")
	}

	init := inpos.Get()
	this.f1.Uncompress(in, inpos, inlength, out, outpos)
	//fmt.Printf("composition/Uncompress: f1 inpos = %d, outpos = %d, inlength = %d\n", inpos.Get(), outpos.Get(), inlength)
	inlength -= inpos.Get() - init
	this.f2.Uncompress(in, inpos, inlength, out, outpos)
	//fmt.Printf("composition/Uncompress: f2 inpos = %d, outpos = %d, inlength = %d\n", inpos.Get(), outpos.Get(), inlength)

	return nil
}
