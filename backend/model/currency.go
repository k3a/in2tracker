package model

import "time"

type Currency struct {
	ID   int64  `meddler:"id,pk"`
	Code string `meddler:"code"`
	Name string `meddler:"name"`
}

type CurrencyPair struct {
	Date          time.Time `meddler:"date,localtime"`
	SrcCurrencyID int64     `meddler:"src_currency_id"`
	DstCurrencyID int64     `meddler:"dst_currency_id"`
	Multiplier    float64   `meddler:"multiplier"`
}
