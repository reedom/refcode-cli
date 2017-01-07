package refcode

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
)

// ErrLineTooLong means that the file contains too long line
// for the read buffer.
// To prevent this error, the caller should prepare more larger
// buffer.
var ErrLineTooLong = errors.New("file line is too long")

// ErrBinaryFile means that the file is likely a binary data file.
var ErrBinaryFile = errors.New("file content is likely binary data")

// BufferedTextFile provides an interface for reading a block of text
// in a effective fasion.
type BytesIterator interface {
	Iterate(fn func([]byte) bool) error
	Rewind() error
}

type bufferdTextFile struct {
	f   io.ReadSeeker
	buf []byte
	off int
	pos int
	eof bool
}

// NewBytesIterator returns a new BytesIterator to read from f.
// It uses buf for a temporaly buffer.
// It is the caller's responsibility to prepare buf with enough capacity.
func NewBytesIterator(f io.ReadSeeker, buf []byte) (BytesIterator, error) {
	if len(buf) == 0 {
		return nil, errors.New("buf should have meaningful size")
	}

	return &bufferdTextFile{
		f:   f,
		buf: buf,
		off: -1,
	}, nil
}

func (f *bufferdTextFile) hasEntireContent() bool {
	return f.eof && f.off == f.pos
}

func (f *bufferdTextFile) Iterate(fn func([]byte) bool) error {
	if f.hasEntireContent() {
		b := f.buf[0:f.off]
		if 0 < bytes.IndexByte(b, 0x00) {
			return ErrBinaryFile
		}

		fn(b)
		return nil
	}

	f.off = 0
	for {
		free := f.buf[f.off:]
		c, err := f.f.Read(free)
		if err != nil && err != io.EOF {
			return err
		}
		if c == 0 && f.off == 0 { // EOF
			f.eof = true
			return nil
		}
		f.pos += c

		f.off += c
		b := f.buf[0:f.off]
		if 0 < bytes.IndexByte(b, 0x00) {
			return ErrBinaryFile
		}

		if f.off < len(f.buf) {
			// reached to the end
			f.eof = true
			fn(b)
			return nil
		}

		// possibly more contents left in the stream

		i := bytes.LastIndexByte(b, '\n')
		if i < 0 {
			return ErrLineTooLong
		}

		if !fn(b[0 : i+1]) {
			// callee wants to cancel
			return nil
		}

		// slide the remainig data to the head
		tail := b[i+1 : f.off]
		copy(f.buf, tail)
		f.off = len(tail)
	}
}

func (f *bufferdTextFile) Rewind() error {
	if f.off < 0 {
		// initial state.
		return nil
	}

	if f.hasEntireContent() {
		return nil
	}

	// rewind
	_, err := f.f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	f.pos = 0
	f.off = 0
	f.eof = false
	return nil
}
