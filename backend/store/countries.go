package store

import "k3a.me/money/backend/model"
import "github.com/russross/meddler"

const countriesTable = "countries"

func (s *Store) GetCountry(id int64) (*model.Country, error) {
	country := new(model.Country)
	err := meddler.Load(s.db, countriesTable, country, id)
	return country, err
}

func (s *Store) CreateCountry(c *model.Country) error {
	return meddler.Insert(s.db, countriesTable, c)
}

func (s *Store) UpdateCountry(c *model.Country) error {
	return meddler.Update(s.db, countriesTable, c)
}
