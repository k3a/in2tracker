package marketdata

import (
	"testing"

	"time"

	"github.com/k3a/in2tracker/backend/currency"
)

func TestGoogleStockFetchRealtime(t *testing.T) {
	provider := NewGoogleProvider()
	md, err := provider.GetMarketData("NASDAQ", "AAPL", time.Now())

	if err == ErrNotAvailable {
		t.Skip("realtime data unavalable, must skip")
	}

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

func TestGoogleStockFetchHistorical(t *testing.T) {
	provider := NewGoogleProvider()
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

func TestGoogleStockFind(t *testing.T) {
	objs, err := gfinanceFind("AAPL")

	if err != nil {
		t.Fatal(err)
	}

	if len(objs) == 0 {
		t.Fatal("search doesn't work")
	}

	if objs[0].Ticker != "AAPL" {
		t.Fatal("search didn't found AAPL")
	}
}
