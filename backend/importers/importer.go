package importers

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/k3a/in2tracker/backend/currency"
	"github.com/k3a/in2tracker/backend/utils"
)

// TransactionType holds a type of transaction
type TransactionType string

// transaction types
const (
	// Invalid / unknown transaction
	TTInvalid = TransactionType("TTInvalid")
	// Fee not directly associated with other transaction
	TTFee = TransactionType("TTFee")
	// Item purchase, also used for currency conversions (purchasing XXX for YYY)
	TTBuy = TransactionType("TTBuy")
	// Item sale, also used for currency conversion (selling XXX for YYY)
	TTSell = TransactionType("TTSell")
	// Dividend paid for a stock
	TTDividend = TransactionType("TTDividend")
	// Interest paid for an item
	TTInterest = TransactionType("TTInterest")
	// Cash deposit
	TTDeposit = TransactionType("TTDeposit")
	// Cash withdrawal
	TTWithdrawal = TransactionType("TTWithdrawal")
	// Stock split with defined split ratio in Quantity, for example Quantity 1.5 would mean split 3:2
	TTSplitMultiplier = TransactionType("TTSplitMultiplier")
	// Return of capital, that decreases investment value
	TTReturnOfCapital = TransactionType("TTReturnOfCapital")
	// Cash returned because of stock merger
	TTMergerCash = TransactionType("TTMergerCash")
)

func (tt TransactionType) String() string {
	return string(tt)
}

// Transaction represents an imported transaction
type Transaction struct {
	// time when the transaction happened
	Time time.Time
	// type of the transaction
	Type TransactionType
	// item this transaction belongs to
	Item string
	// for purchases/sales of items - number of items sold/bought - UNSIGNED
	Quantity float64
	// price at which an item was bought/sold - UNSIGNED
	Price float64
	// net total of this transaction, can change cash ballance, + when selling, - when buying
	NetTotal float64
	// currency of the buy/sell transaction (should apply to both Price and NetTotal)
	Currency currency.Currency
	// fee and commission paid to the exchange or someone else - UNSIGNED
	Fee float64
	// fee currency
	FeeCurrency currency.Currency
	// text reference
	Reference string
}

// String returns printable representation for debug
func (t *Transaction) String() string {
	return fmt.Sprintf("{[%17s on %-30s] %7s %.2f @ %.2f = %.2f %s, fee %.2f %s, %s}",
		t.Type, t.Time, t.Item, t.Quantity, t.Price, t.NetTotal, t.Currency,
		t.Fee, t.FeeCurrency, utils.StringPreview(t.Reference, 10))
}

// Hash returns sha1 hash representing transaction uniquely
func (t *Transaction) Hash() string {
	hashInp := fmt.Sprintf("%d%s%f%f", t.Time.Unix(), t.Item, t.Quantity, t.NetTotal)
	hashBytes := sha1.Sum([]byte(hashInp))
	return hex.EncodeToString(hashBytes[:])
}

func e(format string, args ...interface{}) error {
	return fmt.Errorf("importer: "+format, args...)
}

// Errors
var (
	ErrBadFormat = e("unable to parse the provided data")
)

// Importer allows importing transaction data
type Importer interface {
	// Name returns a name of the importer
	Name() string
	// Import parses data from the reader and returns transactions
	Import(reader io.Reader) ([]*Transaction, error)
}
