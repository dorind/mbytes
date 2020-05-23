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
	"io"
	"testing"
)

func errOrStr(err error, s string) string {
	if err != nil {
		return err.Error()
	}
	return s
}

func errOrNilStr(err error) string {
	return errOrStr(err, "<NIL>")
}

func TestNewByteBuffer(t *testing.T) {
	tag := "NewByteBuffer()"

	test_sizes := []uint{0, 1, 3, 5, 1024}

	for i := 0; i < len(test_sizes); i++ {
		expected_size := test_sizes[i]
		b := NewByteBuffer(expected_size)
		size := b.Size()
		if size != expected_size {
			t.Fatalf(tag+" size error, expected %v, found %v", expected_size, size)
		}
	}
}

func TestByteBufferReset(t *testing.T) {
	tag := "ByteBuffer.Reset()"

	test_sizes := []uint{0, 1, 3, 5, 1024}
	for i := 0; i < len(test_sizes); i++ {
		b := NewByteBuffer(test_sizes[i])
		b.Reset(0)
		size := b.Size()
		if size != 0 {
			t.Fatalf(tag+" size error, expected 0, found %v", size)
		}
		p := b.Pos()
		if p != 0 {
			t.Fatalf(tag+" pos error, expected 0, found %v", p)
		}
	}
}

func TestByteBufferCmpWith(t *testing.T) {
	tag := "ByteBuffer.CmpWith()"

	left := NewByteBuffer(64)
	right := NewByteBuffer(65)
	rcmp := left.CmpWith(right)
	if rcmp != -1 {
		t.Fatalf(tag+" error, expected 1, found %v", rcmp)
	}

	right.Reset(64)
	rcmp = left.CmpWith(right)
	if rcmp != 0 {
		t.Fatalf(tag+" error, expected 0, found %v", rcmp)
	}

	left.Reset(65)
	rcmp = left.CmpWith(right)
	if rcmp != 1 {
		t.Fatalf(tag+" error, expected -1, found %v", rcmp)
	}
}

func TestByteBufferClone(t *testing.T) {
	tag := "ByteBuffer.Clone()"

	original := NewByteBuffer(64)
	clone := original.Clone()

	rcmp := original.CmpWith(clone)
	if rcmp != 0 {
		t.Fatalf(tag+" error, cmp result %v", rcmp)
	}
}

func TestByteBufferBytes(t *testing.T) {
	tag := "ByteBuffer.Bytes()"

	s := "abracadabra"
	b := NewByteBuffer(0)
	n, err := b.Write([]byte(s))
	if err != nil {
		t.Fatalf(tag+" unexpected error: %v", err.Error())
	}
	if n != len(s) {
		t.Fatalf(tag+" unexpected write size, expected %v, found %v", len(s), n)
	}
	buf := b.Bytes()
	if string(buf) != s {
		t.Fatalf(tag+" bytes mismatch, expected [%v], found [%v]", []byte(s), buf)
	}
}

func TestByteBufferSeekStart(t *testing.T) {
	tag := "ByteBuffer.Seek(start)"

	b := NewByteBuffer(64)
	p := int64(2)
	pos, err := b.SeekFromStart(p)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %s", err.Error())
	}
	bpos := b.Pos()
	if int64(bpos) != p {
		t.Fatalf(tag+" pos error, expected %v, found %v", p, bpos)
	}
	if pos != int64(bpos) {
		t.Fatalf(tag+" seek pos(%v) mismatches Pos(%v)", pos, bpos)
	}
}

func TestByteBufferSeekStartNegative(t *testing.T) {
	tag := "ByteBuffer.Seek(start-negative)"

	b := NewByteBuffer(64)
	p := int64(-1)
	_, err := b.SeekFromStart(p)
	if err != ErrSeekNegative {
		t.Fatalf(tag+" expected error [%v], found [%v]", ErrSeekNegative.Error(), errOrNilStr(err))
	}
	// ByteBuffer.Pos() should NOT be affected
	bpos := b.Pos()
	if bpos != 0 {
		t.Fatalf(tag+" pos error, expected 0, found %v", bpos)
	}
}

func TestByteBufferSeekStartOverflow(t *testing.T) {
	tag := "ByteBuffer.Seek(start-overflow)"

	b := NewByteBuffer(64)
	p := b.Size() + 1
	_, err := b.SeekFromStart(int64(p))
	if err != ErrSeekOverflow {
		t.Fatalf(tag+" expected error [%v], found [%v]", ErrSeekOverflow.Error(), errOrNilStr(err))
	}
	// ByteBuffer.Pos() should NOT be affected
	bpos := b.Pos()
	if bpos != 0 {
		t.Fatalf(tag+" pos error, expected 0, found %v", bpos)
	}
}

func TestByteBufferSeekCurrent(t *testing.T) {
	tag := "ByteBuffer.Seek(current)"
	b := NewByteBuffer(64)
	p := int64(32)
	pos, err := b.SeekFromCurrent(p)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %s", err.Error())
	}
	if pos != p {
		t.Fatalf(tag+" unexpected seek pos, expected %v, found %v", p, pos)
	}
	bpos := b.Pos()
	if int64(bpos) != p {
		t.Fatalf(tag+" pos error, expected %v, found %v", p, bpos)
	}
	if pos != int64(bpos) {
		t.Fatalf(tag+" seek pos(%v) mismatches Pos(%v)", pos, bpos)
	}
}

func TestByteBufferSeekCurrentNegative(t *testing.T) {
	tag := "ByteBuffer.Seek(current-negative)"

	b := NewByteBuffer(64)
	ip := int64(32)
	p := ip
	_, err := b.SeekFromCurrent(p)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %s", err.Error())
	}
	p = int64(-33)
	_, err = b.SeekFromCurrent(p)
	if err != ErrSeekNegative {
		t.Fatalf(tag+" expected error [%v], found [%v]", ErrSeekNegative.Error(), errOrNilStr(err))
	}
	// ByteBuffer.Pos() should NOT be affected
	bpos := b.Pos()
	if int64(bpos) != ip {
		t.Fatalf(tag+" pos error, expected %v, found %v", ip, bpos)
	}
}

func TestByteBufferSeekCurrentOverflow(t *testing.T) {
	tag := "ByteBuffer.Seek(current-overflow)"

	b := NewByteBuffer(64)
	p := int64(24)
	_, err := b.SeekFromCurrent(p)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %s", err.Error())
	}
	p += 45
	ppos := b.Pos()
	_, err = b.SeekFromCurrent(int64(p))
	if err != ErrSeekOverflow {
		t.Fatalf(tag+" expected error [%v], found [%v]", ErrSeekOverflow.Error(), errOrNilStr(err))
	}
	// ByteBuffer.Pos() should NOT be affected
	bpos := b.Pos()
	if bpos != ppos {
		t.Fatalf(tag+" pos error, expected 0, found %v", bpos)
	}
}

func TestByteBufferSeekEnd(t *testing.T) {
	tag := "ByteBuffer.Seek(end)"

	size := uint(64)
	p := int64(-8)
	b := NewByteBuffer(size)
	pos, err := b.SeekFromEnd(p)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %s", err.Error())
	}
	bpos := b.Pos()
	if int64(bpos) != pos {
		t.Fatalf(tag+" pos error, expected %v, found %v", p, bpos)
	}
	epos := int64(size) + p
	if int64(bpos) != epos {
		t.Fatalf(tag+" seek expected pos(%v) mismatches Pos(%v)", epos, bpos)
	}
}

func TestByteBufferSeekEndNegative(t *testing.T) {
	tag := "ByteBuffer.Seek(end-negative)"

	size := uint(64)
	b := NewByteBuffer(size)
	p := -int64(size + 1)
	_, err := b.SeekFromEnd(p)
	if err != ErrSeekNegative {
		t.Fatalf(tag+" expected error [%v], found [%v]", ErrSeekNegative.Error(), errOrNilStr(err))
	}
	// ByteBuffer.Pos() should NOT be affected
	bpos := b.Pos()
	if bpos != 0 {
		t.Fatalf(tag+" pos error, expected 0, found %v", bpos)
	}
}

func TestByteBufferSeekEndOverflow(t *testing.T) {
	tag := "ByteBuffer.Seek(end-overflow)"

	b := NewByteBuffer(64)
	p := 1
	_, err := b.SeekFromEnd(int64(p))
	if err != ErrSeekOverflow {
		t.Fatalf(tag+" expected error [%v], found [%v]", ErrSeekOverflow.Error(), errOrNilStr(err))
	}
	// ByteBuffer.Pos() should NOT be affected
	bpos := b.Pos()
	if bpos != 0 {
		t.Fatalf(tag+" pos error, expected 0, found %v", bpos)
	}
}

func TestByteBufferRead(t *testing.T) {
	tag := "ByteBuffer.Read()"

	b := NewByteBuffer(64)
	size := 32
	buff := make([]byte, size)
	n, err := b.Read(buff)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %v", err.Error())
	}
	if n != size {
		t.Fatalf(tag+" unexpected read size, expected %v, found %v", size, n)
	}
	n, err = b.Read(buff)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %v", err.Error())
	}
	if n != size {
		t.Fatalf(tag+" unexpected read size, expected %v, found %v", size, n)
	}
}

func TestByteBufferReadEmpty(t *testing.T) {
	tag := "ByteBuffer.Read(Empty)"

	b := NewByteBuffer(0)
	buff := make([]byte, 1)
	n, err := b.Read(buff)
	if err != io.EOF {
		t.Fatalf(tag+" unexpected error, expected [%v], found [%v]", io.EOF.Error(), errOrNilStr(err))
	}
	if n > 0 {
		t.Fatalf(tag+" unexpected read bytes in return, expected < 1, found %v", n)
	}
}

func TestByteBufferReadEOF(t *testing.T) {
	tag := "ByteBuffer.Read(EOF)"

	b := NewByteBuffer(48)
	size := 32
	buff := make([]byte, size)
	n, err := b.Read(buff)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %v", err.Error())
	}
	if n != size {
		t.Fatalf(tag+" unexpected read size, expected %v, found %v", size, n)
	}
	n, err = b.Read(buff)
	if err != io.EOF {
		t.Fatalf(tag+" unexpected error, expected [%v], found [%v]", io.EOF.Error(), errOrNilStr(err))
	}
	if n != size*2-int(b.Size()) {
		t.Fatalf(tag+" unexpected read size, expected %v, found %v", size, n)
	}
}

func TestByteBufferReadAt(t *testing.T) {
	tag := "ByteBuffer.ReadAt()"

	b := NewByteBuffer(64)
	size := 32
	buff := make([]byte, size)
	n, err := b.ReadAt(buff, 16)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %v", err.Error())
	}
	if n != size {
		t.Fatalf(tag+" unexpected read size, expected %v, found %v", size, n)
	}
	pos := b.Pos()
	if pos != 0 {
		t.Fatalf(tag+" unexpected position, expected 0, found %v", pos)
	}
	n, err = b.ReadAt(buff, 32)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %v", err.Error())
	}
	if n != size {
		t.Fatalf(tag+" unexpected read size, expected %v, found %v", size, n)
	}
	pos = b.Pos()
	if pos != 0 {
		t.Fatalf(tag+" unexpected position, expected 0, found %v", pos)
	}
}

func TestByteBufferReadAtEmpty(t *testing.T) {
	tag := "ByteBuffer.ReadAt(Empty)"

	b := NewByteBuffer(0)
	size := 2
	buff := make([]byte, size)
	n, err := b.ReadAt(buff, 16)
	if err != ErrOffsetOverflow {
		t.Fatalf(tag+" unexpected error, expected [%v], found [%v]", ErrOffsetOverflow.Error(), errOrNilStr(err))
	}
	if n > 0 {
		t.Fatalf(tag+" unexpected read size, expected <1, found %v", n)
	}
	pos := b.Pos()
	if pos != 0 {
		t.Fatalf(tag+" unexpected position, expected 0, found %v", pos)
	}
}

func TestByteBufferReadAtEOF(t *testing.T) {
	tag := "ByteBuffer.ReadAt(EOF)"

	b := NewByteBuffer(64)
	size := 32
	buff := make([]byte, size)
	n, err := b.ReadAt(buff, 16)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %v", err.Error())
	}
	if n != size {
		t.Fatalf(tag+" unexpected read size, expected %v, found %v", size, n)
	}
	pos := b.Pos()
	if pos != 0 {
		t.Fatalf(tag+" unexpected position, expected 0, found %v", pos)
	}
	n, err = b.ReadAt(buff, 48)
	if err != io.EOF {
		t.Fatalf(tag+" unexpected error, expected [%v], found [%v]", io.EOF.Error(), errOrNilStr(err))
	}
	pos = b.Pos()
	if pos != 0 {
		t.Fatalf(tag+" unexpected position, expected 0, found %v", pos)
	}
}

func TestByteBufferWrite(t *testing.T) {
	tag := "ByteBuffer.Write()"

	b := NewByteBuffer(0)
	buff := []byte{'a', 'b', 'c', 'd', 'e', 'f'}
	rbuff := make([]byte, len(buff))

	// write buffer and check
	n, err := b.Write(buff)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %v", err.Error())
	}
	if n != len(buff) {
		t.Fatalf(tag+" write size error, expected %v, found %v", len(buff), n)
	}
	pos := b.Pos()
	if pos != len(buff) {
		t.Fatalf(tag+" unexpected position, expected %v, found %v", len(buff), pos)
	}

	// validate what we wrote
	s, err := b.SeekFromStart(0)
	if err != nil {
		t.Fatalf(tag+" unexpected seek error: %v", err.Error())
	}
	if s != 0 {
		t.Fatalf(tag+" seek position error, expected 0, found %v", s)
	}
	n, err = b.Read(rbuff)
	if err != nil && err != io.EOF {
		t.Fatalf(tag+" unexpected read error %v", err.Error())
	}
	if n != len(rbuff) {
		t.Fatalf(tag+" unexpected read size, expected %v, found %v", len(rbuff), n)
	}
	if bytes.Compare(buff, rbuff) != 0 {
		t.Fatalf(tag+" write validation failed, wrote [%v], read [%v]", buff, rbuff)
	}

	// write append
	n, err = b.Write(buff)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %v", err.Error())
	}
	if n != len(buff) {
		t.Fatalf(tag+" write size error, expected %v, found %v", len(buff), n)
	}
	pos = b.Pos()
	if pos != len(buff)*2 {
		t.Fatalf(tag+" unexpected position, expected %v, found %v", len(buff)*2, pos)
	}

	// validate what we wrote
	s, err = b.SeekFromCurrent(-int64(len(buff)))
	if err != nil {
		t.Fatalf(tag+" unexpected seek error: %v", err.Error())
	}
	es := int64(b.Size() - uint(len(buff)))
	if s != es {
		t.Fatalf(tag+" seek position error, expected %v, found %v", es, s)
	}
	b.Read(rbuff)
	if bytes.Compare(buff, rbuff) != 0 {
		t.Fatalf(tag+" write validation failed, wrote [%v], read [%v]", buff, rbuff)
	}

	// write with overlap
	// first half of buff will overwrite last few bytes
	// second half of buff will be appened
	//   initial  -> a b c d e f a b c d e f
	// overlapped -> a b c d e f a b c a b c d e f
	// marked     -> a b c d e f a b c A B C d e f
	soff := int64(len(buff) / 2)
	s, err = b.SeekFromCurrent(-soff)
	if err != nil {
		t.Fatalf(tag+" unexpected seek error: %v", err.Error())
	}
	es = int64(b.Size()) - soff
	if s != es {
		t.Fatalf(tag+" seek position error, expected %v, found %v", es, s)
	}
	n, err = b.Write(buff)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %v", err.Error())
	}
	if n != len(buff) {
		t.Fatalf(tag+" write size error, expected %v, found %v", len(buff), n)
	}

	// validate what we wrote
	s, err = b.SeekFromEnd(-int64(len(buff)))
	if err != nil {
		t.Fatalf(tag+" unexpected seek error: %v", err.Error())
	}
	es = int64(b.Size() - uint(len(buff)))
	if s != es {
		t.Fatalf(tag+" seek position error, expected %v, found %v", es, s)
	}
	b.Read(rbuff)
	if bytes.Compare(buff, rbuff) != 0 {
		t.Fatalf(tag+" write validation failed, wrote [%v], read [%v]", buff, rbuff)
	}

	//
	// write overlap only
	//
	s, err = b.SeekFromEnd(-int64(len(buff)))
	if err != nil {
		t.Fatalf(tag+" unexpected seek error: %v", err.Error())
	}
	es = int64(b.Size() - uint(len(buff)))
	if s != es {
		t.Fatalf(tag+" seek position error, expected %v, found %v", es, s)
	}
	n, err = b.Write(buff)
	if err != nil {
		t.Fatalf(tag+" unexpected error: %v", err.Error())
	}
	if n != len(buff) {
		t.Fatalf(tag+" write size error, expected %v, found %v", len(buff), n)
	}

	// validate write overlap
	s, err = b.SeekFromEnd(-int64(len(buff)))
	if err != nil {
		t.Fatalf(tag+" unexpected seek error: %v", err.Error())
	}
	es = int64(b.Size() - uint(len(buff)))
	if s != es {
		t.Fatalf(tag+" seek position error, expected %v, found %v", es, s)
	}
	b.Read(rbuff)
	if bytes.Compare(buff, rbuff) != 0 {
		t.Fatalf(tag+" write validation failed, wrote [%v], read [%v]", buff, rbuff)
	}
}

func TestByteBufferWriteAt(t *testing.T) {
	tag := "ByteBuffer.WriteAt()"

	b := NewByteBuffer(0)
	buff := []byte{'a', 'b', 'c', 'd', 'e', 'f'}

	buff_times := 5

	// build buffer
	for i := 1; i <= buff_times; i++ {
		n, err := b.Write(buff)
		if err != nil {
			t.Fatalf(tag+" unexpected error: %v", err.Error())
		}
		if n != len(buff) {
			t.Fatalf(tag+" write size error, expected %v, found %v", len(buff), n)
		}
	}

	clone := b.Clone()

	bs := b.Size()
	es := uint(buff_times * len(buff))
	if bs != es {
		t.Fatalf(tag+" unexpected buffer size, expected %v, found %v", es, bs)
	}

	// overwrite data entirely
	for i := uint(0); i < es; i += uint(len(buff)) {
		n, err := b.WriteAt(buff, int64(i))
		if err != nil {
			t.Fatalf(tag+" unexpected error: %v", err.Error())
		}
		if n != len(buff) {
			t.Fatalf(tag+" write size error, expected %v, found %v", len(buff), n)
		}
	}

	// validate data
	if clone.CmpWith(b) != 0 {
		t.Fatalf(tag+" data mismatch\nexpected(%v) [%v]\nfound(%v) [%v]",
			clone.Size(), clone.Bytes(),
			b.Size(), b.Bytes())
	}

	// overwrite and append
	m := len(buff) / 2
	off := int64(b.Size() - uint(m))
	n, err := b.WriteAt(buff, off)
	if err != nil {
		t.Fatalf(tag+" unexpected write error: %v", err.Error())
	}
	if n != len(buff) {
		t.Fatalf(tag+" unexpected write size, expected %v, found %v", len(buff), n)
	}

	s, err := b.SeekFromEnd(-off)
	if err != nil {
		t.Fatalf(tag+" unexpected seek error: %v", err.Error())
	}
	esp := int64(b.Size()) - off
	if s != esp {
		t.Fatalf(tag+" unexpected seek position, expected %v, found %v", esp, s)
	}

	// validate overwrite and append data
	rbuff := make([]byte, len(buff))
	n, err = b.Read(rbuff)
	if err != nil {
		t.Fatalf(tag+" unexpected read error: %v", err.Error())
	}
	if n != len(buff) {
		t.Fatalf(tag+" unexpected read size, expected %v, found %v", len(buff), n)
	}

	if bytes.Compare(buff, rbuff) != 0 {
		t.Fatalf(tag+" unexpected data in read, expected [%v], found [%v]", buff, rbuff)
	}
}

func TestByteBufferIOCopy(t *testing.T) {
	tag := "ByteBuffer@io.Copy"

	s := "abracadabra"
	src := NewByteBuffer(0)

	write_times := 3

	for i := 1; i <= write_times; i++ {
		n, err := src.Write([]byte(s))
		if err != nil {
			t.Fatalf(tag+" unexpected write error: %v", err.Error())
		}
		if n != len(s) {
			t.Fatalf(tag+" unexpected write size, expected %v, found %v", len(s), n)
		}
	}

	sp, err := src.SeekFromStart(0)
	if err != nil {
		t.Fatalf(tag+" unexpected seek error: %v", err.Error())
	}
	if sp != 0 {
		t.Fatalf(tag+" unexpected seek position, expected 0, found %v", sp)
	}

	dst := NewByteBuffer(0)
	bw, err := io.Copy(dst, src)
	if err != nil {
		t.Fatalf(tag+" unexpected error in io.Copy: %v", err.Error())
	}
	es := int64(src.Size())
	if bw != es {
		t.Fatalf(tag+" unexpected total size, expected %v, found %v", es, bw)
	}

	if dst.CmpWith(src) != 0 {
		t.Fatalf(tag+" data mismatch, expected [%v], found [%v]", src.Bytes(), dst.Bytes())
	}
}

func TestByteBufferReadWriteByte(t *testing.T) {
	tag := "ByteBuffer.ReadWriteByte"

	b := NewByteBuffer(0)

	// x will overflow to zero at first inc
	x := byte(255)
	for i := 0; i < 256; i++ {
		x++
		err := b.WriteByte(x)
		if err != nil {
			t.Fatalf(tag+" unexpected write error: %v", err.Error())
		}
	}

	_, err := b.SeekToStart()
	if err != nil {
		t.Fatalf(tag+" unexpected seek error: %v", err.Error())
	}

	for i := 0; i < 256; i++ {
		x, err = b.ReadByte()
		if err != nil {
			t.Fatalf(tag+" unexpected read error: %v", err.Error())
		}
		if x != byte(i) {
			t.Fatalf(tag+" unexpected read value, expected %v, found %v", i, x)
		}
	}

	// next read should err io.EOF
	_, err = b.ReadByte()
	if err != io.EOF {
		t.Fatalf(tag+" expected io.EOF, found %v", errOrNilStr(err))
	}
}

func TestByteBufferByteAt(t *testing.T) {
	tag := "ByteBuffer.ByteAt"

	b := NewByteBuffer(0)

	// x will overflow to zero at first inc
	x := byte(255)
	for i := 0; i < 256; i++ {
		x++
		err := b.WriteByte(x)
		if err != nil {
			t.Fatalf(tag+" unexpected write error: %v", err.Error())
		}
	}

	for i := 0; i < 256; i++ {
		x, err := b.ByteAt(i)
		if err != nil {
			t.Fatalf(tag+" unexpected read error: %v", err.Error())
		}
		if x != byte(i) {
			t.Fatalf(tag+" unexpected read value, expected %v, found %v", i, x)
		}
	}
}

func TestByteBufferUInt64Var(t *testing.T) {
	tag := "ByteBuffer.ReadWriteUInt64Var"

	vtimes := 63

	validation_list := make([]uint64, vtimes)

	b := NewByteBuffer(0)

	v := uint64(1)
	for i := 0; i < vtimes; i++ {
		validation_list[i] = v
		_, err := b.WriteUInt64Var(v)
		if err != nil {
			t.Fatalf(tag+" unexpected error: %v", err.Error())
		}
		v *= 2
	}

	_, err := b.SeekToStart()
	if err != nil {
		t.Fatalf(tag+" unexpected seek error: %v", err.Error())
	}

	for i := 0; i < vtimes; i++ {
		x, err := b.ReadUInt64Var()
		if err != nil {
			t.Fatalf(tag+" unexpected read error: %v", err.Error())
		}
		if x != validation_list[i] {
			t.Fatalf(tag+" unexpected read value @%v expected %v, found %v", i, validation_list[i], x)
		}
	}
}
