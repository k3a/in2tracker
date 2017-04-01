package companydata

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYahoo(t *testing.T) {
	p := NewYahooProvider()

	cd, err := p.GetCompanyData("AAPL")
	require.Nil(t, err)

	require.Equal(t, "1 Infinite Loop, Cupertino, CA 95014, United States", cd.GetAddress().String())
	require.Equal(t, "Apple Inc.", cd.GetLongName())
}
