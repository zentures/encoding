/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package encoding

import (
	"fmt"
)

func FloorBy(value, factor int) int {
	return value - value%factor
}

func CeilBy(value, factor int) int {
	return value + factor - value%factor
}

func LeadingBitPosition(x uint32) int32 {
	return 32 - int32(nlz1a(x))
}

func DeltaMaxBits(initoffset int32, buf []int32) int32 {
	var mask int32

	for _, cur := range buf {
		mask |= cur - initoffset
		initoffset = cur
	}

	return LeadingBitPosition(uint32(mask))
}

func MaxBits(buf []int32) int32 {
	var mask int32

	for _, v := range buf {
		mask |= v
	}

	return LeadingBitPosition(uint32(mask))
}

func PrintInt32sInBits(buf []int32) {
	fmt.Println("                           10987654321098765432109876543210")
	//for i := pos; i < pos + length; i++ {
    for i, v := range buf {
		fmt.Printf("%4d: %20d %032b\n", i, v, uint32(v))
	}
}

func Delta(in, out []int32) {
	offset := int32(0)

	for i, v := range in {
		out[i] = v - offset
		offset = v
	}
}

func InverseDelta(in, out []int32) {
	offset := int32(0)

	for i, v := range in {
		out[i] = v + offset
		offset = out[i]
	}
}

// https://developers.google.com/protocol-buffers/docs/encoding#types
func ZigZagDelta(in, out []int32) {
	offset := int32(0)

	for i, v := range in {
		n := v - offset
		out[i] = (n << 1) ^ (n >> 31)
		offset = v
	}
}

func InverseZigZagDelta(in, out []int32) {
	offset := int32(0)

	for i, v := range in {
		//n := int32(uint32(v) >> 1) ^ (-(v & 1))
		n := int32(uint32(v) >> 1) ^ ((v << 31) >> 31)
		out[i] = n + offset
		offset = out[i]
	}
}

// Copied from http://www.hackersdelight.org/hdcodetxt/nlz.c.txt - nlz1a
func nlz1a(x uint32) uint32 {
	var n uint32 = 0
	if (x <= 0) { return (^x >> 26) & 32 }

	n = 1

	if ((x >> 16) == 0) { n = n +16; x = x <<16 }
	if ((x >> 24) == 0) { n = n + 8; x = x << 8 }
	if ((x >> 28) == 0) { n = n + 4; x = x << 4 }
	if ((x >> 30) == 0) { n = n + 2; x = x << 2 }
	n = n - (x >> 31)
	return n
}

