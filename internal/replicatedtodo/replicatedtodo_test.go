package replicatedtodo_test

import (
	"example/user/sift/internal/replicatedtodo"
	"testing"
)

func TestModel_NewTodo(t *testing.T) {
	model := replicatedtodo.New()
	model.NewTodo("hello")

	if len(model.Items) != 1 {
		t.Errorf("len(model.Items) = %v, want %v", len(model.Items), 1)
	}

	if len(model.States) != 1 {
		t.Errorf("len(model.States) = %v, want %v", len(model.States), 1)
	}
}

func TestNew(t *testing.T) {
	got := replicatedtodo.New()
	if got.States == nil {
		t.Errorf("New() = %v, want non-nil States", got)
	}
}
