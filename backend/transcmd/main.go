package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	_ "github.com/mattn/go-sqlite3" //sqlite driver
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
	storePtr := store.New("sqlite", "database.db")

	// do the job
	proc := NewTransactionProcessor(trs, storePtr)

	if args.TransactionsOnly {
		if err := proc.PrintTransactions(); err != nil {
			panic(err)
		}
	} else {
		if err := proc.Process(); err != nil {
			panic(err)
		}
	}
}
