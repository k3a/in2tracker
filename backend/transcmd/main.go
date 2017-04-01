package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alexflint/go-arg"
	"k3a.me/money/backend/currency"
	"k3a.me/money/backend/importers"
	"k3a.me/money/backend/store"
)

func main() {
	var args struct {
		TransactionsOnly bool     `arg:"-t,help:only print transactions"`
		Files            []string `arg:"positional,required,help:CSV files to import"`
	}
	arg.MustParse(&args)

	var trs []*importers.Transaction

	imp := importers.NewCZFioImporter()

	for _, filePath := range args.Files {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot open %s\n", err)
			os.Exit(1)
		}
		defer file.Close()

		curTrs, err := imp.Import(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error importing file %s: %s\n", filePath, err)
			os.Exit(1)
		}

		trs = append(trs, curTrs...)
	}

	// open store
	storePtr := store.New("sqlite3", "database.db")

	// do the job
	proc := NewTransactionProcessor(trs, storePtr, currency.CZK)

	if args.TransactionsOnly {
		if err := proc.PrintTransactions(); err != nil {
			panic(err)
		}
	} else {
		// process
		res, err := proc.Process()
		if err != nil {
			panic(err)
		}

		// for each country..
		for countryName, pc := range res.Countries {
			fmt.Printf("\nCOUNTRY %s\n", countryName)
			// for each company from the country ...
			for _, it := range pc.Items {
				fmt.Printf("  * COMPANY %s - %s - %s\n", it.Item.Code, it.Item.Name, it.Item.Address)
				fmt.Printf("    * Dividend Income: %.2f %s\n", it.DividendIncomeInPrimaryCurrency, proc.PrimaryCurrency)
				fmt.Printf("    * Dividend Tax Paid (local currency): %.2f %s\n", it.DividendTaxPaid, it.Currency)
				fmt.Printf("    * Dividend Tax Paid (in primary): %.2f %s\n", it.DividendTaxPaidInPrimaryCurrency, proc.PrimaryCurrency)
			}
			fmt.Printf("  * Total Dividend Tax Paid in %s: %.2f %s\n",
				countryName, pc.TotalDividendTaxPaidInPrimaryCurrency, proc.PrimaryCurrency)
			fmt.Printf("  * Total Dividend Revenues in %s: %.2f %s\n",
				countryName, pc.TotalDividendIncomeInPrimaryCurrency, proc.PrimaryCurrency)
		}

		// print exp/rev in primary currency
		fmt.Printf("\nTOTAL IN %s (excl. dividends):\n", proc.PrimaryCurrency)
		fmt.Printf("  * Expenses: %.2f %s\n", res.TotalExpensesInPrimaryCurrency, proc.PrimaryCurrency)
		fmt.Printf("  * Revenues: %.2f %s\n", res.TotalRevenuesInPrimaryCurrency, proc.PrimaryCurrency)

		totalDividendIncomePrimary := 0.0
		for _, pc := range res.Countries {
			totalDividendIncomePrimary += pc.TotalDividendIncomeInPrimaryCurrency
		}
		fmt.Printf("\nTOTAL DIVIDEND INCOME IN %s: %.2f\n", proc.PrimaryCurrency, totalDividendIncomePrimary)

		// print net total gain/loss in individual currencies
		currencyCache := NewCurrencyCache(storePtr)
		totalGainLossPrimary := 0.0
		fmt.Printf("\nTOTAL NET GAIN/LOSS IN ORIGINAL CURRENCIES (excl. dividends):\n")
		for currency, total := range res.TotalGainLossByCurrency {
			totalInPrimary, err := currencyCache.Convert(total, currency, proc.PrimaryCurrency, time.Now())
			if err != nil {
				fmt.Printf("  * %.2f %s\n", total, currency)
			} else {
				totalGainLossPrimary += totalInPrimary
				fmt.Printf("  * %.2f %s (= %.2f %s today)\n",
					total, currency, totalInPrimary, proc.PrimaryCurrency)
			}
		}
		fmt.Printf("  => SUM IN %s TODAY: %.2f\n", proc.PrimaryCurrency, totalGainLossPrimary)
	}
}
