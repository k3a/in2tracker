package companydata

import (
	"testing"

	"github.com/k3a/in2tracker/backend/marketdata"

	"github.com/stretchr/testify/require"
)

func TestYahoo(t *testing.T) {
	p := NewYahooProvider()

	cd, err := p.GetCompanyData(marketdata.MarketUSANasdaq, "AAPL")
	require.Nil(t, err)

	require.Equal(t, "1 Infinite Loop, Cupertino, CA 95014, United States", cd.GetAddress().String())
	require.Equal(t, "Apple Inc.", cd.GetLongName())
}

func TestYahooFindMarket(t *testing.T) {
	p := NewYahooProvider()

	mkt := p.tryFindMarket("LHA")
	if !marketdata.MarketEquals(mkt, marketdata.MarketsEuropeFrankfurtXETRA) {
		t.Fatalf("wrong market %s reported instad of %s", mkt, marketdata.MarketsEuropeFrankfurtXETRA)
	}
}
