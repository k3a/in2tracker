package marketdata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/k3a/in2tracker/backend/currency"
)

func iexError(format string, args ...interface{}) error {
	return fmt.Errorf("iex: "+format, args...)
}

// IEXProvider is a market data provider
type IEXProvider struct {
	httpClient *http.Client
}

// NewIEXProvider creates a new data fetcher
func NewIEXProvider() *IEXProvider {
	return &IEXProvider{
		&http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the name of the provider
func (md *IEXProvider) Name() string {
	return "IEX"
}

// Supports returns true if the item-market pair is supported by the provider.
// Should return fast and not make any http requests (except for the first time it is called)
func (md *IEXProvider) Supports(market *Market, ticker string) bool {
	if market == nil {
		// allow supporting queries without market for IEX
		return true
	}

	if MarketEquals(market, MarketUSANYSE) {
		return true
	}

	if MarketEquals(market, MarketUSANYSEArca) {
		return true
	}

	if MarketEquals(market, MarketUSANasdaq) {
		return true
	}

	return false
}

// SupportsDateRange returns true if the provider supports returning data for date range
// and GetPriceDateRange works
func (md *IEXProvider) SupportsDateRange() bool {
	return false
}

// GetMarketDataForDateRange returns historical data from tfrom to tto dates.
func (md *IEXProvider) GetMarketDataForDateRange(market *Market, ticker string, tfrom time.Time, tto time.Time) ([]*TimedMarketData, error) {
	return nil, ErrNotAvailable
}

// GetMarketData gets the market price at the specific time.
// market: market identifier (NASDAQ, CURRENCY)
// ticker: stock ticker (APPLE, USDCZK)
func (md *IEXProvider) GetMarketData(market *Market, ticker string, at time.Time) (*MarketData, error) {
	// unfortunatelly it doesn't differentiate between markets, it just uses tickers
	if !md.Supports(market, ticker) {
		return nil, ErrNotAvailable
	}

	wantsMostRecent := time.Since(at) < time.Minute

	url := fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/quote", url.PathEscape(strings.ToLower(ticker)))

	resp, err := md.httpClient.Get(url)
	if err != nil {
		return nil, iexError("http error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, iexError("server returned code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var obj struct {
		LatestPrice      float64 `json:"latestPrice"`
		LatestUpdate     int64   `json:"latestUpdate"`
		LatestUpdateTime time.Time
	}

	err = json.NewDecoder(resp.Body).Decode(&obj)
	if err != nil {
		return nil, iexError("problem decoding response: %v", err)
	}

	obj.LatestUpdateTime = time.Unix(obj.LatestUpdate/1000, 0)

	durationSinceUpdate := at.Sub(obj.LatestUpdateTime)
	if wantsMostRecent || durationSinceUpdate < 30*time.Minute {
		return &MarketData{obj.LatestUpdateTime, obj.LatestPrice, currency.USD /*let's hope we are right*/}, nil
	}

	return nil, ErrNotAvailable
}

// SupportsItemInfo returns true if the provider supports returning info about the item
func (md *IEXProvider) SupportsItemInfo() bool {
	return false
}

// GetItemInfo returns item information
// Parameter market can be empty.
func (md *IEXProvider) GetItemInfo(market *Market, item string) (*ItemInfo, error) {
	return nil, ErrNotAvailable
}

func init() {
	RegisterProvider(NewIEXProvider())
}
