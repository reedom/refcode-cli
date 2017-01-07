package refcode_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/reedom/refcode-cli/lib"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestStore(t *testing.T) {
	dir, err := ioutil.TempDir("", "refcode")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	db, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := refcode.NewStore(db)

	if _, err = store.GetTime("hello"); err != refcode.ErrNotFound {
		t.Errorf("expected ErrNotFound but %v", err)
	}

	now := time.Now()
	if err = store.PutTime("hello", now); err != nil {
		t.Errorf("no error expected but %v", err)
	}

	if tm, err := store.GetTime("hello"); err != nil {
		t.Errorf("no error expected but %v", err)
	} else if !tm.Equal(now) {
		t.Errorf("expected %v but %v", now, tm)
	}
}
