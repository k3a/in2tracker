package main

import (
	"time"

	"k3a.me/money/backend/currency"
	"k3a.me/money/backend/store"
)

//TODO: move to ./currency module with optional interface to store for caching

type CurrencyCache struct {
	store *store.Store
}

func NewCurrencyCache(store *store.Store) *CurrencyCache {
	return &CurrencyCache{store}
}

func (cc *CurrencyCache) Convert(amount float64, from currency.Currency, to currency.Currency, at time.Time) (float64, error) {
	// if same currency copy over
	if from == to {
		return amount, nil
	}

	// try find direct rate
	mult, err := cc.store.GetCurrencyMultiplier(at, from, to)
	if err != nil {
		// try reverse
		mult, err = cc.store.GetCurrencyMultiplier(at, to, from)
		mult = 1.0 / mult
	}

	if err != nil {
		// get live data
		mult, err = currency.Convert(1.0, from, to, at)
		if err != nil {
			return 0, err
		}

		// cache data
		err = cc.store.StoreCurrencyMultiplier(at, from, to, mult)
		if err != nil {
			return 0, err
		}
	}

	return amount * mult, nil
}
