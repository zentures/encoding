/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

// Uint32Buffer is a Go implementation of the Java IntBuffer
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

type Uint32Buffer struct {
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
func (this *Uint32Buffer) Order() binary.ByteOrder {
	return this.buf.Order()
}

// SetOrder modifies this buffer's byte order.
func (this *Uint32Buffer) SetOrder(bo binary.ByteOrder) {
	this.buf.SetOrder(bo)
}

// Reset resets this buffer's position to the previously-marked position.
// Invoking this method neither changes nor discards the mark's value.
func (this *Uint32Buffer) Reset() error {
	if this.mark == -1 {
		return errors.New("Uint32Buffer/Reset: Invalid Mark (has not been set)")
	}

	this.pos = this.mark

	return nil
}

// Clears this buffer. The position is set to zero, the limit is set to the capacity, and the mark is discarded.
func (this *Uint32Buffer) Clear() error {
	if this.readOnly {
		return errors.New("Uint32Buffer/Resize: Cannot clear a read-only buffer")
	}

	this.pos = 0
	this.mark = -1
	this.limit = this.size
	this.buf.Clear()

	return nil
}

// Flips this buffer. The limit is set to the current position and then the position is set to zero.
// If the mark is defined then it is discarded.
func (this *Uint32Buffer) Flip() error {
	this.limit = this.pos
	this.pos = 0
	return nil
}

// Rewinds this buffer. The position is set to zero and the mark is discarded.
func (this *Uint32Buffer) Rewind() error {
	this.pos = 0
	this.mark = -1
	return nil
}

// Returns this buffer's capacity.
func (this *Uint32Buffer) Capacity() int {
	return this.size
}

// Position returns this buffer's position.
func (this *Uint32Buffer) Position() int {
	return this.pos
}

// SetPosition sets this buffer's position.
func (this *Uint32Buffer) SetPosition(pos int) error {
	if pos > this.limit {
		return errors.New("Uint32Buffer/SetPosition: Position must not be greater than buffer limit")
	}

	if pos < 0 {
		return errors.New("Uint32Buffer/SetPosition: Position must be a non-negative number")

	}

	this.pos = pos
	return nil
}

// Mark returns this buffer's mark.
func (this *Uint32Buffer) Mark() int {
	return this.mark
}

// Limit returns this buffer's position.
func (this *Uint32Buffer) Limit() int {
	return this.limit
}

// SetLimit sets this buffer's limit. If the position is larger than the new limit then it is set to the
// new limit. If the mark is defined and larger than the new limit then it is discarded.
func (this *Uint32Buffer) SetLimit(limit int) error {
	if this.limit > this.size {
		return errors.New("Uint32Buffer/SetLimit: Limit must not be greater than buffer size")
	}

	if this.limit < 0 {
		return errors.New("Uint32Buffer/SetLimit: Limit must be a non-negative number")

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
func (this *Uint32Buffer) Remaining() int {
	return this.limit - this.pos
}

// HasRemaining tells whether there are any elements between the current position and the limit.
func (this *Uint32Buffer) HasRemaining() bool {
	return this.limit - this.pos != 0
}

// IsReadOnly tells whether or not this buffer is read-only.
func (this *Uint32Buffer) IsReadOnly() bool {
	return this.readOnly
}

// GetUint32 is a relative get method for reading a uint32 value.
// Reads the next four bytes at this buffer's current position, composing them into a uint32 value
// according to the current byte order, and then increments the position by two.
func (this *Uint32Buffer) Get() (uint32, error) {
	//fmt.Printf("uint32buffer/Get: remaining = %d\n", this.Remaining())

	if !this.HasRemaining() {
		return 0, errors.New("Uint32Buffer/Get: Insufficient remaining buffer for Uint32")
	}

	result, err := this.buf.GetUint32()
	if err == nil {
		this.pos += 1
	}
	return result, err
}

// GetUint16At is an absolute get method for reading a uint32 value.
// Reads four bytes at the given index, composing them into a uint32 value according to the current byte order.
func (this *Uint32Buffer) GetAt(index int) (uint32, error) {
	if index < 0 || index + 1 > this.limit {
		return 0, errors.New("Uint32Buffer/GetAt: Index must be non-negative and not larger than the buffer limit.")
	}

	return this.buf.GetUint32At(index*4)
}

// GetUint32s is a relative bulk get method.
//
// This method transfers ints from this buffer into the given destination array. If there are fewer
// ints remaining in the buffer than are required to satisfy the request, that is, if
// length > remaining(), then no ints are transferred and a BufferUnderflowException is thrown.
//
// Otherwise, this method copies length ints from this buffer into the given array, starting at the
// current position of this buffer and at the given offset in the array. The position of this buffer
// is then incremented by length.
func (this *Uint32Buffer) GetUint32s(dst []uint32, offset, length int) error {
	//fmt.Printf("uint32buffer/GetUint32s: length = %d, remaining = %d, offset = %d\n", length, this.Remaining(), offset)

	if offset < 0 || offset > cap(dst) {
		return errors.New("Uint32Buffer/GetUint32s: Offset must be non-negative and no larger than length of dst")
	}

	if length < 0 || length > cap(dst) - offset {
		//fmt.Printf("Uint32Buffer/GetUint32s: cap(dst)-offset = %d\n", cap(dst) - offset)
		//fmt.Printf("Uint32Buffer/GetUint32s: buf = %v\n", this.buf)
		return errors.New("Uint32Buffer/GetUint32s: Length must be non-negative and no larger than length of dst - offset ")
	}

	if length > this.Remaining() {
		return errors.New("Uint32Buffer/GetUint32s: Insufficient uint32s to get. Length is greater than remaining bytes.")
	}

	for i := offset; i < length + offset; i++ {
		var err error
		//fmt.Printf("Uint32Buffer/GetUint32s: i = %d\n", i)
		dst[i], err = this.Get()
		if err != nil {
			return errors.New("Uint32Buffer/GetUint32s: " + err.Error())
		}
	}

	return nil
}

// GetUint32 is a relative get method for reading a uint32 value.
// Reads the next four bytes at this buffer's current position, composing them into a uint32 value
// according to the current byte order, and then increments the position by two.
func (this *Uint32Buffer) Put(value uint32) error {
	if this.HasRemaining() {
		return errors.New("Uint32Buffer/PutUint32: Insufficient remaining space for putting uint32")
	}

	err := this.buf.PutUint32(value)
	if err != nil {
		this.pos += 1
		return nil
	}
	return err
}

// PutUint32At is an absolute put method for writing a uint32 value
// Writes four bytes containing the given uint32 value, in the current byte order, into this buffer at the given index.
func (this *Uint32Buffer) PutAt(index int, value uint32) error {
	if index < 0 || index + 1 > this.limit {
		return errors.New("Uint32Buffer/PutUint32At: Index must be non-negative and not larger than the buffer limit.")
	}

	return this.buf.PutUint32At(index*4, value)
}

// PutUint32s is a Relative bulk put method
// This method transfers ints into this buffer from the given source array. If there are more ints
// to be copied from the array than remain in this buffer, that is, if length > remaining(), then no
// ints are transferred and a BufferOverflowException is thrown.
//
// Otherwise, this method copies length ints from the given array into this buffer, starting at the g
// iven offset in the array and at the current position of this buffer. The position of this buffer
// is then incremented by length.
func (this *Uint32Buffer) PutUint32s(dst []uint32, offset, length int) error {
	if offset < 0 || offset > cap(dst) {
		return errors.New("Uint32Buffer/PutUint32s: Offset must be non-negative and no larger than length of dst")
	}

	if length < 0 || length > cap(dst) - offset {
		return errors.New("Uint32Buffer/PutUint32s: Length must be non-negative and no larger than length of dst - offset")
	}

	if length > this.Remaining() {
		return errors.New("Uint32Buffer/PutUint32s: Insufficient buffer size. Length is greater than remaining bytes.")
	}

	for i := offset; i < length + offset; i++ {
		err := this.Put(dst[i])
		if err != nil {
			return errors.New("Uint32Buffer/PutUint32s: " + err.Error())
		}
		this.pos += 1
	}

	return nil
}

func (this *Uint32Buffer) String() string {
	return fmt.Sprintf("Uint32Buffer/String: Capacity = %d, limit = %d, mark = %d, position = %d\n", this.Capacity(), this.Limit(), this.Mark(), this.Position())
}
