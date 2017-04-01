package main

import (
	"github.com/k3a/in2tracker/backend/currency"
	"github.com/k3a/in2tracker/backend/model"
)

// ProcessItem holds processed result for a single item (like a single stock)
// It can be result of many processed transactions
type ProcessItem struct {
	Item    *model.Item
	Country *model.Country

	// item currency
	Currency currency.Currency
	// tax paid in item currency
	DividendTaxPaid                  float64
	DividendTaxPaidInPrimaryCurrency float64
	DividendIncomeInPrimaryCurrency  float64
}

// ProcessCountry holds processed result for a single country.
// It is used to store country-data like income per country
type ProcessCountry struct {
	// items belonging to the country
	Items                                 []*ProcessItem
	TotalDividendIncomeInPrimaryCurrency  float64
	TotalDividendTaxPaidInPrimaryCurrency float64
}

// ProcessResult holds the complete result of process operation
type ProcessResult struct {
	// primary currency used during processing (must be set in the constructor only)
	PrimaryCurrency currency.Currency
	// results split by individual countries
	Countries map[string]*ProcessCountry
	// total net gain/loss (income) by currency
	TotalGainLossByCurrency map[currency.Currency]float64
	// total revenues from stock/item sells (excl. any fees)
	TotalRevenuesInPrimaryCurrency float64
	// total expenses from stock/item purchases and sells (costs + fees)
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
