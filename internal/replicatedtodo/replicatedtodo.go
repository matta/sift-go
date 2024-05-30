package replicatedtodo

import (
	"log"
	"time"

	"github.com/google/uuid"
)

// Model is a collection of CRTDs used to store a synchronized set of items.
//
// Initial design inspired from https://adamreeve.co.nz/blog/todo-crdt.html
type Model struct {
	// This is a "grow only set" (G-Set) of items.
	Items []Item

	// This is a map of item IDs to their current state.
	States map[string]State
}

type Item struct {
	ID    string
	Title string
}

type State struct {
	State     string
	Timestamp time.Time
}

func (model *Model) NewTodo(title string) string {
	id, err := uuid.NewV7()
	if err != nil {
		log.Fatalf("Can't generate UUID: %v", err)
	}

	todo := Item{ID: id.String(), Title: title}
	model.Items = append(model.Items, todo)
	model.States[todo.ID] = State{"unchecked", time.Now()}

	return todo.ID
}

func (model *Model) GetState(id string) string {
	return model.States[id].State
}

func (model *Model) ToggleDone(id string) {
	switch model.GetState(id) {
	case "unchecked":
		model.States[id] = State{"checked", time.Now()}
	case "checked":
		model.States[id] = State{"unchecked", time.Now()}
	}
}

func New() Model {
	return Model{
		States: make(map[string]State),
		Items:  make([]Item, 0),
	}
}
