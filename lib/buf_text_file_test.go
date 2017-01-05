package refcode_test

import (
	"io/ioutil"
	"os"
	"testing"

	"bytes"
	"github.com/reedom/refcode-cli/lib"
	"io"
	"strings"
)

var testBufTextFileLine = strings.Repeat("a", 70) + "\n"

func createTestFileForBufTextFile(t *testing.T) *os.File {
	f, err := ioutil.TempFile("", "refcode-cli")
	if err != nil {
		t.Fatal(err)
	}

	line := []byte(testBufTextFileLine)
	_, err = f.Write(line)
	_, err = f.Write(line)
	if err != nil {
		t.Fatal(err)
	}

	f.Seek(0, io.SeekStart)
	return f
}

func createTestFileForBufTextFileBinary(t *testing.T) *os.File {
	f, err := ioutil.TempFile("", "refcode-cli")
	if err != nil {
		t.Fatal(err)
	}

	line := []byte("abc\x00def")
	_, err = f.Write(line)
	if err != nil {
		t.Fatal(err)
	}

	f.Seek(0, io.SeekStart)
	return f
}

type testBufTextFile struct {
	capacity int
	expected []string
	cancel   bool
}

var testBufTextFiles = []testBufTextFile{
	{len(testBufTextFileLine), []string{testBufTextFileLine, testBufTextFileLine}, false},
	{len(testBufTextFileLine), []string{testBufTextFileLine}, true /* cancel */},
	{len(testBufTextFileLine)*2 - 1, []string{testBufTextFileLine, testBufTextFileLine}, false},
	{len(testBufTextFileLine) * 2, []string{testBufTextFileLine + testBufTextFileLine}, false},
}

func TestBufTextFile(t *testing.T) {
	f := createTestFileForBufTextFile(t)
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	for _, test := range testBufTextFiles {
		buf := &bytes.Buffer{}
		buf.Grow(test.capacity)
		f.Seek(0, io.SeekStart)
		tf, err := refcode.NewBufferedTextFile(f, buf)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
			continue
		}

		var actual []string
		err = tf.Iterate(func(b []byte) bool {
			actual = append(actual, string(b))
			return !test.cancel
		})
		if err != nil {
			t.Errorf("Unexpected error %v", err)
			continue
		}

		// Ensure these files were not returned
		if !sliceEquals(test.expected, actual) {
			t.Errorf("Expects %v but found %v", test.expected, actual)
		}
	}
}

func TestBufTextFileErrorForShortBuf(t *testing.T) {
	f := createTestFileForBufTextFile(t)
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	buf := &bytes.Buffer{}
	_, err := refcode.NewBufferedTextFile(f, buf)
	if err == nil {
		t.Errorf("Expected error, actually no error")
	}
}

func TestBufTextFileErrorForBinaryFile(t *testing.T) {
	f := createTestFileForBufTextFileBinary(t)
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	buf := &bytes.Buffer{}
	buf.Grow(64)
	tf, err := refcode.NewBufferedTextFile(f, buf)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	err = tf.Iterate(func(b []byte) bool {
		return true
	})
	if err != refcode.ErrBinaryFile {
		t.Errorf("Expected ErrBinaryFile but %v", err)
	}
}
