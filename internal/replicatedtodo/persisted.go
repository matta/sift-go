// inspired by https://adamreeve.co.nz/blog/todo-crdt.html
package replicatedtodo

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

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
	Order *big.Rat
	ID    uuid.UUID
}

func (i *PersistedItem) String() string {
	return fmt.Sprintf("PersistedItem @%p title=%q state=%q order=%d/%d id=%q",
		i, i.Title, i.State, i.Order.Num(), i.Order.Denom(), i.ID.String())
}

func newPersistedItem(title string, order *big.Rat) (*PersistedItem, error) {
	if big.NewRat(0, 1).Cmp(order) != -1 || order.Cmp(big.NewRat(1, 1)) != -1 {
		return nil, errors.New("Order out of range, need (0..1) (non-inclusive)")
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	item := &PersistedItem{
		Title: newPersistedString(title),
		State: newPersistedString("unchecked"),
		Order: order,
		ID:    id,
	}
	return item, nil
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

func (model *PersistedModel) GetItem(id uuid.UUID) *Item {
	item := model.getItem(id)

	return &Item{
		ID:    id,
		Title: item.Title.Value,
		State: item.State.Value,
	}
}

func (model *PersistedModel) sorted() []*PersistedItem {
	var items []*PersistedItem
	for v := range maps.Values(model.Items) {
		items = append(items, v)
	}
	slices.SortFunc(items, func(i, j *PersistedItem) int {
		c := i.Order.Cmp(j.Order)
		if c != 0 {
			return c
		}
		var a, b [16]byte
		a = i.ID
		b = j.ID
		return bytes.Compare(a[:], b[:])
	})
	return items
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

func (model *PersistedModel) NewTodo(title string, order *big.Rat) (uuid.UUID, error) {
	item, err := newPersistedItem(title, order)
	if err != nil {
		return uuid.UUID{}, err
	}

	if model.Items == nil {
		model.Items = make(map[uuid.UUID]*PersistedItem)
	}
	model.Items[item.ID] = item

	return item.ID, nil
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
		builder.WriteString(fmt.Sprintf("%v:\n  %s\n", id, item))
	}

	return builder.String()
}

func (model *PersistedModel) Model() ItemList {
	return ItemList{
		// TODO: deep clone?
		replicated: *model,
	}
}
