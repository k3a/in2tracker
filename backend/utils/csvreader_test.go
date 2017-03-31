package utils

import (
	"encoding/csv"
	"strings"
	"testing"
)

func TestCSVReader(t *testing.T) {
	rd := strings.NewReader(`Some bullsh*t
other bull*hit

almost;good;header;
first;1,2;third;`)

	okrd := NewCSVReader(rd)
	okrd.Comma = ';'

	oneByte := make([]byte, 1)
	n, err := okrd.Read(oneByte)
	if n != 1 {
		t.Fatal("cannot read from csv reader")
	}
	if oneByte[0] != 'a' {
		t.Fatal("csv column header not found properly")
	}
	if err != nil {
		t.Fatal(err)
	}

	r := csv.NewReader(okrd)
	r.Comma = ';'

	rec, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(rec) != 2 {
		t.Fatalf("wrong number of parsed csv lines (%d)\n%#v",
			len(rec), rec)
	}

	if len(rec[0]) != 4 {
		t.Fatalf("wrong number of parsed columns (%d)\n%#v",
			len(rec[0]), rec)
	}

}
