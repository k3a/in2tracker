package companydata

// GetCompanyData returns company data using the first working provider
func GetCompanyData(ticker string) (data CompanyData, err error) {
	for _, p := range Providers {
		if data, err = p.GetCompanyData(ticker); err != nil {
			return
		}
	}
	return
}
