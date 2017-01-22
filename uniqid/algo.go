package uniqid

import (
	"encoding/binary"
	"io"
	"math/rand"
	"strconv"
)

// MaxInt is the maximum value of int64.
const MaxInt = (1<<63 - 1)

type Algorithm interface {
	NextValue() ([]byte, error)
	Load(r io.Reader) error
	Save(w io.Writer) error
}

type randomStringGen struct {
	size int
}

// NewRandomStringGen returns a fixed size random string generator.
// It panics if size is not in [1,32].
func NewRandomStringGen(size int) Algorithm {
	if size < 1 || 32 < size {
		panic("size should be in [1,32]")
	}
	return randomStringGen{size}
}

var AlnumChars = []byte("ABCDEFGHJKLMNPQRSTUVWXY23456789")

func (g randomStringGen) NextValue() ([]byte, error) {
	var buf [32]byte
	for i := 0; i < g.size; i++ {
		buf[i] = AlnumChars[rand.Intn(len(AlnumChars))]
	}
	return buf[0:g.size], nil
}

func (g randomStringGen) Load(r io.Reader) error {
	// nothing to do
	return nil
}

func (g randomStringGen) Save(w io.Writer) error {
	// nothing to do
	return nil
}

type randomNumberGen struct {
	min int64
	max int64
}

// NewRandomNumberGen returns a random numeric string generator.
// It panics if max is not less than min.
func NewRandomNumberGen(min, max int64) Algorithm {
	if max <= min {
		panic("max should be greater than min")
	}
	return randomNumberGen{min, max}
}

func (g randomNumberGen) NextValue() ([]byte, error) {
	return strconv.AppendInt(nil, rand.Int63n(g.max-g.min+1), 10), nil
}

func (g randomNumberGen) Load(r io.Reader) error {
	// nothing to do
	return nil
}

func (g randomNumberGen) Save(w io.Writer) error {
	// nothing to do
	return nil
}

type seqNumberGen struct {
	min int64
	max int64
	cur int64
}

func NewSeqNumberGen(opt SeqNumberGenOption) Algorithm {
	return &seqNumberGen{
		min: opt.Min,
		max: opt.Max,
		cur: opt.Min,
	}
}

func (g *seqNumberGen) NextValue() ([]byte, error) {
	if g.max < g.cur {
		return nil, ErrOutOfRange
	}

	id := strconv.AppendInt(nil, g.cur, 10)
	g.cur++
	return id, nil
}

func (g *seqNumberGen) Load(r io.Reader) error {
	var c int64
	if err := binary.Read(r, binary.LittleEndian, &c); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	g.cur = c
	return nil
}

func (g *seqNumberGen) Save(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, g.cur)
}
