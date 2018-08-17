package companydata

import "github.com/k3a/in2tracker/backend/marketdata"

// GetCompanyData returns company data using the first working provider
func GetCompanyData(mkt *marketdata.Market, ticker string) (data CompanyData, err error) {
	for _, p := range Providers {
		if data, err = p.GetCompanyData(mkt, ticker); err != nil {
			return
		}
	}
	return
}
