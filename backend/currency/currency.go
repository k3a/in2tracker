package currency

// Currency specifies a currency type
// Types are defined in this package as ISO 4217 identifiers
type Currency string

// ISO 4217 codes
const (
	Invalid = Currency("N/A")
	AUD     = Currency("AUD")
	BRL     = Currency("BRL")
	CZK     = Currency("CZK")
	GBN     = Currency("GBN")
	CNY     = Currency("CNY")
	DKK     = Currency("DKK")
	EUR     = Currency("EUR")
	PHP     = Currency("PHP")
	HKD     = Currency("HKD")
	HRK     = Currency("HRK")
	INR     = Currency("INR")
	IDR     = Currency("IDR")
	ILS     = Currency("ILS")
	JPY     = Currency("JPY")
	ZAR     = Currency("ZAR")
	KRW     = Currency("KRW")
	CAD     = Currency("CAD")
	HUF     = Currency("HUF")
	MYR     = Currency("MYR")
	MXN     = Currency("MXN")
	XDR     = Currency("XDR")
	NOK     = Currency("NOK")
	NZD     = Currency("NZD")
	PLN     = Currency("PLN")
	RON     = Currency("RON")
	RUB     = Currency("RUB")
	SGD     = Currency("SGD")
	SEK     = Currency("SEK")
	CHF     = Currency("CHF")
	THB     = Currency("THB")
	TRY     = Currency("TRY")
	USD     = Currency("USD")
	GBP     = Currency("GBP")
	ARS     = Currency("ARS")
	ISK     = Currency("ISK")
	ZAC     = Currency("ZAC")
	SAR     = Currency("SAR")
	ILA     = Currency("ILA")
	TWD     = Currency("TWD")
)

var currencyNameMap = map[Currency]string{
	AUD: "Australian dollar",
	BRL: "Brazilian real",
	CZK: "Czech koruna",
	EUR: "Euro",
	JPY: "Japanese yen",
	CAD: "Canadian dollar",
	PLN: "Polish złoty",
	RUB: "Russian ruble",
	CHF: "Swiss franc",
	USD: "United States dollar",
	GBP: "Pound sterling",
}

func (c Currency) String() string {
	return string(c)
}

// Name returns human-readable currency name if known
func (c Currency) Name() string {
	if name, ok := currencyNameMap[c]; ok {
		return name
	}
	return c.String()
}

// FromString returns a Currency from its string identiier
func FromString(currencyIdent string) Currency {
	return Currency(currencyIdent)
}

// UnmarshalJSON unmarshals the currency from JSON
func (c *Currency) UnmarshalJSON(inp []byte) (err error) {
	if len(inp) < 2 {
		return e("cannot convert %s to currency type", string(inp))
	}
	return c.UnmarshalCSV(string(inp[1 : len(inp)-1]))
}

// UnmarshalCSV unmarshals the currency from CSV
func (c *Currency) UnmarshalCSV(csv string) (err error) {
	*c = FromString(csv)
	return
}
