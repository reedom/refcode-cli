package uniqid_test

import (
	"context"
	"testing"

	"github.com/reedom/refcode-cli/uniqid"
)

func xTestRemote(t *testing.T) {
	s := uniqid.NewRemoteStore("http://example.com")

	ids, err := s.Generate(context.Background(), []byte("test"), nil, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != 10 {
		t.Errorf("expected len(ids) == 10 but %d", len(ids))
	}
}
