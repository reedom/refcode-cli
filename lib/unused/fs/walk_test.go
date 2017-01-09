package fs_test

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/reedom/refcode-cli/lib/fs"
)

func TestWalkDontFollowSymlink(t *testing.T) {
	tempdir, err := newTempdir(t)
	if err != nil {
		t.Fatal(err)
	}
	defer tempdir.removeAll()

	tempdir.file("top.txt")
	tempdir.file("a/b.txt")
	tempdir.file("a/b/c.txt")
	tempdir.file("a/b/c/d.txt")
	tempdir.file("x/y/z.txt")
	tempdir.file("bottom.txt")
	tempdir.symlink("a/b/c.txt", "x/txtlink")
	tempdir.symlink("a", "x/dirlink")
	os.Chdir(tempdir.path)

	found := make([]string, 0)
	fn := func(ctx context.Context, f fs.FoundFile) error {
		if !f.IsDir() {
			found = append(found, f.Path())
		}
		if f.Path() == "x/dirlink/b/c" {
			return filepath.SkipDir
		}
		return nil
	}

	err = fs.Walk(context.Background(), "x", fn)
	if err != nil {
		t.Error(err)
	}
	sort.Strings(found)
	if expected := "x/dirlink x/textlink x/y/z.txt"; strings.Join(found, " ") != expected {
		t.Errorf("expected [%s] but %s", expected, found)
	}
}
