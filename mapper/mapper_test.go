package mapper_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/reedom/refcode-cli/mapper"
)

func TestMapper(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "refcode-cli")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	code := []byte(`
// var refcode = "@@REFCODE@@REFCODE";
alertError("@@REFCODE");
return;
`)
	expected := []byte(`
// var refcode = "12";
alertError("3");
return;
`)

	path := filepath.Join(tmpdir, "one.js")
	err = ioutil.WriteFile(path, code, 0644)
	if err != nil {
		t.Fatal(err)
	}

	opts := mapper.Option{
		Codespace:      "tests",
		DataDir:        filepath.Join(tmpdir, "data"),
		Marker:         "@@REFCODE",
		ReplaceFormat:  "%d",
		InChannelCount: 2,
		ParallelCount:  2,
		WorkBufSize:    2,
	}

	mapper, err := mapper.NewMapper(opts, finder{path}, idgen{})
	if err != nil {
		t.Fatal(err)
	}

	// mapper.EnableVerboseLog()
	err = mapper.Run(context.Background())
	if err != nil {
		t.Error("mapper.Run returns error:", err)
		return
	}
	actual, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(expected, actual) {
		t.Errorf("mapper result unmatch\nexpected: %s\nactual: %s", string(expected), string(actual))
	}
}

type finder struct {
	path string
}

func (f finder) Start(ctx context.Context, out chan string) {
	out <- f.path
	close(out)
}

type idgen struct {
}

func (g idgen) Generate(ctx context.Context, key, sub []byte, n int) ([][]byte, error) {
	codes := make([][]byte, n)
	for i := range codes {
		codes[i] = strconv.AppendInt(nil, int64(i+1), 10)
	}
	return codes, nil
}
