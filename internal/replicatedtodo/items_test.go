package replicatedtodo

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func ignoreField(path string) cmp.Option {
	return cmp.FilterPath(func(p cmp.Path) bool {
		return p.String() == path
	}, cmp.Ignore())
}

func TestNewTodo(t *testing.T) {
	list := ItemList{}
	item_a, err := list.NewTodo("title a", uuid.UUID{})
	if err != nil {
		t.Fatalf("error creating todo: %s", err)
	}

	item_a_expected := &Item{
		Title: "title a",
		State: "unchecked",
	}
	if diff := cmp.Diff(item_a_expected, item_a, ignoreField("ID")); diff != "" {
		t.Errorf("NewTodo(\"title a\") mismatch (-want, +got):\n%s", diff)
	}

	items := list.Items()
	items_expected := []Item{*item_a_expected}
	if diff := cmp.Diff(items_expected, items, ignoreField("ID")); diff != "" {
		t.Errorf("NewTodo(\"title a\") mismatch (-want, +got):\n%s", diff)
	}

	item_b, err := list.NewTodo("title b", item_a.ID)
	if err != nil {
		t.Fatalf("error creating todo: %s", err)
	}

	item_b_expected := &Item{
		Title: "title b",
		State: "unchecked",
	}
	if diff := cmp.Diff(item_b_expected, item_b, ignoreField("ID")); diff != "" {
		t.Errorf("NewTodo(\"title a\") mismatch (-want, +got):\n%s", diff)
	}

	items = list.Items()
	items_expected = []Item{*item_a_expected, *item_b_expected}
	if diff := cmp.Diff(items_expected, items, ignoreField("ID")); diff != "" {
		t.Errorf("NewTodo(\"title a\") mismatch (-want, +got):\n%s", diff)
	}

	fmt.Println("ADD ITEM C")
	item_c, err := list.NewTodo("title c", item_a.ID)
	if err != nil {
		t.Fatalf("error creating todo: %s", err)
	}

	item_c_expected := &Item{
		Title: "title c",
		State: "unchecked",
	}
	if diff := cmp.Diff(item_c_expected, item_c, ignoreField("ID")); diff != "" {
		t.Errorf("NewTodo(\"title a\") mismatch (-want, +got):\n%s", diff)
	}

	items = list.Items()
	items_expected = []Item{*item_a_expected, *item_c_expected, *item_b_expected}
	if diff := cmp.Diff(items_expected, items, ignoreField("ID")); diff != "" {
		t.Errorf("NewTodo(\"title a\") mismatch (-want, +got):\n%s", diff)
	}
}
