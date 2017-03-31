package store

import (
	"github.com/russross/meddler"
	"k3a.me/money/backend/model"
)

// GetItem returns item (company) by ID
func (s *Store) GetItem(id int64) (*model.Item, error) {
	item := new(model.Item)
	err := meddler.Load(s.db, "items", item, id)
	return item, err
}
