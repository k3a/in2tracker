package companydata

import (
	"fmt"
	"reflect"

	"github.com/k3a/in2tracker/backend/marketdata"
	"github.com/k3a/in2tracker/backend/model"
)

func e(format string, args ...interface{}) error {
	return fmt.Errorf("companydata: "+format, args...)
}

// Errors
var (
	ErrNotAvailable = e("data not available")
	ErrBadFormat    = e("provider received wrongly formatted data")
)

// CompanyData gives access to various information about a company
type CompanyData interface {
	GetAddress() *model.Address
	GetBusinessSummary() string
	GetIndustry() string
	GetSector() string
	GetLongName() string
	GetMarkets() []*marketdata.Market
}

// Provider is common interface for all company data providers
type Provider interface {
	GetCompanyData(mkt *marketdata.Market, ticker string) (CompanyData, error)
}

// Providers holds all available currency rate providers
var Providers []Provider

// RegisterProvider registers a new provider
func RegisterProvider(provider Provider) {
	for _, p := range Providers {
		if p == provider {
			panic("Attempt to register already-registered provider " +
				reflect.TypeOf(provider).String())
		}
	}
	Providers = append(Providers, provider)
}
