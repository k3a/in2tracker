package store

import (
	"database/sql"

	"github.com/russross/meddler"
	"k3a.me/money/backend/model"
)

const countriesTable = "countries"

func (s *Store) GetCountry(id int64) (*model.Country, error) {
	country := new(model.Country)
	err := meddler.Load(s.db, countriesTable, country, id)
	return country, err
}

func (s *Store) GetCountryByName(name string) (*model.Country, error) {
	country := new(model.Country)
	err := meddler.QueryRow(s.db, country, `SELECT * FROM `+countriesTable+
		` WHERE name = ?`, name)
	return country, err
}

func (s *Store) GetOrCreateCountry(name string) (*model.Country, error) {
	country, err := s.GetCountryByName(name)
	if err == sql.ErrNoRows {
		country = &model.Country{
			Name: name,
		}
		err = s.CreateCountry(country)
	}
	return country, err
}

func (s *Store) CreateCountry(c *model.Country) error {
	return meddler.Insert(s.db, countriesTable, c)
}

func (s *Store) UpdateCountry(c *model.Country) error {
	return meddler.Update(s.db, countriesTable, c)
}
