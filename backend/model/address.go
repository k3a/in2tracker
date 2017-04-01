package model

import "fmt"

// Address holds complete address
type Address struct {
	Address string
	City    string
	State   string
	Zip     string
	Country string
}

func (a *Address) String() string {
	if len(a.Address) == 0 && len(a.City) == 0 && len(a.State) == 0 &&
		len(a.Zip) == 0 && len(a.Country) == 0 {
		return ""
	}

	return fmt.Sprintf("%s, %s, %s %s, %s",
		a.Address, a.City, a.State, a.Zip, a.Country)
}
