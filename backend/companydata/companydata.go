package companydata

// GetCompanyData returns company data using the first working provider
func GetCompanyData(ticker string) (data CompanyData, err error) {
	//TODO: somehow solve this... different ticker names
	switch ticker {
	case "VOW3":
		ticker = "VOW3.DE"
	}

	for _, p := range Providers {
		if data, err = p.GetCompanyData(ticker); err != nil {
			return
		}
	}
	return
}
