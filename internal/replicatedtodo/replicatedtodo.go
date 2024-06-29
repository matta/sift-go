// inspired by https://adamreeve.co.nz/blog/todo-crdt.html
package replicatedtodo

import (
	"cmp"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/log" // TODO: remove this dependency
	"github.com/google/uuid"
)

type Model struct {
	replicated PersistedModel
}

func (m *Model) AllItems() []Item {
	var items []Item
	for v := range maps.Values(m.replicated.Items) {
		items = append(items, v.Item())
	}
	slices.SortFunc(items, func(i, j Item) int {
		slog.Warn("sort by more than ID here")
		return cmp.Compare(i.ID.String(), j.ID.String())
	})
	return items
}

func (m *Model) NewTodo(title string) Item {
	id := m.replicated.NewTodo(title)
	return m.replicated.GetItem(id)
}

// MarshalJSON implements the json.Marshaller interface
func (m *Model) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(m.replicated)
	return bytes, err
}

// UnmarshalJSON implements the json.Unmarshaller interface
func (m *Model) UnmarshalJSON(bytes []byte) error {
	var replicated PersistedModel
	if err := json.Unmarshal(bytes, &replicated); err != nil {
		return err
	}
	m.replicated = replicated
	return nil
}

type Item struct {
	Title string
	State string
	ID    uuid.UUID
}

// PersistedModel is a collection of CRTDs used to store a synchronized set of items.
//
// Initial design inspired from https://adamreeve.co.nz/blog/todo-crdt.html
type PersistedModel struct {
	// This is a "grow only set" (G-Set) of items, keyed by UUID.
	Items map[uuid.UUID]*PersistedItem
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
	Order big.Rat
	ID    uuid.UUID
}

func (i *PersistedItem) DebugString() any {
	return fmt.Sprintf("PersistedItem @%p title=%q state=%q", i, i.Title, i.State)
}

func newPersistedItem(title string) *PersistedItem {
	uuid, err := uuid.NewV7()
	if err != nil {
		log.Fatalf("Can't generate UUID: %v", err)
	}
	return &PersistedItem{Title: newPersistedString(title), State: newPersistedString("unchecked"), ID: uuid}
}

func (i *PersistedItem) Item() Item {
	return Item{
		Title: i.Title.Value,
		State: i.State.Value,
		ID:    i.ID,
	}
}

func New() *PersistedModel {
	return &PersistedModel{
		Items: make(map[uuid.UUID]*PersistedItem),
	}
}

func (model *PersistedModel) GetItem(id uuid.UUID) Item {
	item := model.getItem(id)

	return Item{
		ID:    id,
		Title: item.Title.Value,
		State: item.State.Value,
	}
}

func (model *PersistedModel) GetAllItems() []Item {
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

func (model *PersistedModel) NewTodo(title string) uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		log.Fatalf("Can't generate UUID: %v", err)
	}

	if model.Items == nil {
		model.Items = make(map[uuid.UUID]*PersistedItem)
	}
	model.Items[id] = newPersistedItem(title)

	return id
}

func (model *PersistedModel) getItem(id uuid.UUID) *PersistedItem {
	return model.Items[id]
}

func (model *PersistedModel) GetState(id uuid.UUID) string {
	return model.getItem(id).State.Value
}

func (model *PersistedModel) ToggleDone(id uuid.UUID) {
	item := model.getItem(id)
	switch item.State.Value {
	case "unchecked":
		item.State = newPersistedString("checked")
	case "checked":
		item.State = newPersistedString("unchecked")
	}
}

func (model *PersistedModel) SetTitle(id uuid.UUID, title string) {
	model.getItem(id).Title = newPersistedString(title)
}

func (model *PersistedModel) DebugString() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Model @%p\n", model))
	for id, item := range model.Items {
		builder.WriteString(fmt.Sprintf("%v:\n  %s\n", id, item.DebugString()))
	}

	return builder.String()
}

func (model *PersistedModel) Model() Model {
	return Model{
		// TODO: deep clone?
		replicated: *model,
	}
}
