package companydata

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"k3a.me/money/backend/model"
)

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

// YahooProvider is yahoo.com provider
type YahooProvider struct {
	httpClient *http.Client
}

// GetCompanyData return info about the company
func (yp *YahooProvider) GetCompanyData(ticker string) (CompanyData, error) {
	u := fmt.Sprintf("https://query1.finance.yahoo.com/v10/finance/quoteSummary/%s"+
		"?formatted=true&lang=en-US&region=US&modules=price,assetProfile&corsDomain=finance.yahoo.com",
		url.QueryEscape(ticker))

	resp, err := yp.httpClient.Get(u)
	if err != nil {
		log.Printf("ERR for %s: %s\n", ticker, err)
		return nil, ErrNotAvailable
	}
	defer resp.Body.Close()

	var respObj yahooData
	err = json.NewDecoder(resp.Body).Decode(&respObj)
	if err != nil {
		log.Printf("ERR for %s: %s\n", ticker, err)
		return nil, ErrBadFormat
	}

	if len(respObj.QuoteSummary.Result) == 0 {
		return nil, ErrNotAvailable
	}

	return &respObj, nil
}

// NewYahooProvider creates a new yahoo.com provider
func NewYahooProvider() Provider {
	return &YahooProvider{
		&http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func init() {
	RegisterProvider(NewYahooProvider())
}
