package fs

import (
	"os"
	"path/filepath"
)

// FoundFile represents found file by FileFinder.
type FoundFile interface {
	os.FileInfo
	Path() string
	IsSymlink() bool
	IsNamedPipe() bool
	ResolveSymlink() (FoundFile, error)
}

// FoundFile represents found file by FileFinder.
type foundFile struct {
	path string
	os.FileInfo
}

// newFoundFile returns new FoundFile object.
func newFoundFile(path string) (FoundFile, error) {
	info, err := os.Lstat(path)
	return foundFile{path, info}, err
}

// wrapFoundFile wraps os.FileInfo and returns new FoundFile object.
func wrapFoundFile(path string, info os.FileInfo) FoundFile {
	return foundFile{path, info}
}

// Path returns f.path.
func (f foundFile) Path() string {
	return f.path
}

// IsSymlink reports whether f describes a symbolic link.
func (f foundFile) IsSymlink() bool {
	return f.FileInfo.Mode()&os.ModeSymlink == os.ModeSymlink
}

// IsNamedPipe reports whether f describes a named pipe.
func (f foundFile) IsNamedPipe() bool {
	return f.FileInfo.Mode()&os.ModeNamedPipe == os.ModeNamedPipe
}

// ResolveSymlink determines whether f describes a symbolic link of a directory.
func (f foundFile) ResolveSymlink() (FoundFile, error) {
	entity, err := filepath.EvalSymlinks(f.path)
	if err != nil {
		return nil, err
	}
	return newFoundFile(entity)
}
