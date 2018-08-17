package importers

import (
	"io"

	"regexp"

	"strings"

	"time"

	"github.com/k3a/in2tracker/backend/currency"
	"github.com/k3a/in2tracker/backend/utils"
	"golang.org/x/text/encoding/charmap"
)

type CZFioImporter struct {
	// mapping colum names to column array index
	colNameToIndex map[string]int
	// mapping volumes for individual currencies to column array index
	volumeCols map[currency.Currency]int
	// mapping fees for individual currencies to column array index
	feeCols map[currency.Currency]int
}

// NewCZFioImporter creates a fio.cz transaction importer
func NewCZFioImporter() *CZFioImporter {
	return &CZFioImporter{
		make(map[string]int),
		make(map[currency.Currency]int),
		make(map[currency.Currency]int),
	}
}

func (imp *CZFioImporter) validateColumns(reqCols []string) error {
	for _, rc := range reqCols {
		if _, has := imp.colNameToIndex[rc]; !has {
			return e("missing column %s", rc)
		}
	}
	return nil
}

// Name returns a name of the importer
func (imp *CZFioImporter) Name() string {
	return "fio.cz"
}

var reDividendText = regexp.MustCompile(`(?i)divid\.|dividenda|Korekce výnosu|Stock Dividend Cash Distribution|Refundable U.S. Fed Tax`)
var reFeeText = regexp.MustCompile(`(?i)fee|poplatek`)
var reDeposit = regexp.MustCompile(`(?i)Vloženo na účet|Převod z účtu`)
var reWithdrawal = regexp.MustCompile(`(?i)Vybráno z|Převod na účet`)

// Import parses data from the reader and returns transactions
func (imp *CZFioImporter) Import(reader io.Reader) ([]*Transaction, error) {

	csvrd := utils.NewCSVReaderWithEncoding(reader, charmap.Windows1250)
	csvrd.Comma = ';'

	// parse columns
	columns, err := csvrd.GoCSVReader().Read()
	if err != nil {
		return nil, err
	}

	// process columns
	reVolume := regexp.MustCompile(`Objem v (\S+)`)
	reFee := regexp.MustCompile(`Poplatky v (\S+)`)
	for i, cname := range columns {
		imp.colNameToIndex[cname] = i

		if arr := reVolume.FindStringSubmatch(cname); len(arr) == 2 {
			imp.volumeCols[currency.FromString(arr[1])] = i
		} else if arr := reFee.FindStringSubmatch(cname); len(arr) == 2 {
			imp.feeCols[currency.FromString(arr[1])] = i
		}
	}

	// ensure we have all the important columns
	err = imp.validateColumns([]string{"Datum obchodu", "Směr", "Symbol", "Cena", "Počet", "Měna", "Text FIO"})
	if err != nil {
		return nil, err
	}

	// process rows
	var outArr []*Transaction
	var row []string
	lineNum := 1
	for {
		lineNum++
		row, err = csvrd.GoCSVReader().Read()
		if err != nil {
			break
		}

		newTransaction := &Transaction{}

		// direction
		trDir := strings.TrimSpace(row[imp.colNameToIndex["Směr"]])

		// item
		newTransaction.Item = strings.TrimSpace(row[imp.colNameToIndex["Symbol"]])
		if newTransaction.Item == "Součet" {
			// skip the sum line
			continue
		}

		// text
		newTransaction.Reference = strings.TrimSpace(row[imp.colNameToIndex["Text FIO"]])
		if strings.Contains(newTransaction.Reference, "volitelné dividendy") {
			// skip "distribuce prav volitelne dividendy" line
			//TODO: verify it's ok
			continue
		}

		// quantity
		field := strings.TrimSpace(row[imp.colNameToIndex["Počet"]])
		newTransaction.Quantity, err = utils.ParseCZFloat(field)
		if err != nil {
			return nil, e("unable to parse Počet on line %d", lineNum)
		}

		// price
		field = strings.TrimSpace(row[imp.colNameToIndex["Cena"]])
		newTransaction.Price, err = utils.ParseCZFloat(field)
		if err != nil {
			return nil, e("unable to parse Cena on line %d", lineNum)
		}

		// date/time
		field = strings.TrimSpace(row[imp.colNameToIndex["Datum obchodu"]])
		timeLoc, err := time.LoadLocation("Europe/Prague")
		if err != nil {
			return nil, err
		}
		newTransaction.Time, err = time.ParseInLocation("02.01.2006 15:04", field, timeLoc)
		if err != nil {
			return nil, e("unable to parse date string %s on line %d", field, lineNum)
		}

		// fee and fee currency
		newTransaction.FeeCurrency = currency.Invalid
		for c, i := range imp.feeCols {
			field = strings.TrimSpace(row[i])
			if len(field) > 0 {
				newTransaction.Fee, err = utils.ParseCZFloat(field)
				if err != nil {
					return nil, e("unable to parse fee on line %d", len(outArr)+2)
				}
				newTransaction.FeeCurrency = c
				break
			}
		}

		// net total
		trNetTotalCurrency := currency.Invalid
		for c, i := range imp.volumeCols {
			field = strings.TrimSpace(row[i])
			if len(field) > 0 {
				newTransaction.NetTotal, err = utils.ParseCZFloat(field)
				if err != nil {
					return nil, e("unable to parse net total on line %d", len(outArr)+2)
				}
				trNetTotalCurrency = c
				break
			}
		}

		// transaction type decision
		newTransaction.Type = TTInvalid // invalid by default
		if newTransaction.Reference == "Nákup" || trDir == "Nákup" {
			// purchasing currency for other currency or purchasing stock
			newTransaction.Type = TTBuy
		} else if newTransaction.Reference == "Prodej" || trDir == "Prodej" {
			// selling currency for other currency or selling stock
			newTransaction.Type = TTSell
		}
		if newTransaction.Type == TTInvalid && len(newTransaction.Item) == 0 && newTransaction.Price == 0 {
			// type still not set and no item nor price specified
			if reWithdrawal.MatchString(newTransaction.Reference) && newTransaction.NetTotal <= 0 {
				newTransaction.Type = TTWithdrawal
			} else if reDeposit.MatchString(newTransaction.Reference) && newTransaction.NetTotal >= 0 {
				newTransaction.Type = TTDeposit
			} else if strings.Contains(newTransaction.Reference, "Poplatek za převod na OU") {
				newTransaction.Type = TTDeposit
			}
		}
		if newTransaction.Type == TTInvalid && len(trDir) == 0 {
			// type still not set and no direction specified
			if strings.Contains(newTransaction.Reference, "CAPITAL GAIN") {
				// capital gain? Like a dividend?
				newTransaction.Type = TTDividend
			} else if reDividendText.MatchString(newTransaction.Reference) {
				// text matches dividend
				newTransaction.Type = TTDividend
			} else if reFeeText.MatchString(newTransaction.Reference) {
				// text matches fee regexp
				newTransaction.Type = TTFee
				newTransaction.Quantity = 0
				newTransaction.Price = 0
				newTransaction.Fee = -newTransaction.NetTotal
			}
		}
		if strings.Contains(newTransaction.Reference, "Spin-off") {
			newTransaction.Type = TTBuy
			newTransaction.Price = 0
			newTransaction.NetTotal = 0
			newTransaction.Currency = currency.Invalid
			newTransaction.FeeCurrency = currency.Invalid
		} else if strings.Contains(newTransaction.Reference, "Return of Principal") {
			newTransaction.Type = TTReturnOfCapital
		} else if newTransaction.Type == TTInvalid &&
			strings.Contains(newTransaction.Reference, "Merger") {
			newTransaction.Type = TTMergerCash
		}
		// still invalid? signal failture!
		if newTransaction.Type == TTInvalid {
			return nil, e("unknown transaction type on line %d: %#v", lineNum, newTransaction)
		}

		// transaction currency
		field = strings.TrimSpace(row[imp.colNameToIndex["Měna"]])
		if len(field) == 0 {
			// currency not specified, use net total currency
			newTransaction.Currency = trNetTotalCurrency
		} else {
			newTransaction.Currency = currency.FromString(field)
			if newTransaction.Currency == currency.Invalid {
				return nil, e("unknown currency %s on line %d", field, lineNum)
			}
		}

		// last-chance fixes
		if newTransaction.Type == TTDividend {
			newTransaction.Quantity = 0
			newTransaction.Price = 0
		}
		if strings.HasSuffix(newTransaction.Item, "*") {
			// remove * from the end of the ticker
			newTransaction.Item = newTransaction.Item[0 : len(newTransaction.Item)-1]
		}

		outArr = append(outArr, newTransaction)
	}

	if err == io.EOF {
		err = nil
	}

	return outArr, err
}
