package store

import (
	"testing"
	"time"

	"github.com/k3a/in2tracker/backend/currency"
	"github.com/stretchr/testify/require"
)

func TestCurrency(t *testing.T) {
	db := openTest()
	defer db.Close()

	s := From(db)

	// currency should not exist in empty DB
	_, err := s.GetCurrency(currency.USD)
	require.Error(t, err)

	// create a new currency
	createdCurr, err := s.GetOrCreateCurrency(currency.USD)
	require.Nil(t, err)
	require.NotNil(t, createdCurr)
	require.Equal(t, currency.USD.Name(), createdCurr.Name)

	// try get nonexisting pair
	now := time.Now()
	_, err = s.GetCurrencyMultiplier(now, currency.USD, currency.CZK)
	require.NotNil(t, err)

	// store new pair
	err = s.StoreCurrencyMultiplier(now, currency.USD, currency.CZK, 3256.812)
	require.Nil(t, err)

	// read the stored pair
	readMult, err := s.GetCurrencyMultiplier(now, currency.USD, currency.CZK)
	require.Nil(t, err)
	require.Equal(t, 3256.812, readMult)
}
