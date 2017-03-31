package store

import (
	"github.com/russross/meddler"
	"k3a.me/money/backend/model"
)

const itemsTable = "items"

// GetItem returns item (company) by ID
func (s *Store) GetItem(id int64) (*model.Item, error) {
	item := new(model.Item)
	err := meddler.Load(s.db, itemsTable, item, id)
	return item, err
}

func (s *Store) CreateItem(item *model.Item) error {
	return meddler.Insert(s.db, itemsTable, item)
}

func (s *Store) UpdateItem(item *model.Item) error {
	return meddler.Update(s.db, itemsTable, item)
}
