package marketdata

import (
	"testing"
	"time"
)

func TestIEX(t *testing.T) {
	mp := NewIEXProvider()

	tm := time.Now()

	md, err := mp.GetMarketData(MarketFromString("NASDAQ"), "AAPL", tm)
	if err != nil {
		t.Fatal(err)
	}
	if md == nil {
		t.Fatal("market data nil")
	}

	md, err = mp.GetMarketData(MarketFromString("NYSE"), "T", tm)
	if err != nil {
		t.Fatal(err)
	}
	if md == nil {
		t.Fatal("market data nil")
	}

}
