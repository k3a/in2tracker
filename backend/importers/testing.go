package importers

import (
	"testing"

	"math"

	"k3a.me/money/backend/currency"
)

/* BASIC TRANSACTION LIST RULES
- no transaction is allowed to be TTInvalid type
- price, fee must always be positive
- currencies must not be invalid if quantity is nonzero
- TTSell must have NetTotal positive or zero, TBuy must have NetTotal negative or zero
- TTDeposit must have NetTotal positive or zero, TTWithdrawal must have NetTotal negative or zero
- TTWithdrawal and TTDeposit must have quantity, price and item(ticker) empty or zero
- TTDividend and TTInterest must have non-empty item (ticker)
- TTSplitMultiplier must have multiplier in the quantity
*/

var epsilon = math.Nextafter(1.0, 2.0) - 1.0

func verifyImporter(trs []*Transaction, t *testing.T) {
	for _, it := range trs {
		if it.Type == TTInvalid {
			t.Fatalf("no transaction is allowed to be TTInvalid type %v", *it)
		}

		if it.Price < 0 {
			t.Fatalf("Price must not be negative %v", *it)
		}
		if it.Fee < 0 {
			t.Fatalf("Fee must not be negative %v", *it)
		}

		if (it.Price != 0 || it.NetTotal != 0) && it.Currency == currency.Invalid {
			t.Fatalf("Currency can't be invalid when Price or NetTotal != 0 %v", *it)
		}
		if it.Fee != 0 && it.FeeCurrency == currency.Invalid {
			t.Fatalf("FeeCurrency can't be invalid %v", *it)
		}

		if it.Type == TTSell && it.NetTotal < 0 {
			t.Fatalf("TTSell must have positive or zero NetTotal %v", *it)
		} else if it.Type == TTBuy && it.NetTotal > 0 {
			t.Fatalf("TTSell must have negative or zero NetTotal %v", *it)
		}

		fee, _ := currency.Convert(it.Fee, it.FeeCurrency, it.Currency, it.Time)
		if it.Type == TTDeposit {
			if !(it.NetTotal+epsilon >= -fee) {
				t.Fatalf("TTDeposit : NetTotal >= -Fee %v", *it)
			}
		} else if it.Type == TTWithdrawal {
			if !(it.NetTotal <= -fee+epsilon) {
				t.Fatalf("TTWithdrawal : NetTotal <= -Fee %v", *it)
			}
		}

		if it.Type == TTWithdrawal || it.Type == TTDeposit {
			if it.Quantity != 0 {
				t.Fatalf("TTWithdrawal and TTDeposit must have zero Quantity %v", *it)
			}
			if it.Price != 0 {
				t.Fatalf("TTWithdrawal and TTDeposit must have zero Price %v", *it)
			}
			if len(it.Item) > 0 {
				t.Fatalf("TTWithdrawal and TTDeposit must not have ticker (Item) specified %v", *it)
			}
		}

		if it.Type == TTDividend || it.Type == TTInterest {
			if len(it.Item) == 0 {
				t.Fatalf("TTDividend and TTInterest must have non-empty Item (ticker) %v", *it)
			}
			if it.Quantity != 0 {
				t.Fatalf("TTDividend and TTInterest must have zero Quantity (use NetTotal) %v", *it)
			}
			if it.Price != 0 {
				t.Fatalf("TTDividend and TTInterest must have zero Price (use NetTotal) %v", *it)
			}
		}

		if it.Type == TTSplitMultiplier {
			if it.Quantity <= 0 {
				t.Fatalf("Quantity for TTSplitMultiplier must be >= 0 %v", *it)
			} else if it.Quantity >= 10 {
				t.Fatalf("Really there was a split >= 10:1? Maybe you wanted to use TTSplitNewShares? %v", *it)
			}
		}

	}
}
