package utils

import (
	"strconv"
	"strings"
)

// Float64String represents a 64bit float encoded as a string in computer format (12345.678)
type Float64String struct {
	// Float64 returns inner value
	Float64 float64
}

func (fs *Float64String) String() string {
	return strconv.FormatFloat(fs.Float64, 'f', 2, 64)
}

// UnmarshalJSON unmarshals the value fron the JSON
func (fs *Float64String) UnmarshalJSON(inp []byte) (err error) {
	if len(inp) <= 2 {
		fs.Float64 = 0
		return
	}

	return fs.UnmarshalCSV(string(inp[1 : len(inp)-1]))
}

// UnmarshalCSV unmarshals the value fron the CSV
func (fs *Float64String) UnmarshalCSV(str string) (err error) {
	fs.Float64, err = strconv.ParseFloat(str, 64)
	return err
}

// CZFloat64String represents a 64bit float encoded as a string in Czech format (1 234, 56)
type CZFloat64String struct {
	// Float64 returns inner value
	Float64 float64
}

func (fs *CZFloat64String) String() string {
	return strconv.FormatFloat(fs.Float64, 'f', 2, 64)
}

var czFloatReplacer = strings.NewReplacer(" ", "", ",", ".")

// ParseCZFloat parses string as czech float format
func ParseCZFloat(str string) (float64, error) {
	if len(str) == 0 {
		return 0, nil
	}

	normalizedFloatStr := czFloatReplacer.Replace(str)
	return strconv.ParseFloat(normalizedFloatStr, 64)
}

// UnmarshalJSON unmarshals the value fron the JSON
func (fs *CZFloat64String) UnmarshalJSON(inp []byte) (err error) {
	if len(inp) <= 2 {
		fs.Float64 = 0
		return nil
	}

	return fs.UnmarshalCSV(string(inp[1 : len(inp)-1]))
}

// UnmarshalCSV unmarshals the value fron the CSV
func (fs *CZFloat64String) UnmarshalCSV(csv string) (err error) {
	fs.Float64, err = ParseCZFloat(csv)
	return
}
