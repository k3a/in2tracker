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
	return fmt.Sprintf("%s, %s, %s %s, %s",
		a.Address, a.City, a.State, a.Zip, a.Country)
}
