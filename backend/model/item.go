package model

// Item hold item or country info
type Item struct {
	ID         int64  `meddler:"id,pk"`
	MarketID   int64  `meddler:"market_id"`
	CountryID  int64  `meddler:"country_id"`
	CurrencyID int64  `meddler:"currency_id"`
	Code       string `meddler:"code"`
	Name       string `meddler:"name"`
	Address    string `meddler:"address"`
}
