package marketdata

import (
	"fmt"
	"reflect"
	"time"

	"github.com/k3a/in2tracker/backend/currency"
)

func e(format string, args ...interface{}) error {
	return fmt.Errorf("marketdata: "+format, args...)
}

// Errors
var (
	ErrNotAvailable = e("data for the selected item and time not available")
)

// MarketData specifies market info for a particular item and time
type MarketData struct {
	Time      time.Time
	LastTrade float64 // last trade price
	Currency  currency.Currency
}

// TimedMarketData represents an item value at the specific time
type TimedMarketData struct {
	Time     time.Time
	Open     float64
	Low      float64
	High     float64
	Close    float64
	Volume   float64
	Currency currency.Currency
}

// ItemInfo contains basic item information
type ItemInfo struct {
	Name   string
	Market string
}

// Provider provides market data
type Provider interface {
	// Name returns the name of the provider
	Name() string
	// Supports returns true if the item-market pair is supported by the provider.
	// Should return fast and not make any http requests (except for the first time it is called)
	// Parameter market can be empty.
	Supports(market *Market, item string) bool
	// GetMarketData gets the market price at the specific time.
	// market: market identifier (NASDAQ, CURRENCY)
	// item: stock ticker or item identifier (APPLE, USDCZK)
	GetMarketData(market *Market, item string, at time.Time) (*MarketData, error)

	// SupportsDateRange returns true if the provider supports returning data for date range
	// and GetMarketDataForDateRange works
	SupportsDateRange() bool
	// GetPriceDateRange returns historical data from tfrom to tto dates.
	GetMarketDataForDateRange(market *Market, item string, tfrom time.Time, tto time.Time) ([]*TimedMarketData, error)

	// SupportsItemInfo returns true if the provider supports returning info about the item
	SupportsItemInfo() bool
	// GetItemInfo returns item information
	// Parameter market can be empty.
	GetItemInfo(market *Market, item string) (*ItemInfo, error)
}

// Providers holds all available currency rate providers
var Providers []Provider

// RegisterProvider registers a new provider
func RegisterProvider(provider Provider) {
	for _, p := range Providers {
		if p == provider {
			panic("Attempt to register already-registered marketdata provider " +
				reflect.TypeOf(provider).String())
		}
	}
	Providers = append(Providers, provider)
}
