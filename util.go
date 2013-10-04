/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package encoding

import (
	"fmt"
)

func FloorBy(value, factor uint32) uint32 {
	return value - value%factor
}

func CeilBy(value, factor uint32) uint32 {
	return value + factor - value%factor
}

// Copied from http://www.hackersdelight.org/hdcodetxt/nlz.c.txt - nlz2
func NumberOfLeadingZerosUint32_3(x uint32) uint32 {
	var n uint32 = 32
	var y uint32 = 0

	y = x >>16;  if (y != 0) {n = n -16;  x = y;}
	y = x >> 8;  if (y != 0) {n = n - 8;  x = y;}
	y = x >> 4;  if (y != 0) {n = n - 4;  x = y;}
	y = x >> 2;  if (y != 0) {n = n - 2;  x = y;}
	y = x >> 1;  if (y != 0) { return n - 2 }
	return n - x

}

// Copied from http://www.hackersdelight.org/hdcodetxt/nlz.c.txt - nlz1a
func NumberOfLeadingZerosUint32_2(x uint32) uint32 {
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

// Copied from http://www.hackersdelight.org/hdcodetxt/nlz.c.txt - nlz1
func NumberOfLeadingZerosUint32(x uint32) uint32 {
	var n uint32 = 0

	if (x == 0) { return 32 }

	if (x <= 0x0000FFFF) { n = n + 16; x = x<<16 }
	if (x <= 0x00FFFFFF) { n = n + 8; x = x<<8 }
	if (x <= 0x0FFFFFFF) { n = n + 4; x = x<<4 }
	if (x <= 0x3FFFFFFF) { n = n + 2; x = x<<2 }
	if (x <= 0x7FFFFFFF) { n = n + 1; }

	return n
}

func NumberOfBitsUint32(x uint32) uint32 {
	return 32 - NumberOfLeadingZerosUint32_2(x)
}

func MaxDiffBits(initoffset uint32, i []uint32, pos, length int) uint32 {
	return MaxDiffBits3(initoffset, i, pos, length)
}

func MaxDiffBits1(initoffset uint32, i []uint32, pos, length int) uint32 {
	var mask uint32
	mask |= (i[pos] - initoffset)

	for k := pos + 1; k < pos + length; k++ {
		mask |= i[k] - i[k - 1]
	}

	return NumberOfBitsUint32(mask)
}

func MaxDiffBits2(initoffset uint32, i []uint32, pos, length int) uint32 {
	var mask uint32

	last := i[pos]
	mask |= (last - initoffset)

	for k := pos + 1; k < pos + length; k++ {
		here := i[k]
		mask |= here - last
		last = here
	}

	return NumberOfBitsUint32(mask)
}

func MaxDiffBits3(initoffset uint32, i []uint32, pos, length int) uint32 {
	var mask uint32

	last := i[pos]
	mask |= (last - initoffset)

	for _, here := range i[pos+1:pos+length] {
		mask |= here - last
		last = here
	}

	return NumberOfBitsUint32(mask)
}

func MaxBits(i []uint32, pos, length int) uint32 {
	var mask uint32

	for _, v := range i[pos:pos+length] {
		mask |= v
	}

	return NumberOfBitsUint32(mask)
}

func PrintUint32sInBits(buf []uint32, pos, length int) {
	fmt.Println("                           10987654321098765432109876543210")
	for i := pos; i < pos + length; i++ {
		fmt.Printf("%4d: %20d %032b\n", i, uint32(buf[i]), uint32(buf[i]))
	}
}

