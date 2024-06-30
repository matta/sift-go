package replicatedtodo

import (
	"testing"
)

func TestNew(t *testing.T) {
	got := New()
	if got.Items == nil {
		t.Errorf("New() = %v, want non-nil Items", got)
	}
	if len(got.Items) != 0 {
		t.Errorf("New() = %v, want non-nil Items", got)
	}
}
