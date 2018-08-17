package marketdata

import (
	"net/http"
	"time"
)

// DummyProvider is a market data provider
type DummyProvider struct {
	httpClient *http.Client
}

// NewDummyProvider creates a new data fetcher
func NewDummyProvider() *DummyProvider {
	return &DummyProvider{
		&http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the name of the provider
func (md *DummyProvider) Name() string {
	return "Dummy"
}

// Supports returns true if the item-market pair is supported by the provider.
// Should return fast and not make any http requests (except for the first time it is called)
func (md *DummyProvider) Supports(market *Market, ticker string) bool {
	return false
}

// SupportsDateRange returns true if the provider supports returning data for date range
// and GetPriceDateRange works
func (md *DummyProvider) SupportsDateRange() bool {
	return false
}

// GetMarketDataForDateRange returns historical data from tfrom to tto dates.
func (md *DummyProvider) GetMarketDataForDateRange(market *Market, ticker string, tfrom time.Time, tto time.Time) ([]*TimedMarketData, error) {
	return nil, ErrNotAvailable
}

// GetMarketData gets the market price at the specific time.
// market: market identifier (NASDAQ, CURRENCY)
// ticker: stock ticker (APPLE, USDCZK)
func (md *DummyProvider) GetMarketData(market *Market, ticker string, at time.Time) (*MarketData, error) {
	return nil, ErrNotAvailable
}

// SupportsItemInfo returns true if the provider supports returning info about the item
func (md *DummyProvider) SupportsItemInfo() bool {
	return false
}

// GetItemInfo returns item information
// Parameter market can be empty.
func (md *DummyProvider) GetItemInfo(market *Market, item string) (*ItemInfo, error) {
	return nil, ErrNotAvailable
}

func init() {
	RegisterProvider(NewDummyProvider())
}
