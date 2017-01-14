package uniqid_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/reedom/refcode-cli/uniqid"
)

func Test(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "refcode-cli")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)

	s := uniqid.NewFileSeq(tempdir)
	ids, err := s.Generate(context.Background(), []byte("a"), []byte("b"), 3)
	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != 3 {
		t.Errorf("expected len(ids) == 3 but %d", len(ids))
	} else {
		for i, expected := range []string{"1", "2", "3"} {
			if !bytes.Equal(ids[i], []byte(expected)) {
				t.Errorf("expected ids[%v] == %v but %v", i, expected, string(ids[i]))
			}
		}
	}
}
