package store

import (
	"fmt"
	"time"

	"github.com/russross/meddler"
	"k3a.me/money/backend/currency"
	"k3a.me/money/backend/model"
)

const currenciesTable = "currencies"
const currencyPairsTable = "currency_pairs"

// GetCurrency returns currency detail for specified currency code
func (s *Store) GetCurrency(code currency.Currency) (*model.Currency, error) {
	var c model.Currency
	err := meddler.QueryRow(s.db, &c, "SELECT * FROM currencies WHERE code = ?", code)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetOrCreateCurrency returns currency detail or creates a new currency
// Nil is returned on error.
func (s *Store) GetOrCreateCurrency(code currency.Currency) *model.Currency {
	ret, err := s.GetCurrency(code)
	if err == nil {
		return ret
	}

	def := &model.Currency{
		Code: code.String(),
		Name: code.Name(),
	}

	err = meddler.Insert(s.db, currenciesTable, def)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return def
}

// GetCurrencyMultiplier finds multiplier for converting "from" currency to "to" currency.
// It does not check reverse record.
func (s *Store) GetCurrencyMultiplier(date time.Time, from currency.Currency, to currency.Currency) (float64, error) {
	src, err := s.GetCurrency(from)
	if err != nil {
		return 0, err
	}

	dst, err := s.GetCurrency(to)
	if err != nil {
		return 0, err
	}

	var cp model.CurrencyPair
	err = meddler.QueryRow(s.db, &cp, `SELECT * FROM `+currencyPairsTable+` 
		WHERE src_currency_id = $1 AND dst_currency_id = $2 AND date = $3`,
		src.ID, dst.ID, date.UTC()) // time is always stored in UTC in the DB
	if err != nil {
		return 0, err
	}

	return cp.Multiplier, nil
}

// StoreCurrencyMultiplier stores multiplier for the specified date
func (s *Store) StoreCurrencyMultiplier(date time.Time, from currency.Currency, to currency.Currency, mult float64) error {
	src := s.GetOrCreateCurrency(from)
	if src == nil {
		return e("unable to create currency")
	}

	dst := s.GetOrCreateCurrency(to)
	if dst == nil {
		return e("unable to create currency")
	}

	return meddler.Save(s.db, currencyPairsTable, &model.CurrencyPair{
		Date:          date,
		SrcCurrencyID: src.ID,
		DstCurrencyID: dst.ID,
		Multiplier:    mult,
	})
}
