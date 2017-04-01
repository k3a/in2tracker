package main

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"

	"k3a.me/money/backend/companydata"
	"k3a.me/money/backend/currency"
	"k3a.me/money/backend/importers"
	"k3a.me/money/backend/model"
	"k3a.me/money/backend/store"
)

type processorTransaction struct {
	Transaction   *importers.Transaction
	RemainingBuys float64 // buy only: remaining purchased items to be used
	BuyCost       float64 // sell only: amount it cost buy this sell in transaction currency
}

type transactionWithAmount struct {
	Transaction *processorTransaction
	Amount      float64
}

// TransactionProcessor is used to process transactions to prepare financial results.
// PrimaryCurrency is the main currency to which we want to convert some
// types of financtial amounts to (probably taxpayer's national currency).
type TransactionProcessor struct {
	store           *store.Store
	currencyCache   *CurrencyCache
	Transactions    []*processorTransaction
	PrimaryCurrency currency.Currency
}

// NewTransactionProcessor creates a new transaction processor.
// trs - transactions to process (can contain duplicates)
// storePtr - pointer to store to find/store country and currency data
// primaryCurrency -the main currency to which we want to convert some
// types of financtial amounts to (probably taxpayer's national currency).
func NewTransactionProcessor(trs []*importers.Transaction, storePtr *store.Store, primaryCurrency currency.Currency) *TransactionProcessor {
	var trsToProcess []*processorTransaction
	duplicates := make(map[string]bool)

	now := time.Now()
	firstDayThisYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	for _, t := range trs {
		// prevent duplicates
		if _, yes := duplicates[t.Hash()]; yes {
			continue
		}
		duplicates[t.Hash()] = true

		// only for previous year
		if t.Time.Before(firstDayThisYear) {
			trsToProcess = append(trsToProcess, &processorTransaction{Transaction: t})
		}
	}

	// sort from the newest to the oldest
	sort.SliceStable(trsToProcess, func(i, j int) bool {
		return trsToProcess[i].Transaction.Time.After(trsToProcess[j].Transaction.Time)
	})

	return &TransactionProcessor{
		storePtr,
		NewCurrencyCache(storePtr),
		trsToProcess,
		primaryCurrency,
	}
}

// findOldestAvailableBuys finds oldest buy-type transactions containing needAmount
// amount of items. Argument notAfter specifies the latest time at which
// the buy transaction could have happened.
// Returns list of transactions with amount and remaining quantity which couldn't be found.
func (tp *TransactionProcessor) findOldestAvailableBuys(item string, neededAmount float64, notAfter time.Time) (trs []*transactionWithAmount, missingQuantity float64) {
	if len(tp.Transactions) == 0 {
		return nil, neededAmount
	}

	// from the oldest.. (thus revere)
	for it := len(tp.Transactions) - 1; it >= 0; it-- {
		if neededAmount <= 0 {
			break // done
		}
		t := tp.Transactions[it]

		if t.Transaction.Item == item && t.RemainingBuys > 0 {
			if t.Transaction.Time.After(notAfter) {
				continue // happened after
			}

			takenBuys := math.Min(t.RemainingBuys, neededAmount)
			neededAmount -= takenBuys

			trs = append(trs, &transactionWithAmount{
				Transaction: t,
				Amount:      takenBuys,
			})
		}
	}

	return trs, neededAmount
}

// findCurrencyForItem finds item currency from historical transactions
func (tp *TransactionProcessor) findCurrencyForItem(item string) currency.Currency {
	// from the oldest.. (thus revere)
	for it := len(tp.Transactions) - 1; it >= 0; it-- {
		t := tp.Transactions[it]
		if t.Transaction.Currency != currency.Invalid {
			return t.Transaction.Currency
		}
	}
	return currency.Invalid
}

// findFeeCurrencyForItem finds item fee currency from historical transactions
func (tp *TransactionProcessor) findFeeCurrencyForItem(item string) currency.Currency {
	// from the oldest.. (thus revere)
	for it := len(tp.Transactions) - 1; it >= 0; it-- {
		t := tp.Transactions[it]
		if t.Transaction.FeeCurrency != currency.Invalid {
			return t.Transaction.FeeCurrency
		}
	}
	return currency.Invalid
}

// fixMissingCurrencies fixes missing currencies (that happens mainly because
// of splits). This works by assuming the currency never changes for the item.
func (tp *TransactionProcessor) fixMissingCurrencies() {
	for _, t := range tp.Transactions {
		if t.Transaction.Currency == currency.Invalid {
			t.Transaction.Currency = tp.findCurrencyForItem(t.Transaction.Item)
		}
		if t.Transaction.FeeCurrency == currency.Invalid {
			t.Transaction.FeeCurrency = tp.findFeeCurrencyForItem(t.Transaction.Item)
		}
	}
}

// processSell processes the sell-type transaction
func (tp *TransactionProcessor) processSell(processRes *ProcessResult, ptr *processorTransaction) error {
	sellTr := ptr.Transaction

	// sell gain in primary currency
	sellFee, err := tp.currencyCache.Convert(
		sellTr.Fee, sellTr.FeeCurrency, sellTr.Currency, sellTr.Time)
	if err != nil {
		return err
	}
	sellFeePrimary, err := tp.currencyCache.Convert(
		sellTr.Fee, sellTr.FeeCurrency, processRes.PrimaryCurrency, sellTr.Time)
	if err != nil {
		return err
	}
	sellRevenueInPrimary, err := tp.currencyCache.Convert(
		sellTr.NetTotal+sellFee, sellTr.Currency, processRes.PrimaryCurrency, sellTr.Time)
	if err != nil {
		return err
	}

	// update totals by this sell
	processRes.TotalRevenuesInPrimaryCurrency += sellRevenueInPrimary
	processRes.TotalExpensesInPrimaryCurrency += sellFeePrimary

	// print
	fmt.Printf("* %s - SOLD %.2f items and got %.2f net on %s\n",
		sellTr.Item, sellTr.Quantity, sellTr.NetTotal, sellTr.Time)

	// find relevant buys
	buys, remain := tp.findOldestAvailableBuys(sellTr.Item, sellTr.Quantity, sellTr.Time)
	if remain > 0 {
		fmt.Printf("!!! WARN: Cannot find a purchase of %.2f items of %s sold on %s\n",
			remain, ptr.Transaction.Item, ptr.Transaction.Time)
	}

	// sum buy cost from buys
	buyExpenses := 0.0
	for _, buy := range buys {
		buyTr := buy.Transaction.Transaction

		// item cost
		buyExpenses += buy.Amount * buyTr.Price

		// cost fee fraction converted to transaction currency
		fee, err := tp.currencyCache.Convert(buyTr.Fee/buyTr.Quantity*buy.Amount,
			buyTr.FeeCurrency, buyTr.Currency, buyTr.Time)
		if err != nil {
			return err
		}
		buyExpenses += fee

		// remove used number of items bought
		buy.Transaction.RemainingBuys -= buy.Amount

		fmt.Printf("  bought %.2f items on %s (%s ago) for %.2f net\n",
			buy.Amount, buyTr.Time, TimeDifference(buyTr.Time, sellTr.Time), buyExpenses)
	}
	ptr.BuyCost = buyExpenses

	// buy expenses in primary
	buyExpensesInPrimary, err := tp.currencyCache.Convert(
		ptr.BuyCost, sellTr.Currency, processRes.PrimaryCurrency, sellTr.Time)
	if err != nil {
		return err
	}
	processRes.TotalExpensesInPrimaryCurrency += buyExpensesInPrimary

	thisGainLoss := sellTr.NetTotal - ptr.BuyCost
	fmt.Printf("  => gainLoss: %.2f %s \n\n", thisGainLoss, sellTr.Currency)

	// add to total net gain/loss for the currency
	processRes.TotalGainLossByCurrency[sellTr.Currency] += thisGainLoss

	return nil
}

// processDividend processes the dividend-type transaction
// (transactions with positive net total being income and negative being taxes)
func (tp *TransactionProcessor) processDividend(processRes *ProcessResult, ptr *processorTransaction) error {
	tr := ptr.Transaction

	processItem := processRes.GetItem(tr.Item)
	if processItem == nil {
		// market
		//market := (*model.Market)(nil) //TODO: identify market somehow

		// item and country info
		var country *model.Country
		item, err := tp.store.GetItemByCode(tr.Item)
		if err == nil {
			// get country
			country, err = tp.store.GetCountry(item.CountryID)
			if err != nil {
				return err
			}
		} else if err == sql.ErrNoRows {
			// try fetch company data
			companyData, err := companydata.GetCompanyData(tr.Item)
			if err != nil {
				return fmt.Errorf("unable to get company data for %s: %s", tr.Item, err)
			}

			// country
			country, err = tp.store.GetOrCreateCountry(companyData.GetAddress().Country)
			if err != nil {
				return err
			}

			// currency
			currency, err := tp.store.GetOrCreateCurrency(tr.Currency)
			if err != nil {
				return err
			}

			// create item info
			item = &model.Item{
				MarketID:   0, // TODO: market ID
				CountryID:  country.ID,
				CurrencyID: currency.ID,
				Code:       tr.Item,
				Name:       companyData.GetLongName(),
				Address:    companyData.GetAddress().String(),
			}
			if err := tp.store.CreateItem(item); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		processItem = processRes.AddItem(item, country)
	}

	processItem.Currency = tr.Currency

	if tr.NetTotal >= 0 {
		// dividend revenue in primary
		revenueInPrimary, err := tp.currencyCache.Convert(tr.NetTotal, tr.Currency, tp.PrimaryCurrency, tr.Time)
		if err != nil {
			return err
		}

		// item totals
		processItem.DividendIncomeInPrimaryCurrency += revenueInPrimary

		// country totals
		processRes.Countries[processItem.Country.Name].TotalDividendIncomeInPrimaryCurrency += revenueInPrimary
	} else {
		// tax paid
		taxPaid := -tr.NetTotal

		// tax paid in primary
		taxPaidInPrimary, err := tp.currencyCache.Convert(taxPaid, tr.Currency, tp.PrimaryCurrency, tr.Time)
		if err != nil {
			return err
		}

		// item totals
		processItem.DividendTaxPaid += taxPaid
		processItem.DividendTaxPaidInPrimaryCurrency += taxPaidInPrimary

		// country totals
		processRes.Countries[processItem.Country.Name].TotalDividendTaxPaidInPrimaryCurrency += taxPaidInPrimary
	}

	return nil
}

// processCashAndCapital processes capital returns, merger cash (positive) and fees (negative)
func (tp *TransactionProcessor) processCashAndCapital(processRes *ProcessResult, ptr *processorTransaction) error {
	tr := ptr.Transaction

	expenses := 0.0
	revenues := 0.0

	if tr.NetTotal >= 0 {
		revenues += tr.NetTotal
	} else {
		expenses += -tr.NetTotal
	}

	expensesInPrimary, err := tp.currencyCache.Convert(expenses, tr.Currency, tp.PrimaryCurrency, tr.Time)
	if err != nil {
		return err
	}

	revenuesInPrimary, err := tp.currencyCache.Convert(revenues, tr.Currency, tp.PrimaryCurrency, tr.Time)
	if err != nil {
		return err
	}

	fmt.Printf("* %s - Cash/CapitalReturn/Fee - %s\n", tr.Item, tr.Reference)
	fmt.Printf("  revenues: %.2f %s\n", revenues, tr.Currency)
	fmt.Printf("  expenses: %.2f %s\n\n", expenses, tr.Currency)

	// add to totals
	processRes.TotalExpensesInPrimaryCurrency += expensesInPrimary
	processRes.TotalRevenuesInPrimaryCurrency += revenuesInPrimary
	processRes.TotalGainLossByCurrency[tr.Currency] += revenues - expenses

	return nil
}

// Process processes sells to find gain and loss
func (tp *TransactionProcessor) Process() (*ProcessResult, error) {
	tp.fixMissingCurrencies()

	// set initial numbers
	for _, ptr := range tp.Transactions {
		if ptr.Transaction.Type == importers.TTBuy {
			ptr.RemainingBuys = ptr.Transaction.Quantity
		}
	}

	// result obj
	processRes := NewProcessResult(tp.PrimaryCurrency)

	now := time.Now()
	firstDayOfPreviousYear := time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, now.Location())

	// for each sell (from the oldest; thus reverse)
	for it := len(tp.Transactions) - 1; it >= 0; it-- {
		ptr := tp.Transactions[it]

		// skip transactions which are too old
		if ptr.Transaction.Time.Before(firstDayOfPreviousYear) {
			continue
		}

		var err error

		switch ptr.Transaction.Type {
		case importers.TTSell:
			err = tp.processSell(processRes, ptr)
		case importers.TTDividend:
			err = tp.processDividend(processRes, ptr)
		case importers.TTMergerCash, importers.TTFee, importers.TTReturnOfCapital:
			err = tp.processCashAndCapital(processRes, ptr)
		case importers.TTBuy, importers.TTDeposit, importers.TTWithdrawal:
			break // do nothing with these
		default:
			return nil, fmt.Errorf("process: not a known way to handle this transaction: %s",
				ptr.Transaction.String())
		}

		if err != nil {
			return nil, err
		}
	}

	return processRes, nil
}

// PrintTransactions prints all transactions
// (from the most recent, without diplicates)
func (tp *TransactionProcessor) PrintTransactions() error {
	for _, ptr := range tp.Transactions {
		t := ptr.Transaction
		fmt.Printf("%s\n", t.String())
	}
	return nil
}
