package importers

import (
	"os"
	"testing"
)

func TestCZFio(t *testing.T) {
	file, err := os.Open("importer.cz.fio.ebroker_test.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	imp := NewCZFioImporter()

	trs, err := imp.Import(file)
	if err != nil {
		t.Fatal(err)
	}

	wantNum := 16
	if len(trs) != wantNum {
		t.Fatalf("wrong number of parsed transactions (%d parsed != %d)", len(trs), wantNum)
	}

	first := trs[0].Time
	if first.Day() != 12 || first.Month() != 1 || first.Year() != 2017 {
		t.Fatal("bad date parsed")
	}

	if first.Hour() != 15 || first.Minute() != 56 || first.Second() != 0 {
		t.Fatal("bad time parsed")
	}

	// ensure some basic rules
	verifyImporter(trs, t)

	for _, it := range trs {
		t.Logf("%v", *it)
	}
}
