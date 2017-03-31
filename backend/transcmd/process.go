package main

import (
	"fmt"
	"math"
	"sort"
	"time"

	"k3a.me/money/backend/currency"
	"k3a.me/money/backend/importers"
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

type TransactionProcessor struct {
	Transactions []*processorTransaction
	Store        *store.Store
}

func NewTransactionProcessor(trs []*importers.Transaction, storePtr *store.Store) *TransactionProcessor {
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

	return &TransactionProcessor{trsToProcess, storePtr}
}

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

// Process processes sells to find gain and loss
func (tp *TransactionProcessor) Process() error {
	// fix missing currencies (that happens mainly because of splits)
	// this works by assuming the currency never changes for a ticker
	for _, t := range tp.Transactions {
		if t.Transaction.Currency == currency.Invalid {
			t.Transaction.Currency = tp.findCurrencyForItem(t.Transaction.Item)
		}
		if t.Transaction.FeeCurrency == currency.Invalid {
			t.Transaction.FeeCurrency = tp.findFeeCurrencyForItem(t.Transaction.Item)
		}
	}

	// set initial numbers
	for _, ptr := range tp.Transactions {
		if ptr.Transaction.Type == importers.TTBuy {
			ptr.RemainingBuys = ptr.Transaction.Quantity
		}
	}

	// totals
	totalGainLoss := make(map[currency.Currency]float64)

	now := time.Now()
	firstDayOfPreviousYear := time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, now.Location())

	// for each sell (from the oldest; thus reverse)
	for it := len(tp.Transactions) - 1; it >= 0; it-- {
		ptr := tp.Transactions[it]

		// skip transactions old too much
		if ptr.Transaction.Time.Before(firstDayOfPreviousYear) {
			continue
		}

		if ptr.Transaction.Type == importers.TTSell {
			sellTr := ptr.Transaction

			// sell gain
			sellGain := sellTr.Quantity * sellTr.Price
			fee, err := currency.Convert(sellTr.Fee, sellTr.FeeCurrency, sellTr.Currency, sellTr.Time)
			if err != nil {
				return err
			}
			sellGain -= fee

			// print
			fmt.Printf("* %s - SOLD %.2f items for %.2f on %s\n", sellTr.Item, sellTr.Quantity, sellGain, sellTr.Time)

			// find relevant buys
			buys, remain := tp.findOldestAvailableBuys(sellTr.Item, sellTr.Quantity, sellTr.Time)
			if remain > 0 {
				fmt.Printf("!!! WARN: Cannot find a purchase of %.2f items of %s sold on %s\n",
					remain, ptr.Transaction.Item, ptr.Transaction.Time)
			}

			// sum buy cost from buys
			buyCost := 0.0
			for _, buy := range buys {
				buyTr := buy.Transaction.Transaction

				// item cost
				buyCost += buy.Amount * buyTr.Price

				// cost fee fraction converted to transaction currency
				fee, err := currency.Convert(buyTr.Fee/buyTr.Quantity*buy.Amount,
					buyTr.FeeCurrency, buyTr.Currency, buyTr.Time)
				if err != nil {
					return err
				}
				buyCost += fee

				// remove used number of items bought
				buy.Transaction.RemainingBuys -= buy.Amount

				fmt.Printf("  bought %.2f items on %s (%s ago) for %.2f %s\n",
					buy.Amount, buyTr.Time, TimeDifference(buyTr.Time, sellTr.Time),
					buyCost, buyTr.Currency)
			}
			ptr.BuyCost = buyCost

			thisGainLoss := sellGain - ptr.BuyCost
			fmt.Printf("  => gainLoss: %.2f %s \n\n", thisGainLoss, sellTr.Currency)

			prevTotalGainLoss, _ := totalGainLoss[sellTr.Currency]
			prevTotalGainLoss += thisGainLoss
			totalGainLoss[sellTr.Currency] = prevTotalGainLoss
		}
	}

	// print totals
	fmt.Printf("\nTOTALS:\n")
	for c, t := range totalGainLoss {
		fmt.Printf("* %.2f %s\n", t, c)
	}

	return nil
}

// PrintTransactions prints all transactions
func (tp *TransactionProcessor) PrintTransactions() error {
	for _, ptr := range tp.Transactions {
		t := ptr.Transaction
		fmt.Printf("%s\n", t)
	}
	return nil
}
