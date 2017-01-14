package mapper_test

import (
	"bytes"
	"testing"

	"context"
	"fmt"
	"github.com/reedom/refcode-cli/mapper"
)

func TestCountMarkerInContent(t *testing.T) {
	code := []byte(`
// var refcode = "@@REFCODE@@REFCODE";
alertError("@@REFCODE");
return;
`)
	r := bytes.NewReader(code)
	c, err := mapper.CountMarkerInContent(context.Background(), r, []byte("@@REFCODE"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if c != 3 {
		t.Errorf("expected count is 3 but %v", c)
	}
}

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
	err := mapper.TransformContent(context.Background(), r, w, []byte("@@REFCODE"), fn)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if !bytes.Equal(expected, w.Bytes()) {
		t.Errorf("mapper result unmatch\n  expected: %s\n  actual: %s", string(expected), string(w.Bytes()))
	}
}
