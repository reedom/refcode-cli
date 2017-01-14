package uniqid

import (
	"math/rand"
	"strconv"
)

type IdGenerator interface {
	Generate() []byte
}

type randomStringGen struct {
	size int
}

var AlnumChars = []byte("ABCDEFGHJKLMNPQRSTUVWXY23456789")

func (g randomStringGen) Generate() []byte {
	var buf [32]byte
	for i := 0; i < g.size; i++ {
		buf[i] = AlnumChars[rand.Intn(len(AlnumChars))]
	}
	return buf[0:g.size]
}

// NewRandomStringGen returns a fixed size random string generator.
// It panics if size is not in [1,32].
func NewRandomStringGen(size int) IdGenerator {
	if size < 1 || 32 < size {
		panic("size should be in [1,32]")
	}
	return randomStringGen{size}
}

type randomNumberGen struct {
	min int64
	max int64
}

// NewRandomNuberGen returns a random numeric string generator.
// It panics if max is not less than min.
func NewRandomNuberGen(min, max int64) IdGenerator {
	if max <= min {
		panic("max should be greater than min")
	}
	return randomNumberGen{min, max}
}

func (g randomNumberGen) Generate() []byte {
	return strconv.AppendInt(nil, rand.Int63n(g.max-g.min+1), 10)
}
