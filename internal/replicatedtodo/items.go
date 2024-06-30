// inspired by https://adamreeve.co.nz/blog/todo-crdt.html
package replicatedtodo

import (
	"encoding/json"
	"math/big"
	"slices"

	"github.com/google/uuid"
)

type Item struct {
	Title string
	State string
	ID    uuid.UUID
}

type ItemList struct {
	replicated PersistedModel
}

func (m *ItemList) Items() []Item {
	var items []Item
	for _, v := range m.replicated.sorted() {
		items = append(items, v.Item())
	}
	return items
}

func (m *ItemList) NewTodo(title string, previous uuid.UUID) (*Item, error) {
	items := m.replicated.sorted()
	previousIndex := slices.IndexFunc(items, func(item *PersistedItem) bool {
		return item.ID == previous
	})
	var order big.Rat
	if previousIndex >= 0 {
		order.Set(&items[previousIndex].Order)
	}
	if previousIndex+1 < len(items) {
		high := &items[previousIndex+1].Order
		order.Add(&order, high)
	} else {
		order.Add(&order, big.NewRat(1, 1))
	}
	order.Quo(&order, big.NewRat(2, 1))

	id, err := m.replicated.NewTodo(title, order)
	if err != nil {
		return nil, err
	}
	return m.replicated.GetItem(id), nil
}

// MarshalJSON implements the json.Marshaller interface
func (m *ItemList) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(m.replicated)
	return bytes, err
}

var _ json.Marshaler = &ItemList{}

// UnmarshalJSON implements the json.Unmarshaller interface
func (m *ItemList) UnmarshalJSON(bytes []byte) error {
	var replicated PersistedModel
	if err := json.Unmarshal(bytes, &replicated); err != nil {
		return err
	}
	m.replicated = replicated
	return nil
}

var _ json.Unmarshaler = &ItemList{}
