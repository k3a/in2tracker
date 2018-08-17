package companydata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/k3a/in2tracker/backend/marketdata"
	"github.com/k3a/in2tracker/backend/model"
	"github.com/k3a/in2tracker/backend/utils"
	"github.com/lunny/log"
)

const yahooReceiverKey = "yahoo"

type yahooNumber struct {
	Raw float64 `json:"raw"`
	Fmt string  `json:"fmt"`
}

type yahooData struct {
	QuoteSummary struct {
		Result []struct {
			Price struct {
				ShortName    string `json:"shortName"`
				LongName     string `json:"longName"`
				Currency     string `json:"currency"`
				Exchange     string `json:"exchange"`
				ExchangeName string `json:"exchangeName"`

				PreMarketTime          int         `json:"preMarketTime"`
				PreMarketPrice         yahooNumber `json:"preMarketPrice"`
				PreMarketChange        yahooNumber `json:"preMarketChange"`
				PreMarketChangePercent yahooNumber `json:"preMarketChangePercent"`

				PostMarketTime          int         `json:"postMarketTime"`
				PostMarketPrice         yahooNumber `json:"postMarketPrice"`
				PostMarketChange        yahooNumber `json:"postMarketChange"`
				PostMarketChangePercent yahooNumber `json:"postMarketChangePercent"`

				RegularMarketTime          int         `json:"regularMarketTime"`
				RegularMarketPrice         yahooNumber `json:"regularMarketPrice"`
				RegularMarketChange        yahooNumber `json:"regularMarketChange"`
				RegularMarketChangePercent yahooNumber `json:"regularMarketChangePercent"`

				RegularMarketDayLow  yahooNumber `json:"regularMarketDayLow"`
				RegularMarketDayHigh yahooNumber `json:"regularMarketDayHigh"`
			} `json:"price"`

			AssetProfile struct {
				Address1            string `json:"address1"`
				City                string `json:"city"`
				State               string `json:"state"`
				Zip                 string `json:"zip"`
				Country             string `json:"country"`
				Phone               string `json:"phone"`
				Website             string `json:"website"`
				Industry            string `json:"industry"`
				IndustrySymbol      string `json:"industrySymbol"`
				Sector              string `json:"sector"`
				LongBusinessSummary string `json:"longBusinessSummary"`
				FullTimeEmployees   int    `json:"fullTimeEmployees"`
				CompanyOfficers     []struct {
					Name       string `json:"name"`
					Age        int    `json:"age"`
					Title      string `json:"title"`
					FiscalYear int    `json:"fiscalYear"`
					TotalPay   struct {
						Raw     int    `json:"raw"`
						Fmt     string `json:"fmt"`
						LongFmt string `json:"longFmt"`
					} `json:"totalPay"`
				} `json:"companyOfficers"`
			} `json:"assetProfile"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func (yd *yahooData) GetAddress() *model.Address {
	return &model.Address{
		Address: yd.QuoteSummary.Result[0].AssetProfile.Address1,
		City:    yd.QuoteSummary.Result[0].AssetProfile.City,
		State:   yd.QuoteSummary.Result[0].AssetProfile.State,
		Zip:     yd.QuoteSummary.Result[0].AssetProfile.Zip,
		Country: yd.QuoteSummary.Result[0].AssetProfile.Country,
	}
}
func (yd *yahooData) GetBusinessSummary() string {
	return yd.QuoteSummary.Result[0].AssetProfile.LongBusinessSummary
}
func (yd *yahooData) GetIndustry() string {
	return yd.QuoteSummary.Result[0].AssetProfile.Industry
}
func (yd *yahooData) GetSector() string {
	return yd.QuoteSummary.Result[0].AssetProfile.Sector
}
func (yd *yahooData) GetLongName() string {
	return yd.QuoteSummary.Result[0].Price.LongName
}
func (yd *yahooData) GetMarkets() []*marketdata.Market {
	// TODO: get markets the company is listed at
	return nil
}

// YahooProvider is yahoo.com provider
type YahooProvider struct {
	httpClient *http.Client
}

// tryFindMarket tries to find the most relevant market for ticker or nil if not known by Yahoo
func (yp *YahooProvider) tryFindMarket(ticker string) *marketdata.Market {
	switch ticker {
	case "VOW3":
		return marketdata.MarketsEuropeFrankfurtXETRA
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v1/finance/search"+
		"?q=%s&quotesCount=2&newsCount=0&quotesQueryId=tss_match_phrase_query&multiQuoteQueryId=multi_quote_single_token_query&newsQueryId=news_ss_symbols&enableCb=true",
		url.QueryEscape(ticker))

	req, err := utils.NewBrowserRequest("GET", url, nil)
	if err != nil {
		return nil
	}

	resp, err := yp.httpClient.Do(req)
	if err != nil {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		log.Warnf("yahoo: unable to find best market for %s", ticker)
		return nil
	}
	defer resp.Body.Close()

	var obj struct {
		Quotes []struct {
			Exchange string  `json:"exchange"`
			Score    float64 `json:"score"`
			Symbol   string  `json:"symbol"`
		} `json:"quotes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&obj); err != nil {
		log.Warnf("yahoo: unable to decode best market resp: %v\n", err)
		return nil
	}

	if len(obj.Quotes) > 0 {
		symbolMktPair := strings.Split(obj.Quotes[0].Symbol, ".")
		if len(symbolMktPair) > 1 {
			return marketdata.MarketFromStringForReceiver(symbolMktPair[1], yahooReceiverKey)
		}
	}

	return nil
}

// GetCompanyData return info about the company
func (yp *YahooProvider) GetCompanyData(market *marketdata.Market, ticker string) (CompanyData, error) {
	if marketdata.MarketEquals(market, marketdata.MarketAny) {
		market = yp.tryFindMarket(ticker)
	}

	if !marketdata.MarketEquals(market, marketdata.MarketAny) {
		mktIdent := market.IdentifierForReceiver(yahooReceiverKey)
		if len(mktIdent) > 0 {
			ticker = ticker + "." + mktIdent
		}
	}

	u := fmt.Sprintf("https://query1.finance.yahoo.com/v10/finance/quoteSummary/%s"+
		"?formatted=true&lang=en-US&region=US&modules=price,assetProfile&corsDomain=finance.yahoo.com",
		url.QueryEscape(ticker))

	req, err := utils.NewBrowserRequest("GET", u, nil)
	if err != nil {
		log.Warnf("yahoo: problem creating request for %s: %v\n", ticker, err)
		return nil, ErrNotAvailable
	}

	resp, err := yp.httpClient.Do(req)
	if err != nil {
		log.Warnf("yahoo: problem fetching data for %s: %v\n", ticker, err)
		return nil, ErrNotAvailable
	}
	defer resp.Body.Close()

	var respObj yahooData
	err = json.NewDecoder(resp.Body).Decode(&respObj)
	if err != nil {
		log.Warnf("yahoo: problem decoding response for %s: %v\n", ticker, err)
		return nil, ErrBadFormat
	}

	if len(respObj.QuoteSummary.Result) == 0 {
		return nil, ErrNotAvailable
	}

	return &respObj, nil
}

// NewYahooProvider creates a new yahoo.com provider
func NewYahooProvider() *YahooProvider {
	// register market aliases
	marketdata.MarketUSANYSE.AssignIdentifierForReceiver("", yahooReceiverKey)
	marketdata.MarketUSANYSEArca.AssignIdentifierForReceiver("", yahooReceiverKey)
	marketdata.MarketUSANasdaq.AssignIdentifierForReceiver("", yahooReceiverKey)
	marketdata.MarketsEuropeFrankfurtBoerse.AssignIdentifierForReceiver("F", yahooReceiverKey)
	marketdata.MarketsEuropeFrankfurtXETRA.AssignIdentifierForReceiver("DE", yahooReceiverKey)
	marketdata.MarketsEuropeLSE.AssignIdentifierForReceiver("L", yahooReceiverKey)

	return &YahooProvider{
		&http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func init() {
	RegisterProvider(NewYahooProvider())
}
