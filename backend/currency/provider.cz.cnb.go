package currency

import (
	"fmt"
	"net/http"
	"time"

	"strings"

	"github.com/k3a/in2tracker/backend/utils"
)

// CZCNB receives data from
type CZCNB struct {
	allowedSrcCurrencies []Currency
	httpClient           *http.Client
}

// NewCZCNB creates a new cnb.cz currency rates provider
func NewCZCNB() Provider {
	return &CZCNB{
		nil,
		&http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the name of the provider
func (c *CZCNB) Name() string {
	return "CNB.cz"
}

// AllowsReverse specifies whether the provider allows reversing rates
// (e.g. for EURUSD use USDEUR)
func (c *CZCNB) AllowsReverse() bool {
	return true // rates are "middle"
}

// Supports checks whether the providet supports the currency conversion
func (c *CZCNB) Supports(from Currency, to Currency) bool {
	if to != CZK {
		return false
	}

	if c.allowedSrcCurrencies == nil {
		// make at least one request to get the list first
		_, err := c.GetRate(USD, CZK, time.Now().AddDate(0, 0, -1))
		if err != nil {
			fmt.Printf("Unable to fetch supported currencies by cnb.cz! %s\n",
				err.Error())
			return false
		}
	}

	for _, currency := range c.allowedSrcCurrencies {
		if currency == from {
			return true
		}
	}

	return false
}

// GetRate gets the currency rate for the specified time. Returns ErrNotAvailable error
// if the conversion rate for the specified time is not known
func (c *CZCNB) GetRate(from Currency, to Currency, at time.Time) (float64, error) {
	if to != CZK {
		return 0, ErrNotAvailable
	}

	url := fmt.Sprintf("http://www.cnb.cz/cs/financni_trhy/devizovy_trh/kurzy_devizoveho_trhu/"+
		"denni_kurz.txt?date=%02d.%02d.%04d", at.Day(), at.Month(), at.Year())

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, e("cnb.cz server returned code %d", resp.StatusCode)
	}

	csv := utils.NewCSVReader(resp.Body)
	csv.Comma = '|'

	var rows []struct {
		Country      string                `csv:"země"`
		CurrencyText string                `csv:"měna"`
		Amount       utils.CZFloat64String `csv:"množství"`
		Currency     Currency              `csv:"kód"`
		Rate         utils.CZFloat64String `csv:"kurz"`
	}

	err = csv.Unmarshal(&rows)
	if err != nil {
		return 0, err
	}

	// check the garbage - parse the date
	headerParts := strings.Split(csv.GarbageString(), " ")
	if len(headerParts) < 2 {
		// wrong format (should be similar to "06.01.2017 #5")
		return 0, ErrBadFormat
	}
	issued, err := time.Parse("02.01.2006", headerParts[0])
	if err != nil {
		// unable to parse isssue date from the header
		return 0, ErrBadFormat
	}
	if at.Sub(issued) > 5*24*time.Hour {
		return 0, ErrOldData
	}

	// allowed currencies
	if c.allowedSrcCurrencies == nil {
		for _, r := range rows {
			c.allowedSrcCurrencies = append(c.allowedSrcCurrencies, r.Currency)
		}
	}

	// try to find a matching rate
	found := false
	rate := 0.0
	for _, r := range rows {
		if c.allowedSrcCurrencies == nil {
			c.allowedSrcCurrencies = append(c.allowedSrcCurrencies, r.Currency)
		}
		if !found && r.Currency == from {
			rate = r.Rate.Float64 / r.Amount.Float64
			found = true
		}
	}

	if !found {
		return 0, ErrNotAvailable
	}

	return rate, nil
}

func init() {
	RegisterProvider(NewCZCNB())
}
