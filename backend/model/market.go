package model

type Market struct {
	ID                int64  `meddler:"id,pk"`
	Name              string `meddler:"name"`
	DefaultCurrencyID int64  `meddler:"default_currency_id"`
	DefaultCountryID  int64  `meddler:"default_country_id"`
}
