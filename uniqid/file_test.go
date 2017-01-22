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

	a := uniqid.NewSeqNumberGen(uniqid.SeqNumberGenOption{1, 100})
	s := uniqid.NewFileStore(tempdir, a)
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

	ids, err = s.Generate(context.Background(), []byte("a"), []byte("b"), 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != 1 {
		t.Errorf("expected len(ids) == 1 but %d", len(ids))
	} else {
		for i, expected := range []string{"4"} {
			if !bytes.Equal(ids[i], []byte(expected)) {
				t.Errorf("expected ids[%v] == %v but %v", i, expected, string(ids[i]))
			}
		}
	}
}
