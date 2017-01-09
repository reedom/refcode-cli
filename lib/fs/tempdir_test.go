package fs_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type tempdir struct {
	t    *testing.T
	path string
}

func newTempdir(t *testing.T) (tempdir, error) {
	path, err := ioutil.TempDir("", "fileinfo")
	return tempdir{t, path}, err
}

func (t tempdir) removeAll() {
	os.RemoveAll(t.path)
}

func (t tempdir) fullpath(relpath string) string {
	return filepath.Join(t.path, relpath)
}

func (t tempdir) file(relpath string) {
	fullpath := t.fullpath(relpath)
	os.MkdirAll(filepath.Dir(fullpath), 0755)
	if err := ioutil.WriteFile(fullpath, []byte{}, 0644); err != nil {
		t.t.Fatalf("failed to write file %q: %v", relpath, err)
	}
}

func (t tempdir) dir(relpath string) {
	if err := os.MkdirAll(t.fullpath(relpath), 0755); err != nil {
		t.t.Fatalf("failed to mkdir %q: %v", relpath, err)
	}
}

func (t tempdir) symlink(oldname, newname string) {
	if err := os.Symlink(t.fullpath(oldname), t.fullpath(newname)); err != nil {
		t.t.Fatalf("failed to create symlink %q->%q: %v", oldname, newname, err)
	}
}
