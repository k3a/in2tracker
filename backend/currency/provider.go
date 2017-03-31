package currency

import (
	"fmt"
	"reflect"
	"time"
)

func e(format string, args ...interface{}) error {
	return fmt.Errorf("currency: "+format, args...)
}

// Errors
var (
	ErrNotAvailable = e("rate for the requested currency pair and date is not available")
	ErrBadFormat    = e("provider received wrongly formatted data")
	ErrOldData      = e("provider received old data")
)

// Provider provides a currency convertion
type Provider interface {
	// Name returns the name of the provider
	Name() string
	// Supports checks whether the provider supports the currency conversion.
	// Should return fast (except for the first time the method is called).
	Supports(from Currency, to Currency) bool
	// AllowsReverse specifies whether the provider allows reversing rates
	// (e.g. for EURUSD use USDEUR). It should return fast.
	AllowsReverse() bool
	// GetRate gets the currency rate for the specified time.
	// Returns ErrNotAvailable error if the conversion rate for the
	// specified time is not known.
	GetRate(from Currency, to Currency, at time.Time) (float64, error)
}

// Providers holds all available currency rate providers
var Providers []Provider

// RegisterProvider registers a new currency convertion provider
func RegisterProvider(provider Provider) {
	for _, p := range Providers {
		if p == provider {
			panic("Attempt to register already-registered provider " +
				reflect.TypeOf(provider).String())
		}
	}
	Providers = append(Providers, provider)
}
