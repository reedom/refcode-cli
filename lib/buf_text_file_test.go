package refcode_test

import (
	"testing"

	"bytes"
	"github.com/reedom/refcode-cli/lib"
	"io"
	"strings"
)

var testLine1 = strings.Repeat("a", 6) + "\n"
var testLine2 = "ab"

func getReaderOfTextData(t *testing.T) io.ReadSeeker {
	buf := bytes.Buffer{}
	buf.WriteString(testLine1)
	buf.WriteString(testLine1)
	buf.WriteString(testLine2)

	return bytes.NewReader(buf.Bytes())
}

func getReaderOfBinaryData(t *testing.T) io.ReadSeeker {
	buf := bytes.Buffer{}
	buf.WriteString("abc\x00def")
	return bytes.NewReader(buf.Bytes())
}

type testBufTextFile struct {
	capacity int
	expected []string
	cancel   bool
}

var testBufTextFiles = []testBufTextFile{
	{len(testLine1), []string{testLine1, testLine1, testLine2}, false},
	{len(testLine1), []string{testLine1}, true /* cancel */},
	{len(testLine1)*2 + len(testLine2) - 1, []string{testLine1 + testLine1, testLine2}, false},
	{len(testLine1)*2 + len(testLine2) + 0, []string{testLine1 + testLine1, testLine2}, false},
	{len(testLine1)*2 + len(testLine2) + 1, []string{testLine1 + testLine1 + testLine2}, false},
}

func TestBufTextFile(t *testing.T) {
	r := getReaderOfTextData(t)

	refcode.EnableVerboseLog()
	for i, test := range testBufTextFiles {
		r.Seek(0, io.SeekStart)
		buf := make([]byte, test.capacity)
		tf, err := refcode.NewBufferedTextFile(r, buf)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
			continue
		}

		for j := 0; j < 2; j++ {
			var actual []string
			err = tf.Iterate(func(b []byte) bool {
				actual = append(actual, string(b))
				return !test.cancel
			})
			if err != nil {
				t.Errorf("Unexpected error %v", err)
				break
			}

			// Ensure these files were not returned
			if !sliceEquals(test.expected, actual) {
				t.Errorf("[test %d-%d] Expects %#v but found %#v", i, j, test.expected, actual)
			}
			if err = tf.Rewind(); err != nil {
				break
			}
		}
	}
}

func TestBufTextFileErrorForShortBuf(t *testing.T) {
	f := getReaderOfTextData(t)

	var buf []byte
	_, err := refcode.NewBufferedTextFile(f, buf)
	if err == nil {
		t.Errorf("Expected error, actually no error")
	}
}

func TestBufTextFileErrorForBinaryFile(t *testing.T) {
	f := getReaderOfBinaryData(t)

	buf := make([]byte, 64)
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
