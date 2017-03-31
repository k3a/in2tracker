package currency

import "testing"
import "time"

func TestCZCNB(t *testing.T) {
	cnb := NewCZCNB()
	if !cnb.Supports(USD, CZK) {
		t.Fatal("cnb.cz must support USDCZK rate")
	}

	rate, err := cnb.GetRate(USD, CZK, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if rate < 2 {
		t.Fatal("wrong rate")
	}
}
