package refcode_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"context"
	"github.com/reedom/refcode-cli/lib"
)

func setup(t *testing.T) string {
	tmpdir, err := ioutil.TempDir("", "refcode-cli")
	if err != nil {
		t.Fatal(err)
	}

	setupFile := func(filePath string) {
		relPath := filepath.Join(tmpdir, filePath)
		os.MkdirAll(filepath.Dir(relPath), 0755)
		ioutil.WriteFile(relPath, []byte{}, 0644)
	}

	setupFile("./top.txt")
	setupFile("aaa/bbb/ccc.txt")
	setupFile("aaa/bbb/ccc.md")
	setupFile("aaa/.git/readme.txt")
	setupFile("zzz/yyy/xxx.txt")

	return tmpdir
}

type testFindFile struct {
	includes []string
	excludes []string
	expected []string
}

var testFindFiles = []testFindFile{
	{nil, nil, []string{}},
	{[]string{`*`}, []string{}, []string{`top.txt`, `aaa/bbb/ccc.txt`, `aaa/bbb/ccc.md`, `zzz/yyy/xxx.txt`}},
	{[]string{`*.txt`}, []string{`/zzz`}, []string{`top.txt`, `aaa/bbb/ccc.txt`}},
	{[]string{`*.txt`}, []string{`zz`}, []string{`top.txt`, `aaa/bbb/ccc.txt`, `zzz/yyy/xxx.txt`}},
	{[]string{`bbb/*.txt`}, []string{`zz`}, []string{`aaa/bbb/ccc.txt`}},
}

func TestFindFile(t *testing.T) {
	tmpdir := setup(t)
	defer os.RemoveAll(tmpdir)

	curdir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir(tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(curdir)

	tmpdir = "."
	for _, test := range testFindFiles {
		out := make(chan string)
		opts := refcode.FileFinderOpt{
			Includes: test.includes,
			Excludes: test.excludes,
		}
		find := refcode.NewFileFinder(out, opts)
		go find.Start(context.Background(), tmpdir)

		actual := getFoundFiles(out)

		// Ensure these files were not returned
		if !sliceEquals(test.expected, actual) {
			t.Errorf("Expects %v but found %v", test.expected, actual)
		}
	}
}

func getFoundFiles(ch chan string) []string {
	var found []string
	for path := range ch {
		found = append(found, filepath.ToSlash(path))
	}
	return found
}

func sliceEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	contains := func(m string) bool {
		for _, s := range a {
			if m == s {
				return true
			}
		}
		return false
	}

	for _, e := range b {
		if !contains(e) {
			return false
		}
	}
	return true
}
