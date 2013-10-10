/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

// ByteBuffer is a Go implementation of the Java ByteBuffer
// http://docs.oracle.com/javase/7/docs/api/java/nio/ByteBuffer.html#put(int, byte)
//
// The documentation for each of the methods are copied from the above page as reference.


// A byte buffer.
// This class defines six categories of operations upon byte buffers:
//
// - Absolute and relative get and put methods that read and write single bytes;
// - Relative bulk get methods that transfer contiguous sequences of bytes from this buffer into an array;
// - Relative bulk put methods that transfer contiguous sequences of bytes from a byte array or some other
//   byte buffer into this buffer;
// - Absolute and relative get and put methods that read and write values of other primitive types, translating
//   them to and from sequences of bytes in a particular byte order;
// - Methods for creating view buffers, which allow a byte buffer to be viewed as a buffer containing values of
//   some other primitive type; and
// - Methods for compacting, duplicating, and slicing a byte buffer.
//
// Byte buffers can be created either by allocation, which allocates space for the buffer's content, or by
// wrapping an existing byte array into a buffer.
//
// Buffers are not safe for use by multiple concurrent threads. If a buffer is to be used by more than one
// thread then access to the buffer should be controlled by appropriate synchronization.
//
// Differences
// -----------
// This is not a full implementation of the Java ByteBuffer so please review the methods yourself.
// Here are a few immediate differences:
// - There is no direct vs indirect buffers in this implementation. All buffers are backed by []byte.
// - Most of the functions cannot be chained like the Java version
// - Some of the functions are not implemented, partly I don't have any use, partly coz I am lazy
//   - array(), arrayOffset()
//   - getChar(), putChar(), asCharBuffer()
//   - Any of the float or double functions
//   - get(byte[] dst), put(byte[] dst)
//   - isDirect()
//   - compact()
//   - hashCode()
//   - compareTo()
//   - view buffers are not implemented except for readonly buffer
// - Some of the functions are renamed to their Go equivalent
//   - Any Short became Uint16, e.g., getShort() -> GetUint16()
//   - Any Int became Uint32
//   - Any Long became Uint64
package buffers

import (
	"errors"
	"bytes"
	"fmt"
	"encoding/binary"
)

// A buffer is a linear, finite sequence of elements of a specific primitive type. The following invariant
// holds for the mark, position, limit, and capacity values:
// 		0 <= mark <= position <= limit <= capacity
// A newly-created buffer always has a position of zero and a mark that is undefined. The initial limit
// may be zero, or it may be some other value that depends upon the type of the buffer and the manner
// in which it is constructed. Each element of a newly-allocated buffer is initialized to zero.
type ByteBuffer struct {
	buf []byte

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

	// The new buffer will be backed by the given byte array; that is, modifications to the buffer
	// will cause the array to be modified and vice versa. The new buffer's capacity and limit will
	// be array.length, its position will be zero, and its mark will be undefined. Its backing array
	// will be the given array, and its array offset will be zero.
	wrapped bool

	// A read-only buffer does not allow its content to be changed, but its mark, position, and limit values
	// are mutable. Whether or not a buffer is read-only may be determined by invoking its isReadOnly method.
	readOnly bool

	// The byte order is used when reading or writing multibyte values, and when creating buffers that are
	// views of this byte buffer. The order of a newly-created byte buffer is always BigEndian.
	bo binary.ByteOrder
}

func NewByteBuffer(size int) *ByteBuffer {
	buf := make([]byte, size)
	b := NewWrappedBuffer(buf, size)
	b.wrapped = false
	return b
}

func NewWrappedBuffer(buf []byte, size int) *ByteBuffer {
	return &ByteBuffer{
		buf: buf,
		size: size,
		pos: 0,
		limit: size,
		mark: -1,
		wrapped: true,
		bo: binary.BigEndian,
	}
}

// Order retrieves this buffer's byte order.
// The byte order is used when reading or writing multibyte values, and when creating buffers that are
// views of this byte buffer. The order of a newly-created byte buffer is always binary.BigEndian.
func (this *ByteBuffer) Order() binary.ByteOrder {
	return this.bo
}

// SetOrder modifies this buffer's byte order.
func (this *ByteBuffer) SetOrder(bo binary.ByteOrder) {
	this.bo = bo
}

// Reset resets this buffer's position to the previously-marked position.
// Invoking this method neither changes nor discards the mark's value.
func (this *ByteBuffer) Reset() error {
	if this.mark == -1 {
		return errors.New("bytebuffer/Reset: Invalid Mark (has not been set)")
	}

	this.pos = this.mark

	return nil
}

// Clears this buffer. The position is set to zero, the limit is set to the capacity, and the mark is discarded.
func (this *ByteBuffer) Clear() error {
	if this.readOnly {
		return errors.New("bytebuffer/Resize: Cannot clear a read-only buffer")
	}

	this.pos = 0
	this.buf[this.pos] = 0
	this.mark = -1
	this.limit = this.size

	return nil
}

// Flips this buffer. The limit is set to the current position and then the position is set to zero.
// If the mark is defined then it is discarded.
func (this *ByteBuffer) Flip() error {
	this.limit = this.pos
	this.pos = 0
	return nil
}

// Rewinds this buffer. The position is set to zero and the mark is discarded.
func (this *ByteBuffer) Rewind() error {
	this.pos = 0
	this.mark = -1
	return nil
}

// Returns this buffer's capacity.
func (this *ByteBuffer) Capacity() int {
	return this.size
}

// Position returns this buffer's position.
func (this *ByteBuffer) Position() int {
	return this.pos
}

// SetPosition sets this buffer's position.
func (this *ByteBuffer) SetPosition(pos int) error {
	if pos > this.limit {
		return errors.New("bytebuffer/SetPosition: Position must not be greater than buffer limit")
	}

	if pos < 0 {
		return errors.New("bytebuffer/SetPosition: Position must be a non-negative number")

	}

	this.pos = pos
	return nil
}

// Mark returns this buffer's mark.
func (this *ByteBuffer) Mark() int {
	return this.mark
}

// Limit returns this buffer's position.
func (this *ByteBuffer) Limit() int {
	return this.limit
}

// SetLimit sets this buffer's limit. If the position is larger than the new limit then it is set to the
// new limit. If the mark is defined and larger than the new limit then it is discarded.
func (this *ByteBuffer) SetLimit(limit int) error {
	if this.limit > this.size {
		return errors.New("bytebuffer/SetLimit: Limit must not be greater than buffer size")
	}

	if this.limit < 0 {
		return errors.New("bytebuffer/SetLimit: Limit must be a non-negative number")

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
func (this *ByteBuffer) Remaining() int {
	return this.limit - this.pos
}

// HasRemaining tells whether there are any elements between the current position and the limit.
func (this *ByteBuffer) HasRemaining() bool {
	return this.limit - this.pos != 0
}

// IsReadOnly tells whether or not this buffer is read-only.
func (this *ByteBuffer) IsReadOnly() bool {
	return this.readOnly
}

// IsWrapped tells whether or not this buffer wraps another []byte
func (this *ByteBuffer) IsWrapped() bool {
	return this.wrapped
}

func (this *ByteBuffer) Resize(newSize int) error {
	if this.wrapped {
		return errors.New("bytebuffer/Resize: Cannot resize a wrapped buffer")
	}

	if this.readOnly {
		return errors.New("bytebuffer/Resize: Cannot resize a read-only buffer")
	}

	tmpBuf := this.buf
	this.buf = make([]byte, newSize)
	copy(this.buf, tmpBuf)

	this.size = newSize
	this.pos = 0

	return nil
}

func (this *ByteBuffer) Copy(b *ByteBuffer) error {
	if this.readOnly {
		return errors.New("bytebuffer/Resize: Cannot copy into a read-only buffer")
	}

	this.buf = make([]byte, b.Capacity())
	this.pos = b.Position()
	this.size = b.Capacity()
	copy(this.buf, b.buf)
	this.bo = b.bo

	return nil
}

// Slice creates a new byte buffer whose content is a shared subsequence of this buffer's content.
//
// The content of the new buffer will start at this buffer's current position. Changes to this buffer's
// content will be visible in the new buffer, and vice versa; the two buffers' position, limit, and mark
// values will be independent.
//
// The new buffer's position will be zero, its capacity and its limit will be the number of bytes remaining
// in this buffer, and its mark will be undefined. The new buffer will be read-only if, and only if,
// this buffer is read-only.
func (this *ByteBuffer) Slice() *ByteBuffer {
	return &ByteBuffer{
		buf: this.buf[this.pos:],
		pos: 0,
		size: this.size - this.pos,
		limit: this.limit - this.pos,
		mark: -1,
		wrapped: this.wrapped,
		readOnly: this.readOnly,
		bo: this.bo,
	}
}

// Duplicate creates a new byte buffer that shares this buffer's content.
//
// The content of the new buffer will be that of this buffer. Changes to this buffer's content will be visible
// in the new buffer, and vice versa; the two buffers' position, limit, and mark values will be independent.
//
// The new buffer's capacity, limit, position, and mark values will be identical to those of this buffer. The
// new buffer will be direct if, and only if, this buffer is direct, and it will be read-only if, and only if,
// this buffer is read-only.
func (this *ByteBuffer) Duplicate() *ByteBuffer {
	return &ByteBuffer{
		buf: this.buf,
		pos: this.pos,
		size: this.size,
		limit: this.limit,
		mark: this.mark,
		wrapped: this.wrapped,
		readOnly: this.readOnly,
		bo: this.bo,
	}
}

// Equals tells whether or not this buffer is equal to another buffer
func (this *ByteBuffer) Equals(other *ByteBuffer) bool {
	if this.size == other.size && bytes.Equal(this.buf, other.buf) {
		return true
	}

	return false
}

// AsReadOnlyBuffer creates a new, read-only byte buffer that shares this buffer's content.
// The content of the new buffer will be that of this buffer. Changes to this buffer's content will be
// visible in the new buffer; the new buffer itself, however, will be read-only and will not allow the
// shared content to be modified. The two buffers' position, limit, and mark values will be independent.
//
// The new buffer's capacity, limit, position, and mark values will be identical to those of this buffer.
//
// If this buffer is itself read-only then this method behaves in exactly the same way as the duplicate method.
func (this *ByteBuffer) AsReadOnlyBuffer() *ByteBuffer {
	b := this.Duplicate()
	b.readOnly = true
	return b
}

// AsUint32Buffer creates a view of this byte buffer as an uint32 buffer.
//
// The content of the new buffer will start at this buffer's current position. Changes to this
// buffer's content will be visible in the new buffer, and vice versa; the two buffers' position,
// limit, and mark values will be independent.
//
// The new buffer's position will be zero, its capacity and its limit will be the number of bytes
// remaining in this buffer divided by four, and its mark will be undefined. The new buffer will
// be direct if, and only if, this buffer is direct, and it will be read-only if, and only if,
// this buffer is read-only.
func (this *ByteBuffer) AsUint32Buffer() *Uint32Buffer {
	return &Uint32Buffer{
		buf: this.Slice(),
		size: this.Remaining()/4,
		limit: this.Remaining()/4,
		pos: 0,
		mark: -1,
		readOnly: this.readOnly,
	}
}

// AsInt32Buffer creates a view of this byte buffer as an uint32 buffer.
//
// The content of the new buffer will start at this buffer's current position. Changes to this
// buffer's content will be visible in the new buffer, and vice versa; the two buffers' position,
// limit, and mark values will be independent.
//
// The new buffer's position will be zero, its capacity and its limit will be the number of bytes
// remaining in this buffer divided by four, and its mark will be undefined. The new buffer will
// be direct if, and only if, this buffer is direct, and it will be read-only if, and only if,
// this buffer is read-only.
func (this *ByteBuffer) AsInt32Buffer() *Int32Buffer {
	return &Int32Buffer{
		buf: this.Slice(),
		size: this.Remaining()/4,
		limit: this.Remaining()/4,
		pos: 0,
		mark: -1,
		readOnly: this.readOnly,
	}
}

func (this *ByteBuffer) Peek() (byte, error) {
	if !this.HasRemaining() {
		return 0, errors.New("bytebuffer/Peek: No more remaining bytes")
	}

	return this.buf[this.pos], nil
}

// Get is a relative get method. Reads the byte at this buffer's current position, and then increments the position.
func (this *ByteBuffer) Get() (byte, error) {
	if !this.HasRemaining() {
		return 0, errors.New("bytebuffer/Get: No more remaining bytes")
	}

	this.pos += 1
	return this.buf[this.pos-1], nil
}

// GetAsUint32 is the same as Get() except it converts byte to uint32 in the return
// This is mainly a convenience method
func (this *ByteBuffer) GetAsUint32() (uint32, error) {
	if !this.HasRemaining() {
		return 0, errors.New("bytebuffer/Get: No more remaining bytes")
	}

	this.pos += 1
	return uint32(this.buf[this.pos-1]), nil
}

// GetAsInt32 is the same as Get() except it converts byte to int32 in the return
// This is mainly a convenience method
func (this *ByteBuffer) GetAsInt32() (int32, error) {
	if !this.HasRemaining() {
		return 0, errors.New("bytebuffer/Get: No more remaining bytes")
	}

	this.pos += 1
	return int32(this.buf[this.pos-1]), nil
}

// Put is a relative put method
// Writes the given byte into this buffer at the current position, and then increments the position.
func (this *ByteBuffer) Put(b byte) error {
	if !this.HasRemaining() {
		return errors.New("bytebuffer/Put: Byte buffer is full. Cannot put new byte.")
	}

	if this.IsReadOnly() {
		return errors.New("bytebuffer/Put: Byte buffer is read-only. Cannot put new byte.")
	}

	this.buf[this.pos] = b
	this.pos += 1

	//fmt.Printf("bytebuffer/Put(byte): %08b\n", b)
	return nil
}

// GetAt is an absolute get method. Reads the byte at the given index.
func (this *ByteBuffer) GetAt(index int) (byte, error) {
	if index < 0 || index + 1 > this.limit {
		return 0, errors.New("bytebuffer/GetAt: Index must be non-negative and not larger than the buffer limit.")
	}

	return this.buf[index], nil
}

// PutAt is an absolute put method that writes the given byte into this buffer at the given index.
func (this *ByteBuffer) PutAt(index int, b byte) error {
	if this.IsReadOnly() {
		return errors.New("bytebuffer/PutAt: Byte buffer is read-only. Cannot put new byte.")
	}

	if index < 0 || index + 1 > this.limit {
		return errors.New("bytebuffer/PutAt: Index must be non-negative and not larger than the buffer limit.")
	}

	this.buf[index] = b
	return nil
}

// GetBytes is a relative bulk get method.
//
// This method transfers bytes from this buffer into the given destination array. If there are fewer
// bytes remaining in the buffer than are required to satisfy the request, that is, if length > remaining(),
// then no bytes are transferred and a BufferUnderflowException is thrown.
//
// Otherwise, this method copies length bytes from this buffer into the given array, starting at the current
// position of this buffer and at the given offset in the array. The position of this buffer is then incremented
// by length.
//
// dst - The array into which bytes are to be written
// offset - The offset within the array of the first byte to be written; must be non-negative and no larger
// 		than dst.length
// length - The maximum number of bytes to be written to the given array; must be non-negative and no larger
// 		than dst.length - offset
func (this *ByteBuffer) GetBytes(dst []byte, offset, length int) error {
	if offset < 0 || offset > cap(dst) {
		return errors.New("bytebuffer/GetBytes: Offset must be non-negative and no larger than length of dst")
	}

	if length < 0 || length > cap(dst) - offset {
		return errors.New("bytebuffer/GetBytes: Length must be non-negative and no larger than length of dst - offset")
	}

	if length > this.Remaining() {
		return errors.New("bytebuffer/GetBytes: Insufficient bytes to get. Length is greater than remaining bytes.")
	}

	copy(dst[offset:], this.buf[this.pos:this.pos + length])
	this.pos += length

	return nil
}

// PutFrom is a relative bulk put method
//
// This method transfers the bytes remaining in the given source buffer into this buffer. If there are more
// bytes remaining in the source buffer than in this buffer, that is, if src.remaining() > remaining(), then
// no bytes are transferred and a BufferOverflowException is thrown.
//
// Otherwise, this method copies n = src.remaining() bytes from the given buffer into this buffer, starting
// at each buffer's current position. The positions of both buffers are then incremented by n.
func (this *ByteBuffer) PutFrom(src *ByteBuffer) error {
	if this.IsReadOnly() {
		return errors.New("bytebuffer/PutAt: Byte buffer is read-only. Cannot put new byte.")
	}

	if src.Remaining() > this.Remaining() {
		return errors.New("bytebuffer/PutFrom: Insufficient remaining space for copying")
	}

	n := src.Remaining()
	copy(this.buf[this.pos:this.pos + n], src.buf[src.pos:src.pos + n])
	this.pos += n
	src.pos += n

	return nil
}

// PutBytes is a relative bulk put method
//
// This method transfers bytes into this buffer from the given source array. If there are more bytes to be
// copied from the array than remain in this buffer, that is, if length > remaining(), then no bytes are
// transferred and a BufferOverflowException is thrown.
//
// Otherwise, this method copies length bytes from the given array into this buffer, starting at the given
// offset in the array and at the current position of this buffer. The position of this buffer is then
// incremented by length.
func (this *ByteBuffer) PutBytes(src []byte, offset, length int) error {
	if this.IsReadOnly() {
		return errors.New("bytebuffer/PutBytes: Byte buffer is read-only. Cannot put new byte.")
	}

	if offset < 0 || offset > cap(src) {
		return errors.New("bytebuffer/PutBytes: Offset must be non-negative and no larger than length of src")
	}

	if length < 0 || length > cap(src) - offset {
		return errors.New("bytebuffer/PutBytes: Length must be non-negative and no larger than length of src - offset")
	}

	if length > this.Remaining() {
		return errors.New("bytebuffer/PutBytes: Insufficient bytes to get. Length is greater than remaining bytes.")
	}

	copy(this.buf[this.pos:], src[offset:offset + length])
	this.pos += length
	return nil
}

// GetUint16 is a relative get method for reading a short value.
// Reads the next two bytes at this buffer's current position, composing them into a short value
// according to the current byte order, and then increments the position by two.
func (this *ByteBuffer) GetUint16() (uint16, error) {
	if this.Remaining() < 2 {
		return 0, errors.New("bytebuffer/Put: Insufficient remaining buffer for Uint16")
	}

	result := this.bo.Uint16(this.buf[this.pos:])
	this.pos += 2
	return result, nil
}

// GetUint16At is an absolute get method for reading a short value.
// Reads two bytes at the given index, composing them into a short value according to the current byte order.
func (this *ByteBuffer) GetUint16At(index int) (uint16, error) {
	if index < 0 || index + 2 > this.limit {
		return 0, errors.New("bytebuffer/GetUint16At: Index must be non-negative and not larger than the buffer limit.")
	}

	result := this.bo.Uint16(this.buf[index:])
	return result, nil
}

// GetUint16 is a relative get method for reading a short value.
// Reads the next two bytes at this buffer's current position, composing them into a short value
// according to the current byte order, and then increments the position by two.
func (this *ByteBuffer) PutUint16(value uint16) error {
	if this.IsReadOnly() {
		return errors.New("bytebuffer/PutUint16: Byte buffer is read-only. Cannot put new byte.")
	}

	if this.Remaining() < 2 {
		return errors.New("bytebuffer/PutUint16: Insufficient remaining space for putting uint16")
	}

	this.bo.PutUint16(this.buf[this.pos:], value)
	this.pos += 2
	return nil
}

// PutUint16At is an absolute put method for writing a short value
// Writes two bytes containing the given short value, in the current byte order, into this buffer at the given index.
func (this *ByteBuffer) PutUint16At(index int, value uint16) error {
	if this.IsReadOnly() {
		return errors.New("bytebuffer/PutUint16At: Byte buffer is read-only. Cannot put new byte.")
	}

	if index < 0 || index + 2 > this.limit {
		return errors.New("bytebuffer/PutUint16At: Index must be non-negative and not larger than the buffer limit.")
	}

	this.bo.PutUint16(this.buf[index:], value)
	return nil
}

// GetUint32 is a relative get method for reading a uint32 value.
// Reads the next four bytes at this buffer's current position, composing them into a uint32 value
// according to the current byte order, and then increments the position by two.
func (this *ByteBuffer) GetUint32() (uint32, error) {
	if this.Remaining() < 4 {
		return 0, errors.New("bytebuffer/GetUint32: Insufficient remaining buffer for Uint32")
	}

	//fmt.Printf("bytebuffer/GetUint32: next uint32 = %d, byte buffer = %v\n", binary.BigEndian.Uint32(this.buf[this.pos:]), this.buf[this.pos:])
	result := this.bo.Uint32(this.buf[this.pos:])
	//fmt.Printf("bytebuffer/GetUint32: uint32 = %d\n", result)
	this.pos += 4
	return result, nil
}

// GetUint16At is an absolute get method for reading a uint32 value.
// Reads four bytes at the given index, composing them into a uint32 value according to the current byte order.
func (this *ByteBuffer) GetUint32At(index int) (uint32, error) {
	if index < 0 || index + 4 > this.limit {
		return 0, errors.New("bytebuffer/GetUint32At: Index must be non-negative and not larger than the buffer limit.")
	}

	result := this.bo.Uint32(this.buf[index:])
	return result, nil
}

// GetUint32 is a relative get method for reading a uint32 value.
// Reads the next four bytes at this buffer's current position, composing them into a uint32 value
// according to the current byte order, and then increments the position by two.
func (this *ByteBuffer) PutUint32(value uint32) error {
	if this.IsReadOnly() {
		return errors.New("bytebuffer/PutUint32: Byte buffer is read-only. Cannot put new byte.")
	}

	if this.Remaining() < 4 {
		return errors.New("bytebuffer/PutUint32: Insufficient remaining space for putting uint32")
	}

	this.bo.PutUint32(this.buf[this.pos:], value)
	this.pos += 4
	return nil
}

// PutUint32At is an absolute put method for writing a uint32 value
// Writes four bytes containing the given uint32 value, in the current byte order, into this buffer at the given index.
func (this *ByteBuffer) PutUint32At(index int, value uint32) error {
	if this.IsReadOnly() {
		return errors.New("bytebuffer/PutUint32At: Byte buffer is read-only. Cannot put new byte.")
	}

	if index < 0 || index + 4 > this.limit {
		return errors.New("bytebuffer/PutUint32At: Index must be non-negative and not larger than the buffer limit.")
	}

	this.bo.PutUint32(this.buf[index:], value)
	return nil
}

// GetUint64 is a relative get method for reading a uint64 value.
// Reads the next eight bytes at this buffer's current position, composing them into a uint64 value
// according to the current byte order, and then increments the position by two.
func (this *ByteBuffer) GetUint64() (uint64, error) {
	if this.Remaining() < 8 {
		return 0, errors.New("bytebuffer/Put: Insufficient remaining buffer for Uint64")
	}

	result := this.bo.Uint64(this.buf[this.pos:])
	this.pos += 8
	return result, nil
}

// GetUint16At is an absolute get method for reading a uint64 value.
// Reads eight bytes at the given index, composing them into a uint64 value according to the current byte order.
func (this *ByteBuffer) GetUint64At(index int) (uint64, error) {
	if index < 0 || index + 8 > this.limit {
		return 0, errors.New("bytebuffer/GetUint64At: Index must be non-negative and not larger than the buffer limit.")
	}

	result := this.bo.Uint64(this.buf[index:])
	return result, nil
}

// GetUint64 is a relative get method for reading a uint64 value.
// Reads the next eight bytes at this buffer's current position, composing them into a uint64 value
// according to the current byte order, and then increments the position by two.
func (this *ByteBuffer) PutUint64(value uint64) error {
	if this.IsReadOnly() {
		return errors.New("bytebuffer/PutUint64: Byte buffer is read-only. Cannot put new byte.")
	}

	if this.Remaining() < 8 {
		return errors.New("bytebuffer/PutUint64: Insufficient remaining space for putting uint64")
	}

	this.bo.PutUint64(this.buf[this.pos:], value)
	this.pos += 8
	return nil
}

// PutUint64At is an absolute put method for writing a uint64 value
// Writes eight bytes containing the given uint64 value, in the current byte order, into this buffer at the given index.
func (this *ByteBuffer) PutUint64At(index int, value uint64) error {
	if this.IsReadOnly() {
		return errors.New("bytebuffer/PutUint64At: Byte buffer is read-only. Cannot put new byte.")
	}

	if index < 0 || index + 8 > this.limit {
		return errors.New("bytebuffer/PutUint64At: Index must be non-negative and not larger than the buffer limit.")
	}

	this.bo.PutUint64(this.buf[index:], value)
	return nil
}

func (this *ByteBuffer) String() string {
	return fmt.Sprintf("bytebuffer/String: Capacity = %d, limit = %d, mark = %d, position = %d\n", this.Capacity(), this.Limit(), this.Mark(), this.Position())
}
