package refcode_test

import (
	"bytes"
	"testing"

	"context"
	"fmt"
	"github.com/reedom/refcode-cli/lib"
)

func TestTransformContent(t *testing.T) {
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

	i := 0
	fn := func(ctx context.Context) ([]byte, error) {
		i++
		return []byte(fmt.Sprintf("%v", i)), nil
	}

	r := bytes.NewReader(code)
	w := new(bytes.Buffer)
	err := refcode.TransformContent(context.Background(), r, w, []byte("@@REFCODE"), fn)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if !bytes.Equal(expected, w.Bytes()) {
		t.Errorf("mapper result unmatch\n  expected: %s\n  actual: %s", string(expected), string(w.Bytes()))
	}
}
