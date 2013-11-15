/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package cursor

type Cursor struct {
	value int
}

func New() *Cursor {
	return &Cursor{
		value: 0,
	}
}

func (this *Cursor) Get() int {
	return this.value
}

func (this *Cursor) Set(i int) {
	this.value = i
}

func (this *Cursor) Add(i int) {
	this.value += i
}

func (this *Cursor) Increment() {
	this.value += 1
}
