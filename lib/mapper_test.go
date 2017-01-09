package refcode_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/reedom/refcode-cli/lib"
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

	opts := refcode.Option{
		Codespace: "tests",
		DataDir:   filepath.Join(tmpdir, "data"),
		Mapper: refcode.MapperOpt{
			Marker:         "@@REFCODE",
			ReplaceFormat:  "%d",
			InChannelCount: 2,
			ParallelCount:  2,
			WorkBufSize:    2,
		},
		FileFinder: refcode.FileFinderOpt{
			Includes:        []string{"*.js"},
			GlobalGitIgnore: false,
		},
	}

	mapper, err := refcode.NewMapper(opts)
	if err != nil {
		t.Fatal(err)
	}

	// refcode.EnableVerboseLog()
	err = mapper.Run(context.Background(), tmpdir)
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
