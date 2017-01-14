package mapper

import (
	"github.com/pkg/errors"
)

// ErrBinaryFile means that the file is likely a binary data file.
var ErrBinaryFile = errors.New("content is likely binary data")

// ErrNotFound returns when item(file, etc.) is not found.
var ErrNotFound = errors.New("not found")
