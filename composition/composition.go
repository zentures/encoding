/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package composition

import (
	"errors"

	"github.com/dataence/encoding"
	"github.com/dataence/encoding/cursor"
)

type Composition struct {
	f1 encoding.Integer
	f2 encoding.Integer
}

var _ encoding.Integer = (*Composition)(nil)

func New(f1 encoding.Integer, f2 encoding.Integer) encoding.Integer {
	return &Composition{
		f1: f1,
		f2: f2,
	}
}

func (this *Composition) Compress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {
	if inlength == 0 {
		return errors.New("composition/Compress: inlength = 0. No work done.")
	}

	init := inpos.Get()
	this.f1.Compress(in, inpos, inlength, out, outpos)
	if outpos.Get() == 0 {
		out[0] = 0
		outpos.Increment()
	}
	//log.Printf("composition/Compress: f1 inpos = %d, outpos = %d, inlength = %d\n", inpos.Get(), outpos.Get(), inlength)

	inlength -= inpos.Get() - init
	this.f2.Compress(in, inpos, inlength, out, outpos)
	//log.Printf("composition/Compress: f2 inpos = %d, outpos = %d, inlength = %d\n", inpos.Get(), outpos.Get(), inlength)

	return nil
}

func (this *Composition) Uncompress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error {
	if inlength == 0 {
		return errors.New("composition/Uncompress: inlength = 0. No work done.")
	}

	init := inpos.Get()
	this.f1.Uncompress(in, inpos, inlength, out, outpos)
	//log.Printf("composition/Uncompress: f1 inpos = %d, outpos = %d, inlength = %d\n", inpos.Get(), outpos.Get(), inlength)
	inlength -= inpos.Get() - init
	this.f2.Uncompress(in, inpos, inlength, out, outpos)
	//log.Printf("composition/Uncompress: f2 inpos = %d, outpos = %d, inlength = %d\n", inpos.Get(), outpos.Get(), inlength)

	return nil
}
