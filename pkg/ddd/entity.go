package ddd

import "fmt"

type Entity interface {
	EntityID() ID
}

type EntityBase struct {
	id ID
}

var _ Entity = (*EntityBase)(nil)

func NewEntityBase(id ID) (EntityBase, error) {
	if id.IsZero() {
		return EntityBase{}, fmt.Errorf("entity id is required")
	}
	return EntityBase{id: id}, nil
}

func (e EntityBase) EntityID() ID {
	return e.id
}
