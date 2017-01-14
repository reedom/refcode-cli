package finder

import (
	"io/ioutil"
	"os"
)

type fileInfo struct {
	path string
	os.FileInfo
}

func (f fileInfo) isDir(follow bool) bool {
	if follow && f.isSymlink() {
		if _, err := ioutil.ReadDir(f.path); err == nil {
			return true
		}
		return false
	}
	return f.FileInfo.IsDir()
}

func (f fileInfo) isSymlink() bool {
	return f.FileInfo.Mode()&os.ModeSymlink == os.ModeSymlink
}

func (f fileInfo) isNamedPipe() bool {
	return f.FileInfo.Mode()&os.ModeNamedPipe == os.ModeNamedPipe
}

func newFileInfo(path string, info os.FileInfo) fileInfo {
	return fileInfo{path, info}
}
