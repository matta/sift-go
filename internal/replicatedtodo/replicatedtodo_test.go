package replicatedtodo_test

import (
	"testing"

	"github.com/matta/sift/internal/replicatedtodo"
)

func TestModel_NewTodo(t *testing.T) {
	model := replicatedtodo.New()
	model.NewTodo("hello")

	if len(model.Items) != 1 {
		t.Errorf("len(model.Items) = %v, want %v", len(model.Items), 1)
	}
}

func TestNew(t *testing.T) {
	got := replicatedtodo.New()
	if got.Items == nil {
		t.Errorf("New() = %v, want non-nil Items", got)
	}
	if len(got.Items) != 0 {
		t.Errorf("New() = %v, want non-nil Items", got)
	}
}
