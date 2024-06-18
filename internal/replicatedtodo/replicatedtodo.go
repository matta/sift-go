// inspired by https://adamreeve.co.nz/blog/todo-crdt.html
package replicatedtodo

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

// Model is a collection of CRTDs used to store a synchronized set of items.
//
// Initial design inspired from https://adamreeve.co.nz/blog/todo-crdt.html
type Model struct {
	// This is a "grow only set" (G-Set) of items.
	Items map[string]*PersistedItem
}

type PersistedString struct {
	Timestamp time.Time
	Value     string
}

func newPersistedString(value string) PersistedString {
	return PersistedString{Value: value, Timestamp: time.Now()}
}

type PersistedItem struct {
	Title PersistedString
	State PersistedString
}

func (i *PersistedItem) DebugString() any {
	return fmt.Sprintf("PersistedItem @%p title=%q state=%q", i, i.Title, i.State)
}

func newPersistedItem(title string) *PersistedItem {
	return &PersistedItem{Title: newPersistedString(title), State: newPersistedString("unchecked")}
}

type Item struct {
	ID    string
	Title string
	State string
}

func New() *Model {
	return &Model{
		Items: make(map[string]*PersistedItem),
	}
}

func (model *Model) GetItem(id string) Item {
	item := model.getItem(id)

	return Item{
		ID:    id,
		Title: item.Title.Value,
		State: item.State.Value,
	}
}

func (model *Model) GetAllItems() []Item {
	items := make([]Item, 0, len(model.Items))

	for id, item := range model.Items {
		items = append(items, Item{
			ID:    id,
			Title: item.Title.Value,
			State: item.State.Value,
		})
	}

	return items
}

func (model *Model) NewTodo(title string) string {
	id, err := uuid.NewV7()
	if err != nil {
		log.Fatalf("Can't generate UUID: %v", err)
	}

	idStr := id.String()
	model.Items[idStr] = newPersistedItem(title)

	return idStr
}

func (model *Model) getItem(id string) *PersistedItem {
	return model.Items[id]
}

func (model *Model) GetState(id string) string {
	return model.getItem(id).State.Value
}

func (model *Model) ToggleDone(id string) {
	item := model.getItem(id)
	switch item.State.Value {
	case "unchecked":
		item.State = newPersistedString("checked")
	case "checked":
		item.State = newPersistedString("unchecked")
	}
}

func (model *Model) SetTitle(id string, title string) {
	model.getItem(id).Title = newPersistedString(title)
}

func (model *Model) DebugString() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Model @%p\n", model))
	for id, item := range model.Items {
		builder.WriteString(fmt.Sprintf("%v:\n  %s\n", id, item.DebugString()))
	}

	return builder.String()
}
