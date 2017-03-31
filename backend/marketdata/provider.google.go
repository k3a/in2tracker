package marketdata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"strings"

	"k3a.me/money/backend/utils"
)

type gfinanceFindResult struct {
	Ticker   string `json:"t"`
	Name     string `json:"n"`
	Exchange string `json:"e"`
	ID       string `json:"id"`
}

func gfinanceFind(query string) ([]gfinanceFindResult, error) {
	httpCli := &http.Client{Timeout: 30 * time.Second}

	url := fmt.Sprintf("https://www.google.com/finance/match?matchtype=matchall&q=%s",
		url.QueryEscape(query))

	resp, err := httpCli.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, e("server returned code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var resObj struct {
		Matches []gfinanceFindResult `json:"matches"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&resObj); err != nil {
		return nil, err
	}

	return resObj.Matches, nil
}

func gfinanceFindOne(market string, ticker string) (*gfinanceFindResult, error) {
	query := ticker
	if len(market) > 0 && len(ticker) > 0 {
		query = fmt.Sprintf("%s:%s", market, ticker)
	}

	res, err := gfinanceFind(query)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 || !strings.EqualFold(ticker, res[0].Ticker) {
		return nil, e("ticker %s not found", ticker)
	}
	return &res[0], err
}

type gfinanceTimeString struct {
	time.Time
}

func (ts *gfinanceTimeString) UnmarshalJSON(inp []byte) (err error) {
	if len(inp) <= 2 {
		ts.Time = time.Time{}
		return
	}

	t, err := time.Parse("Jan 2, 3:04PM MST", string(inp[1:len(inp)-1]))
	year := time.Now().In(t.Location()).Year()
	ts.Time = time.Date(year, t.Month(), t.Day(), t.Hour(),
		t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	return err
}

type gfinanceCSVDate struct {
	time.Time
}

func (ts *gfinanceCSVDate) UnmarshalCSV(str string) (err error) {
	ts.Time, err = time.Parse(`2-Jan-06`, str) // ok to be in UTC
	return
}

// GoogleProvider represends Google market data source
type GoogleProvider struct {
	httpClient *http.Client
}

// NewGoogleProvider creates a new data fetcher
func NewGoogleProvider() *GoogleProvider {
	return &GoogleProvider{
		&http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the name of the provider
func (md *GoogleProvider) Name() string {
	return "Google"
}

// Supports returns true if the item-market pair is supported by the provider.
// Should return fast and not make any http requests (except for the first time it is called)
func (md *GoogleProvider) Supports(market string, ticker string) bool {
	return true // supports most of the markets
}

// SupportsDateRange returns true if the provider supports returning data for date range
// and GetPriceDateRange works
func (md *GoogleProvider) SupportsDateRange() bool {
	return true
}

// GetMarketDataForDateRange returns historical data from tfrom to tto dates.
func (md *GoogleProvider) GetMarketDataForDateRange(market string, ticker string, tfrom time.Time, tto time.Time) ([]*TimedMarketData, error) {
	tickerInfo, err := gfinanceFindOne(market, ticker)
	if err != nil {
		return nil, err
	}

	timeParamFrom := url.QueryEscape(tfrom.Format("Jan 2 2006"))
	timeParamTo := url.QueryEscape(tto.Format("Jan 2 2006"))

	url := fmt.Sprintf("http://www.google.com/finance/historical?cid=%s&startdate=%s&enddate=%s&num=1&output=csv",
		tickerInfo.ID, timeParamFrom, timeParamTo)

	resp, err := md.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, e("server returned code %d for date range fetch", resp.StatusCode)
	}
	defer resp.Body.Close()

	var rows []*struct {
		Date   gfinanceCSVDate     `csv:"Date"`
		Open   utils.Float64String `csv:"Open"`
		High   utils.Float64String `csv:"High"`
		Low    utils.Float64String `csv:"Low"`
		Close  utils.Float64String `csv:"Close"`
		Volume utils.Float64String `csv:"Volume"`
	}

	// skype 3byte BOM
	bom := make([]byte, 3)
	resp.Body.Read(bom)

	rd := utils.NewCSVReader(resp.Body)
	rd.Comma = ','
	if err := rd.Unmarshal(&rows); err != nil {
		return nil, err
	}

	// safe check
	if len(rows) == 0 {
		return nil, ErrNotAvailable
	}

	var outArr []*TimedMarketData
	currency := StockMarketCurrency(tickerInfo.Exchange)
	for _, v := range rows {
		outArr = append(outArr, &TimedMarketData{
			Time:     v.Date.Time,
			Open:     v.Open.Float64,
			Low:      v.Low.Float64,
			High:     v.High.Float64,
			Close:    v.Close.Float64,
			Volume:   v.Volume.Float64,
			Currency: currency})
	}

	return outArr, nil
}

func (md *GoogleProvider) gfinanceGetRealtime(market string, ticker string, at time.Time) (*MarketData, error) {
	url := fmt.Sprintf("http://www.google.com/finance/info?q=%s:%s",
		url.QueryEscape(market), url.QueryEscape(ticker))

	resp, err := md.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, e("server returned code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var obj []struct {
		ID    string              `json:"id"`
		Price utils.Float64String `json:"l"`
		Time  gfinanceTimeString  `json:"lt"`
	}

	// read the garbage at the beginning
	garbage := make([]byte, 3)
	resp.Body.Read(garbage)

	err = json.NewDecoder(resp.Body).Decode(&obj)
	if err != nil {
		fmt.Println(err)
		return nil, ErrBadFormat
	}

	// safe check
	if len(obj) == 0 || time.Since(obj[0].Time.Time) > 20*time.Minute {
		return nil, ErrNotAvailable
	}

	return &MarketData{
		LastTrade: obj[0].Price.Float64,
		Currency:  StockMarketCurrency(market),
	}, nil
}

// GetMarketData gets the market price at the specific time.
// market: market identifier (NASDAQ, CURRENCY)
// ticker: stock ticker (APPLE, USDCZK)
func (md *GoogleProvider) GetMarketData(market string, ticker string, at time.Time) (*MarketData, error) {
	timeDiff := time.Since(at)
	if timeDiff < 0 /*future*/ {
		return nil, ErrNotAvailable
	}
	if timeDiff > 20*time.Minute {
		prices, err := md.GetMarketDataForDateRange(market, ticker, at, at)
		if err != nil {
			return nil, err
		}

		// safe check
		timeDiff := prices[0].Time.Sub(at)
		if timeDiff > 5*24*time.Hour {
			return nil, ErrNotAvailable
		}

		return &MarketData{
			LastTrade: (prices[0].High-prices[0].Low)/2.0 + prices[0].Low,
			Currency:  prices[0].Currency,
		}, nil
	}

	return md.gfinanceGetRealtime(market, ticker, at)
}

// SupportsItemInfo returns true if the provider supports returning info about the item
func (md *GoogleProvider) SupportsItemInfo() bool {
	return true
}

// GetItemInfo returns item information
// Parameter market can be empty.
func (md *GoogleProvider) GetItemInfo(market string, item string) (*ItemInfo, error) {
	if info, err := gfinanceFindOne(market, item); err == nil {
		return &ItemInfo{Name: info.Name, Market: info.Exchange}, nil
	}
	return nil, ErrNotAvailable
}

func init() {
	RegisterProvider(NewGoogleProvider())
}
