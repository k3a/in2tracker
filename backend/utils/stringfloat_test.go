package utils

import "testing"

func TestStringFloat(t *testing.T) {
	var fs Float64String
	if err := fs.UnmarshalJSON([]byte(`"1000.5"`)); err != nil {
		t.Fatal(err)
	}

	wants := 1000.5
	if fs.Float64 != wants {
		t.Fatalf("wrong value parsed - %f (parsed) != %f", fs.Float64, wants)
	}
}

func TestCZStringFloat(t *testing.T) {
	var fs CZFloat64String
	if err := fs.UnmarshalCSV("1 000, 5"); err != nil {
		t.Fatal(err)
	}

	wants := 1000.5
	if fs.Float64 != wants {
		t.Fatalf("wrong value parsed - %f (parsed) != %f", fs.Float64, wants)
	}
}
