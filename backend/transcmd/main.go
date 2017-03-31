package main

import (
	"fmt"
	"os"

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

		// print exp/rev in primary currency
		fmt.Printf("\nTOTAL IN %s:\n", proc.PrimaryCurrency)
		fmt.Printf("* Expenses: %.2f %s\n", res.TotalExpensesInPrimaryCurrency, proc.PrimaryCurrency)
		fmt.Printf("* Revenues: %.2f %s\n", res.TotalRevenuesInPrimaryCurrency, proc.PrimaryCurrency)

		// print totals
		fmt.Printf("\nTOTAL GAIN/LOSS IN ORIGINAL CURRENCIES:\n")
		for c, t := range res.TotalGainLossByCurrency {
			fmt.Printf("* %.2f %s\n", t, c)
		}
	}
}
