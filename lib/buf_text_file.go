package refcode

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
	"os"
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
type BufferedTextFile interface {
	Iterate(fn func([]byte) bool) error
}

type bufferdTextFile struct {
	f   *os.File
	buf *bytes.Buffer
}

// NewBufferedTextFile returns a new BufferedTextFile to read from f.
// It uses buf for a temporaly buffer.
// It is the caller's responsibility to prepare buf with enough capacity.
func NewBufferedTextFile(f *os.File, buf *bytes.Buffer) (BufferedTextFile, error) {
	if buf.Cap() == 0 {
		return nil, errors.New("buf should be initialized with Grow()")
	}
	if 0 < buf.Len() {
		return nil, errors.New("buf should be empty")
	}

	return bufferdTextFile{f, buf}, nil
}

func (f bufferdTextFile) Iterate(fn func([]byte) bool) error {
	f.Rewind()

	if 0 < f.buf.Len() {
		// it already read entire content

		b := f.buf.Bytes()
		if 0 < bytes.IndexByte(b, 0x00) {
			return ErrBinaryFile
		}

		fn(b)
		return nil
	}

	for {
		freeLen := int64(f.buf.Cap() - f.buf.Len())
		c, err := f.buf.ReadFrom(io.LimitReader(f.f, freeLen))
		if err != nil && err != io.EOF {
			return err
		}
		if c == 0 { // EOF
			return nil
		}

		b := f.buf.Bytes()
		if 0 < bytes.IndexByte(b, 0x00) {
			return ErrBinaryFile
		}

		if c < freeLen {
			// reached to the end
			fn(b)
			return nil
		}

		// possibly more contents left in the disk.

		i := bytes.LastIndexByte(b, '\n')
		if i < 0 {
			return ErrLineTooLong
		}

		next := fn(b[0 : i+1])
		if !next {
			return nil
		}

		left := b[i+1:]
		f.buf.Reset()
		f.buf.Write(left)
	}
}

func (f *bufferdTextFile) Rewind() error {
	if f.buf.Len() == 0 {
		// initial state.
		return nil
	}

	pos, err := f.f.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	if pos < int64(f.buf.Cap()) {
		// it has read entire content.
		return nil
	}

	// it has read some of the content and it is larger than the buffer.
	f.buf.Reset()
	f.f.Seek(0, io.SeekStart)
	return nil
}
