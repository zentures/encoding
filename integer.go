/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package encoding

import (
	"github.com/dataence/encoding/cursor"
)

type Integer interface {
	// Compress data from an array to another array.
	//
	// Both inpos and outpos are modified to represent how much data was read and written to
	// if 12 ints (inlength = 12) are compressed to 3 ints, then inpos will be incremented by 12
	// while outpos will be incremented by 3 we use IntWrapper to pass the values by reference.
	// @param in  input array
	// @param inpos location in the input array
	// @param inlength how many integers to compress
	// @param out output array
	//* @param outpos  where to write in the output array
	Compress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error

	/**
	 * Uncompress data from an array to another array.
	 *
	 * Both inpos and outpos parameters are modified to indicate new positions after read/write.
	 *
	 * @param in array containing data in compressed form
	 * @param inpos where to start reading in the array
	 * @param inlength length of the compressed data (ignored by some schemes)
	 * @param out array where to write the compressed output
	 * @param outpos where to write the compressed output in out
	 */
	Uncompress(in []int32, inpos *cursor.Cursor, inlength int, out []int32, outpos *cursor.Cursor) error
}
