package mbytes

// Copyright(c) Dorin Duminica. All rights reserved.
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
//   1. Redistributions of source code must retain the above copyright notice,
// 	 this list of conditions and the following disclaimer.
//
//   2. Redistributions in binary form must reproduce the above copyright notice,
// 	 this list of conditions and the following disclaimer in the documentation
// 	 and/or other materials provided with the distribution.
//
//   3. Neither the name of the copyright holder nor the names of its
// 	 contributors may be used to endorse or promote products derived from this
// 	 software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// returned when the received whence value is unknown
var ErrWhenceUnknown = errors.New("Unknown whence value")

// returned when computed seek is negative
var ErrSeekNegative = errors.New("Negative seek")

// returned when seeking outside of the buffer
var ErrSeekOverflow = errors.New("Seek overflow")

// returned when offset is less than zero
var ErrOffsetNegative = errors.New("Negative offset")

// returned when offset is outside of the buffer
var ErrOffsetOverflow = errors.New("Offset overflow")

// returned when trying to read a byte from stream and the read size is different than byte size
var ErrByteRead = errors.New("Error reading byte")

// implemented interfaces
//	io.Seeker
//  io.Reader
//  io.ReaderAt
//  io.Writer
//  io.WriteAt
//	io.ByteReader
//	io.ByteWriter
type ByteBuffer struct {
	buff []byte
	pos  int
}

// create a new ByteBuffer with of (size) bytes
// NOTE:
//	- passing ZERO for size is allowed, the internal buffer grows on demand
func NewByteBuffer(size uint) *ByteBuffer {
	return (&ByteBuffer{}).Reset(size)
}

// creates a new internal buffer of size (size), position is reset to ZERO
// NOTE:
//	- any pre-existing data will be LOST
func (m *ByteBuffer) Reset(size uint) *ByteBuffer {
	if size > 0 {
		m.buff = make([]byte, size)
	} else {
		m.buff = []byte{}
	}
	m.pos = 0
	return m
}

// @ByteBuffer.Reset(0)
func (m *ByteBuffer) Clear() *ByteBuffer {
	return m.Reset(0)
}

// returns true if the size of internal buffer is ZERO
// you can also check if it's empty @ByteBuffer.Size() == 0
func (m *ByteBuffer) Empty() bool {
	return len(m.buff) == 0
}

// compare internal buffers of this and other
func (m *ByteBuffer) CmpWith(other *ByteBuffer) int {
	return bytes.Compare(m.buff, other.buff)
}

// returns a new clone of this
// position in the clone is set to ZERO
func (m *ByteBuffer) Clone() *ByteBuffer {
	r := NewByteBuffer(m.Size())
	r.pos = 0
	copy(r.buff, m.buff)
	return r
}

// returns size in bytes of internal buffer
func (m *ByteBuffer) Size() uint {
	return uint(len(m.buff))
}

// returns internal buffer position
func (m *ByteBuffer) Pos() int {
	return m.pos
}

// returns a copy of internal buffer as a byte slice
func (m *ByteBuffer) Bytes() []byte {
	r := make([]byte, len(m.buff))
	copy(r, m.buff)
	return r
}

// check if p is overflowing buffer
func (m *ByteBuffer) posOverflow(p int) bool {
	return p >= len(m.buff)
}

// @ByteBuffer.Seek(offset, io.SeekStart)
func (m *ByteBuffer) SeekFromStart(offset int64) (int64, error) {
	return m.Seek(offset, io.SeekStart)
}

// @ByteBuffer.Seek(offset, io.SeekCurrent)
func (m *ByteBuffer) SeekFromCurrent(offset int64) (int64, error) {
	return m.Seek(offset, io.SeekCurrent)
}

// @ByteBuffer.Seek(offset, io.SeekEnd)
func (m *ByteBuffer) SeekFromEnd(offset int64) (int64, error) {
	return m.Seek(offset, io.SeekEnd)
}

// @ByteBuffer.Seek(0, io.SeekStart)
func (m *ByteBuffer) SeekToStart() (int64, error) {
	return m.Seek(0, io.SeekStart)
}

// @ByteBuffer.Seek(0, io.SeekEnd)
func (m *ByteBuffer) SeekToEnd() (int64, error) {
	return m.Seek(0, io.SeekEnd)
}

// io.Seeker implementation
// returns offset position if err == nil
// errors:
//	ErrSeekNegative
//	ErrSeekOverflow
//	ErrWhenceUnknown
func (m *ByteBuffer) Seek(offset int64, whence int) (int64, error) {
	pos := int(offset)

	// validate whence
	switch whence {
	case io.SeekStart:
		// seeking from the beginning
	case io.SeekCurrent:
		// inc position by offset, offset can be both positive and negative
		pos += m.pos
	case io.SeekEnd:
		// set position to buffer length + offset, offset must be negative
		pos += len(m.buff)
	default:
		return -1, ErrWhenceUnknown
	}

	// sanity checks
	if pos < 0 {
		return -1, ErrSeekNegative
	}

	// check for overflow
	if m.posOverflow(pos) {
		return -1, ErrSeekOverflow
	}

	// update position
	m.pos = pos

	return int64(pos), nil
}

func (m *ByteBuffer) readFromPos(p []byte, pos int, incPos bool) (n int, err error) {
	l := len(p)

	// number of available bytes to read from current position
	avail := len(m.buff) - pos
	if avail > 0 {
		// read the minimum amount of bytes
		n = min_int(avail, l)

		// copy to buffer
		copy(p, m.buff[pos:])

		// increment position only if called by Read, ReadAt also calls this function
		if incPos {
			m.pos += n
		}

		// check if we've read less bytes than the size of p
		if n < l {
			// read less than the size of p, return io.EOF too
			err = io.EOF
		}
		return
	}
	return -1, io.EOF
}

// io.Reader implementation
// returns number of read bytes
// errors:
//	io.EOF
// NOTE:
//	- will return io.EOF error if the number of bytes read is less than the
//		size of p, however, p will contain the first n bytes from buffer
func (m *ByteBuffer) Read(p []byte) (n int, err error) {
	return m.readFromPos(p, m.pos, true)
}

// io.ReaderAt implementation
// reads up to len(p) from buffer at offset off
// returns number of read bytes
// errors:
//	io.EOF
//	ErrOffsetNegative
// NOTE:
//	- ReadAt will NOT modify internal position
//	- multiple readers may read at the same time, provided no write happens in between reads
func (m *ByteBuffer) ReadAt(p []byte, off int64) (n int, err error) {
	pos := int(off)

	// sanity checks
	if pos < 0 {
		return -1, ErrOffsetNegative
	}
	if m.posOverflow(pos) {
		return -1, ErrOffsetOverflow
	}

	return m.readFromPos(p, pos, false)
}

func (m *ByteBuffer) writeFromPos(p []byte, pos int) (appended int, written int, err error) {
	l := len(p)

	// number of overlap bytes
	noverlap := len(m.buff) - pos
	if noverlap > l {
		noverlap = l
	}

	// number of append bytes
	nappend := l - noverlap
	if noverlap > 0 {
		// override noverlap bytes
		copy(m.buff[pos:], p[:noverlap])
	}
	if nappend > 0 {
		// append nappend bytes
		m.buff = append(m.buff, p[noverlap:]...)
	}

	return nappend, l, nil
}

// io.Writer implementation
// writes p to internal buffer at current position
// NOTE:
//	- if current position is within the buffer, some or all of the bytes will be
//		overwritten
func (m *ByteBuffer) Write(p []byte) (n int, err error) {
	appended, written, err := m.writeFromPos(p, m.pos)
	if err != nil {
		return -1, err
	}
	m.pos += appended
	return written, err
}

// io.WriteAt implementation
// returns
//	ErrOffsetNegative
//	ErrOffsetOverflow
func (m *ByteBuffer) WriteAt(p []byte, off int64) (n int, err error) {
	// sanity checks
	if off < 0 {
		return -1, ErrOffsetNegative
	}
	if m.posOverflow(int(off)) {
		return -1, ErrOffsetOverflow
	}

	appended, written, err := m.writeFromPos(p, int(off))
	if err != nil {
		return -1, err
	}

	// in case of overwrite + append, we want to move the position to the last
	// appended byte in buffer
	m.pos += appended

	return written, err
}

// io.ByteReader implementation
func (m *ByteBuffer) ReadByte() (byte, error) {
	// read and return a byte from current position
	p := make([]byte, 1)
	n, err := m.Read(p)
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, ErrByteRead
	}
	return p[0], nil
}

// io.ByteWriter implementation
// NOTE: this function will never return an error, in case we're out of memory
// a panic will most likely occur
func (m *ByteBuffer) WriteByte(c byte) error {
	// append byte to buffer
	m.buff = append(m.buff, c)
	m.pos = len(m.buff)

	return nil
}

// returns a byte at a specific position in buffer
// much like indexing a byte slice
func (m *ByteBuffer) ByteAt(pos int) (byte, error) {
	p := make([]byte, 1)
	n, err := m.ReadAt(p, int64(pos))
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, ErrByteRead
	}
	return p[0], nil
}

// returns the number of bytes written or error
func (m *ByteBuffer) WriteUInt64Var(x uint64) (int, error) {
	buff := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buff, x)
	return m.Write(buff[:n])
}

// reads and returns an uint64s or error
func (m *ByteBuffer) ReadUInt64Var() (uint64, error) {
	return binary.ReadUvarint(m)
}
