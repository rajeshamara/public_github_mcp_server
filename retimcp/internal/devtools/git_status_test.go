package devtools

import (
	"context"
	"testing"
)

func TestCancelledOperation(t *testing.T) {
	s, _, _ := fixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := s.Status(ctx); e == nil {
		t.Fatal("cancelled command succeeded")
	}
}
