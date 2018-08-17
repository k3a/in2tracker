package marketdata

import (
	"net/http"
	"net/url"
	"time"

	"fmt"

	"encoding/json"

	"github.com/k3a/in2tracker/backend/currency"
)

func qandlErr(format string, args ...interface{}) error {
	return fmt.Errorf("quandl: "+format, args...)
}

// QuandlProvider provides data from quandl.com
type QuandlProvider struct {
	httpCli http.Client
	apiKey  string
}

// NewQuandlProvider returns the new provider
func NewQuandlProvider(apiKey string) *QuandlProvider {
	return &QuandlProvider{http.Client{Timeout: 30 * time.Second}, apiKey}
}

// Name returns the name of the provider
func (p *QuandlProvider) Name() string {
	return "Quandl"
}

// Supports returns true if the item-market pair is supported by the provider.
// Should return fast and not make any http requests (except for the first time it is called)
// Parameter market can be empty.
func (p *QuandlProvider) Supports(market *Market, item string) bool {
	// no way to easily get the list of supported items, it seems :(
	// More than 3000 companies means USD markets may be in..
	return true
}

// GetMarketData gets the market price at the specific time.
// market: market identifier (NASDAQ, CURRENCY)
// item: stock ticker or item identifier (APPLE, USDCZK)
func (p *QuandlProvider) GetMarketData(market *Market, item string, at time.Time) (*MarketData, error) {
	if time.Since(at) < 12 {
		// we have EOD data only
		return nil, ErrNotAvailable
	}

	prices, err := p.GetMarketDataForDateRange(market, item, at, at)
	if err != nil {
		return nil, err
	}
	if len(prices) == 0 {
		return nil, ErrNotAvailable
	}

	return &MarketData{
		LastTrade: (prices[0].High-prices[0].Low)/2.0 + prices[0].Low,
		Currency:  prices[0].Currency,
	}, nil
}

// SupportsDateRange returns true if the provider supports returning data for date range
// and GetPriceDateRange works
func (p *QuandlProvider) SupportsDateRange() bool {
	return true
}

// GetMarketDataForDateRange returns historical data from tfrom to tto dates.
func (p *QuandlProvider) GetMarketDataForDateRange(market *Market, item string, tfrom time.Time, tto time.Time) ([]*TimedMarketData, error) {

	fromDateStr := url.QueryEscape(tfrom.Format("20060102"))
	toDateStr := url.QueryEscape(tto.Format("20060102"))

	u := fmt.Sprintf("https://www.quandl.com/api/v3/datatables/WIKI/PRICES.json"+
		"?ticker=%s&qopts.columns=date,open,low,high,close,volume&api_key=%s"+
		"&date.gte=%s&date.lte=%s",
		url.QueryEscape(item), url.QueryEscape(p.apiKey), fromDateStr, toDateStr)

	resp, err := p.httpCli.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		DataTable struct {
			Data [][]interface{}
		} `json:"datatable"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println(err.Error())
		return nil, qandlErr("problem decoding range data: %v", err)
	}

	var outArr []*TimedMarketData
	for _, tp := range data.DataTable.Data {
		if len(tp) != 6 /* must have 6 fields in the response */ {
			continue
		}

		t, err := time.Parse("2006-01-02", tp[0].(string)) // ok to be in UTC
		if err != nil {
			fmt.Println(err.Error())
			return nil, qandlErr("problem parsing date: %v", err)
		}

		outArr = append(outArr, &TimedMarketData{
			Time:     t,
			Open:     tp[1].(float64),
			Low:      tp[2].(float64),
			High:     tp[3].(float64),
			Close:    tp[4].(float64),
			Volume:   tp[5].(float64),
			Currency: currency.USD,
		})
	}

	return outArr, nil
}

// SupportsItemInfo returns true if the provider supports returning info about the item
func (p *QuandlProvider) SupportsItemInfo() bool {
	return false
}

// GetItemInfo returns item information
// Parameter market can be empty.
func (p *QuandlProvider) GetItemInfo(market *Market, item string) (*ItemInfo, error) {
	return nil, ErrNotAvailable
}

func init() {
	// getting deprecated https://www.quandl.com/databases/WIKIP
	RegisterProvider(NewQuandlProvider("xgafvQ_VZLuFbT7yDwxW" /*SECRET: */))
}
