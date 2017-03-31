package currency

import "time"

type currencyPairType string

var pairToProvider = make(map[currencyPairType]Provider)

// currencyPair creates the pair representation
func currencyPair(from Currency, to Currency) currencyPairType {
	return currencyPairType(string(from) + string(to))
}

// ConvertNow converts currencie at rates now using the first available provider
func ConvertNow(amount float64, from Currency, to Currency) (float64, error) {
	return Convert(amount, from, to, time.Now())
}

// Convert converts currency according to rates known for the specified time,
// using the first available provider
func Convert(amount float64, from Currency, to Currency, at time.Time) (float64, error) {
	// check identity
	if from == to {
		return amount, nil
	}

	// try existing provider known to be working first
	provider, has := pairToProvider[currencyPair(from, to)]
	if !has {
		provider, has = pairToProvider[currencyPair(to, from)]
		if has && provider.AllowsReverse() {
			rate, err := provider.GetRate(to, from, at)
			if err == nil {
				return amount / rate, nil
			}
		}
	}

	// try all providers
	var err error
	for _, provider := range Providers {
		var rate, converted float64

		if provider.Supports(from, to) {
			pairToProvider[currencyPair(from, to)] = provider
			rate, err = provider.GetRate(from, to, at)
			converted = amount * rate
		} else if provider.Supports(to, from) && provider.AllowsReverse() {
			pairToProvider[currencyPair(to, from)] = provider
			rate, err = provider.GetRate(to, from, at)
			converted = amount / rate
		}

		if err == nil {
			return converted, nil
		}
	}

	if err != nil {
		return 0, err
	}

	return 0, ErrNotAvailable
}
