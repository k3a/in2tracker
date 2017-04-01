package main

import (
	"k3a.me/money/backend/currency"
	"k3a.me/money/backend/model"
)

type ProcessItem struct {
	Item    *model.Item
	Country *model.Country

	Currency                 currency.Currency
	TaxPaid                  float64 // in company-local currency
	TaxPaidInPrimaryCurrency float64
	RevenueInPrimaryCurrency float64
}

type ProcessCountry struct {
	Items                          []*ProcessItem
	TotalRevenuesInPrimaryCurrency float64
	TotalTaxPaidInPrimaryCurrency  float64
}

type ProcessResult struct {
	PrimaryCurrency                currency.Currency
	Countries                      map[string]*ProcessCountry
	TotalGainLossByCurrency        map[currency.Currency]float64
	TotalRevenuesInPrimaryCurrency float64
	TotalExpensesInPrimaryCurrency float64
}

func NewProcessResult(primaryCurrency currency.Currency) *ProcessResult {
	return &ProcessResult{
		primaryCurrency,
		make(map[string]*ProcessCountry),
		make(map[currency.Currency]float64),
		0,
		0,
	}
}

// GetItem returns ProcessItem pointer or nil if not exists yet
func (pr *ProcessResult) GetItem(code string) *ProcessItem {
	for _, c := range pr.Countries {
		for _, it := range c.Items {
			if it.Item.Code == code {
				return it
			}
		}
	}
	return nil
}

// AddItem adds the ProcessItem to the list
func (pr *ProcessResult) AddItem(item *model.Item, country *model.Country) *ProcessItem {
	if _, has := pr.Countries[country.Name]; !has {
		pr.Countries[country.Name] = new(ProcessCountry)
	}

	pritem := &ProcessItem{
		item,
		country,
		currency.Invalid,
		0,
		0,
		0,
	}

	pr.Countries[country.Name].Items = append(pr.Countries[country.Name].Items, pritem)

	return pritem
}
