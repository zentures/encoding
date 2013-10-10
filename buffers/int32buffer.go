/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

// Int32Buffer is a Go implementation of the Java IntBuffer
// http://docs.oracle.com/javase/7/docs/api/java/nio/IntBuffer.html
// Please refer to the above link for method descriptions.
//
// This is an INCOMPLETE implementation. It's only implemented enough so it can act as a view buffer
// for ByteBuffer.
package buffers

import (
	"fmt"
	"errors"
	"encoding/binary"
)

type Int32Buffer struct {
	buf *ByteBuffer

	// A buffer's capacity is the number of elements it contains. The capacity of a buffer is never
	// negative and never changes.
	size int

	// A buffer's position is the index of the next element to be read or written. A buffer's position is
	// never negative and is never greater than its limit.
	pos int

	// A buffer's mark is the index to which its position will be reset when the reset method is invoked.
	// The mark is not always defined, but when it is defined it is never negative and is never greater
	// than the position. If the mark is defined then it is discarded when the position or the limit is
	// adjusted to a value smaller than the mark. If the mark is not defined then invoking the reset
	// method causes an InvalidMark error.
	mark int

	// A buffer's limit is the index of the first element that should not be read or written. A buffer's
	// limit is never negative and is never greater than its capacity.
	limit int

	// A read-only buffer does not allow its content to be changed, but its mark, position, and limit values
	// are mutable. Whether or not a buffer is read-only may be determined by invoking its isReadOnly method.
	readOnly bool
}

// Order retrieves this buffer's byte order.
// The byte order is used when reading or writing multibyte values, and when creating buffers that are
// views of this byte buffer. The order of a newly-created byte buffer is always binary.BigEndian.
func (this *Int32Buffer) Order() binary.ByteOrder {
	return this.buf.Order()
}

// SetOrder modifies this buffer's byte order.
func (this *Int32Buffer) SetOrder(bo binary.ByteOrder) {
	this.buf.SetOrder(bo)
}

// Reset resets this buffer's position to the previously-marked position.
// Invoking this method neither changes nor discards the mark's value.
func (this *Int32Buffer) Reset() error {
	if this.mark == -1 {
		return errors.New("int32buffer/Reset: Invalid Mark (has not been set)")
	}

	this.pos = this.mark

	return nil
}

// Clears this buffer. The position is set to zero, the limit is set to the capacity, and the mark is discarded.
func (this *Int32Buffer) Clear() error {
	if this.readOnly {
		return errors.New("int32buffer/Resize: Cannot clear a read-only buffer")
	}

	this.pos = 0
	this.mark = -1
	this.limit = this.size
	this.buf.Clear()

	return nil
}

// Flips this buffer. The limit is set to the current position and then the position is set to zero.
// If the mark is defined then it is discarded.
func (this *Int32Buffer) Flip() error {
	this.limit = this.pos
	this.pos = 0
	return nil
}

// Rewinds this buffer. The position is set to zero and the mark is discarded.
func (this *Int32Buffer) Rewind() error {
	this.pos = 0
	this.mark = -1
	return nil
}

// Returns this buffer's capacity.
func (this *Int32Buffer) Capacity() int {
	return this.size
}

// Position returns this buffer's position.
func (this *Int32Buffer) Position() int {
	return this.pos
}

// SetPosition sets this buffer's position.
func (this *Int32Buffer) SetPosition(pos int) error {
	if pos > this.limit {
		return errors.New("int32buffer/SetPosition: Position must not be greater than buffer limit")
	}

	if pos < 0 {
		return errors.New("int32buffer/SetPosition: Position must be a non-negative number")

	}

	this.pos = pos
	return nil
}

// Mark returns this buffer's mark.
func (this *Int32Buffer) Mark() int {
	return this.mark
}

// Limit returns this buffer's position.
func (this *Int32Buffer) Limit() int {
	return this.limit
}

// SetLimit sets this buffer's limit. If the position is larger than the new limit then it is set to the
// new limit. If the mark is defined and larger than the new limit then it is discarded.
func (this *Int32Buffer) SetLimit(limit int) error {
	if this.limit > this.size {
		return errors.New("int32buffer/SetLimit: Limit must not be greater than buffer size")
	}

	if this.limit < 0 {
		return errors.New("int32buffer/SetLimit: Limit must be a non-negative number")

	}

	this.limit = limit

	if this.pos > this.limit {
		this.pos = this.limit
	}

	if this.mark > this.limit {
		this.mark = -1
	}

	return nil
}

// Remaining returns the number of elements between the current position and the capacity
func (this *Int32Buffer) Remaining() int {
	return this.limit - this.pos
}

// HasRemaining tells whether there are any elements between the current position and the limit.
func (this *Int32Buffer) HasRemaining() bool {
	return this.limit - this.pos != 0
}

// IsReadOnly tells whether or not this buffer is read-only.
func (this *Int32Buffer) IsReadOnly() bool {
	return this.readOnly
}

// GetUint32 is a relative get method for reading a int32 value.
// Reads the next four bytes at this buffer's current position, composing them into a int32 value
// according to the current byte order, and then increments the position by two.
func (this *Int32Buffer) Get() (int32, error) {
	//fmt.Printf("int32buffer/Get: remaining = %d\n", this.Remaining())

	if !this.HasRemaining() {
		return 0, errors.New("int32buffer/Get: Insufficient remaining buffer for Uint32")
	}

	result, err := this.buf.GetUint32At(this.pos*4)
	if err == nil {
		this.pos += 1
	}
	return int32(result), err
}

// GetUint16At is an absolute get method for reading a int32 value.
// Reads four bytes at the given index, composing them into a int32 value according to the current byte order.
func (this *Int32Buffer) GetAt(index int) (int32, error) {
	if index < 0 || index + 1 > this.limit {
		return 0, errors.New("int32buffer/GetAt: Index must be non-negative and not larger than the buffer limit.")
	}

	result, err := this.buf.GetUint32At(index*4)
	return int32(result), err
}

// GetInt32s is a relative bulk get method.
//
// This method transfers ints from this buffer into the given destination array. If there are fewer
// ints remaining in the buffer than are required to satisfy the request, that is, if
// length > remaining(), then no ints are transferred and a BufferUnderflowException is thrown.
//
// Otherwise, this method copies length ints from this buffer into the given array, starting at the
// current position of this buffer and at the given offset in the array. The position of this buffer
// is then incremented by length.
func (this *Int32Buffer) GetInt32s(dst []int32, offset, length int) error {
	//fmt.Printf("int32buffer/GetInt32s: length = %d, remaining = %d, offset = %d\n", length, this.Remaining(), offset)

	if offset < 0 || offset > cap(dst) {
		return errors.New("int32buffer/GetInt32s: Offset must be non-negative and no larger than length of dst")
	}

	if length < 0 || length > cap(dst) - offset {
		//fmt.Printf("int32buffer/GetInt32s: cap(dst)-offset = %d\n", cap(dst) - offset)
		//fmt.Printf("int32buffer/GetInt32s: buf = %v\n", this.buf)
		return errors.New("int32buffer/GetInt32s: Length must be non-negative and no larger than length of dst - offset ")
	}

	if length > this.Remaining() {
		return errors.New("int32buffer/GetInt32s: Insufficient int32s to get. Length is greater than remaining bytes.")
	}

	for i := offset; i < length + offset; i++ {
		//fmt.Printf("int32buffer/GetInt32s: i = %d\n", i)
		res, err := this.Get()
		if err != nil {
			return errors.New("int32buffer/GetInt32s: " + err.Error())
		}
		dst[i] = int32(res)
	}

	return nil
}

// GetUint32 is a relative get method for reading a int32 value.
// Reads the next four bytes at this buffer's current position, composing them into a int32 value
// according to the current byte order, and then increments the position by two.
func (this *Int32Buffer) Put(value int32) error {
	if !this.HasRemaining() {
		return fmt.Errorf("int32buffer/Put: Insufficient remaining space (%t) for putting int32", this.HasRemaining())
	}

	if err := this.buf.PutUint32At(this.pos*4, uint32(value)); err != nil {
		return errors.New("int32buffer/Put: " + err.Error())
	}

	this.pos += 1
	return nil
}

// PutUint32At is an absolute put method for writing a int32 value
// Writes four bytes containing the given int32 value, in the current byte order, into this buffer at the given index.
func (this *Int32Buffer) PutAt(index int, value int32) error {
	if index < 0 || index + 1 > this.limit {
		return errors.New("int32buffer/PutAt: Index must be non-negative and not larger than the buffer limit.")
	}

	return this.buf.PutUint32At(index*4, uint32(value))
}

// PutInt32s is a Relative bulk put method
// This method transfers ints into this buffer from the given source array. If there are more ints
// to be copied from the array than remain in this buffer, that is, if length > remaining(), then no
// ints are transferred and a BufferOverflowException is thrown.
//
// Otherwise, this method copies length ints from the given array into this buffer, starting at the g
// iven offset in the array and at the current position of this buffer. The position of this buffer
// is then incremented by length.
func (this *Int32Buffer) PutInt32s(dst []int32, offset, length int) error {
	if offset < 0 || offset > cap(dst) {
		return fmt.Errorf("int32buffer/PutInt32s: Offset (%d) must be non-negative and no larger than length of dst", offset)
	}

	if length < 0 || length > cap(dst) - offset {
		return fmt.Errorf("int32buffer/PutInt32s: Length (%d) must be non-negative and no larger than length of dst - offset (%d)", length, cap(dst) - offset)
	}

	if length > this.Remaining() {
		return fmt.Errorf("int32buffer/PutInt32s: Insufficient buffer size. Length (%d) is greater than remaining buffer (%d).", length, this.Remaining())
	}

	for i := offset; i < length + offset; i++ {
		if err := this.Put(dst[i]); err != nil {
			return errors.New("int32buffer/PutInt32s: " + err.Error())
		}
	}

	return nil
}

func (this *Int32Buffer) String() string {
	return fmt.Sprintf("int32buffer/String: Capacity = %d, limit = %d, mark = %d, position = %d\n", this.Capacity(), this.Limit(), this.Mark(), this.Position())
}
