package uniqid

import (
	"errors"
)

// ErrOutOfRange means that the generator has reached to its maximum value.
var ErrOutOfRange = errors.New("no error code is available")
