package store

import (
	"github.com/russross/meddler"
	"k3a.me/money/backend/model"
)

const marketsTable = "markets"

func (s *Store) GetMarket(id int64) (*model.Market, error) {
	country := new(model.Market)
	err := meddler.Load(s.db, marketsTable, country, id)
	return country, err
}

func (s *Store) CreateMarket(c *model.Market) error {
	return meddler.Insert(s.db, marketsTable, c)
}

func (s *Store) UpdateMarket(c *model.Market) error {
	return meddler.Update(s.db, marketsTable, c)
}
