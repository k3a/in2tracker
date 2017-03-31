package main

import (
	"k3a.me/money/backend/currency"
	"k3a.me/money/backend/model"
)

type ProcessItem struct {
	Item    *model.Item
	Country *model.Country
	Market  *model.Market
}

type ProcessResult struct {
	PrimaryCurrency                currency.Currency
	ByCountryName                  map[string][]*ProcessItem
	TotalGainLossByCurrency        map[currency.Currency]float64
	TotalRevenuesInPrimaryCurrency float64
	TotalExpensesInPrimaryCurrency float64
}

func NewProcessResult(primaryCurrency currency.Currency) *ProcessResult {
	return &ProcessResult{
		primaryCurrency,
		make(map[string][]*ProcessItem),
		make(map[currency.Currency]float64),
		0,
		0,
	}
}

// GetItem returns ProcessItem pointer or nil if not exists yet
func (pr *ProcessResult) GetItem(code string) *ProcessItem {
	for _, ci := range pr.ByCountryName {
		for _, it := range ci {
			if it.Item.Code == code {
				return it
			}
		}
	}
	return nil
}

// AddItem adds the ProcessItem to the list
func (pr *ProcessResult) AddItem(item *model.Item, country *model.Country, market *model.Market) {
	pr.ByCountryName[country.Name] = append(pr.ByCountryName[country.Name], &ProcessItem{
		item, country, market,
	})
}
