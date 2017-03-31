package marketdata

import (
	"strings"

	"k3a.me/money/backend/currency"
)

// StockMarketCurrency returns currency used in the stock market
func StockMarketCurrency(market string) currency.Currency {
	// from https://www.google.com/intl/en/googlefinance/disclaimer/
	switch strings.ToUpper(market) {
	case "BCBA":
		return currency.ARS
	case "BMV":
		return currency.MXN
	case "BVMF":
		return currency.BRL
	case "CNSX", "CVE", "TSE":
		return currency.CAD
	case "NASDAQ", "NYSE", "NYSEARCA", "NYSEMKT", "OTCBB", "OTCMKTS":
		return currency.USD
	case "AMS", "BIT", "BME", "EBR", "ELI", "EPA", "ETR", "FRA", "HEL", "RSE", "TAL", "VIE", "VSE":
		return currency.EUR
	case "CPH":
		return currency.DKK
	case "ICE":
		return currency.ISK
	case "IST":
		return currency.TRY
	case "LON":
		return currency.GBP
	case "MCX":
		return currency.RUB
	case "STO":
		return currency.SEK
	case "SWX", "VTX":
		return currency.CHF
	case "WSE":
		return currency.PLN
	case "JSE":
		return currency.ZAC
	case "TADAWUL":
		return currency.SAR
	case "TLV":
		return currency.ILA
	case "BKK":
		return currency.THB
	case "BOM", "NSE":
		return currency.INR
	case "KLSE":
		return currency.MYR
	case "HKG":
		return currency.HKD
	case "IDX":
		return currency.IDR
	case "KOSDAQ", "KRX":
		return currency.KRW
	case "SGX":
		return currency.SGD
	case "SHA", "SHE":
		return currency.CNY
	case "TPE":
		return currency.TWD
	case "TYO":
		return currency.JPY
	case "ASX":
		return currency.AUD
	case "NZE":
		return currency.NZD
	default:
		return currency.Invalid
	}
}
