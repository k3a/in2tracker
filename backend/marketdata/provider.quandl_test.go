package marketdata

import (
	"testing"
	"time"

	"github.com/k3a/in2tracker/backend/currency"
)

func TestQuandlProvider(t *testing.T) {
	provider := NewQuandlProvider("QUANDL_SECRET" /*SECRET*/)
	md, err := provider.GetMarketData("NASDAQ", "AAPL", time.Date(2017, 1, 10, 19, 00, 00, 00, time.UTC))

	if err != nil {
		t.Fatal(err)
	}

	if md.Currency != currency.USD {
		t.Fatal("wrong currency")
	}

	if md.LastTrade < 2 {
		t.Fatal("AAPL too cheap, provider probably doesn't work")
	}
}
