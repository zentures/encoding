/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package encoding

// This is just a list of functions that are no longer needed.

/*
// Copied from http://www.hackersdelight.org/hdcodetxt/nlz.c.txt - nlz2
func NumberOfLeadingZeros_3(x uint32) uint32 {
	var n uint32 = 32
	var y uint32 = 0

	y = x >>16;  if (y != 0) {n = n -16;  x = y;}
	y = x >> 8;  if (y != 0) {n = n - 8;  x = y;}
	y = x >> 4;  if (y != 0) {n = n - 4;  x = y;}
	y = x >> 2;  if (y != 0) {n = n - 2;  x = y;}
	y = x >> 1;  if (y != 0) { return n - 2 }
	return n - x
}
*/

/*
// Copied from http://www.hackersdelight.org/hdcodetxt/nlz.c.txt - nlz1
func NumberOfLeadingZeros(x uint32) uint32 {
	var n uint32 = 0

	if (x == 0) { return 32 }

	if (x <= 0x0000FFFF) { n = n + 16; x = x<<16 }
	if (x <= 0x00FFFFFF) { n = n + 8; x = x<<8 }
	if (x <= 0x0FFFFFFF) { n = n + 4; x = x<<4 }
	if (x <= 0x3FFFFFFF) { n = n + 2; x = x<<2 }
	if (x <= 0x7FFFFFFF) { n = n + 1; }

	return n
}
*/
