package marketdata

import "time"

var pairToProvider = make(map[string]Provider)

func mipair(market string, item string) string {
	return market + ":" + item
}

// GetItemMarketData returns item price on the market at the specific time
func GetItemMarketData(market string, item string, at time.Time) (*MarketData, error) {
	pair := mipair(market, item)

	if provider, has := pairToProvider[pair]; has {
		if md, err := provider.GetMarketData(market, item, at); err == nil {
			return md, err
		}
	}

	for _, provider := range Providers {
		if md, err := provider.GetMarketData(market, item, at); err == nil {
			pairToProvider[pair] = provider
			return md, err
		}
	}

	return nil, ErrNotAvailable
}

// GetItemMarketDataNow returns item price on the market now
func GetItemMarketDataNow(market string, item string) (*MarketData, error) {
	return GetItemMarketData(market, item, time.Now())
}

// GetItemMarketDataForDateRange returns list of item prices between specified tfrom and tto dates
func GetItemMarketDataForDateRange(market string, item string, tfrom time.Time, tto time.Time) ([]*TimedMarketData, error) {
	pair := mipair(market, item)

	if provider, has := pairToProvider[pair]; has && provider.SupportsDateRange() {
		if prices, err := provider.GetMarketDataForDateRange(market, item, tfrom, tto); err == nil {
			return prices, err
		}
	}

	for _, provider := range Providers {
		if !provider.SupportsDateRange() {
			continue
		}
		if prices, err := provider.GetMarketDataForDateRange(market, item, tfrom, tto); err == nil {
			pairToProvider[pair] = provider
			return prices, err
		}
	}

	return nil, ErrNotAvailable
}

// GetItemInfo returns item info
// Parameter market can be empty
func GetItemInfo(market string, item string) (*ItemInfo, error) {
	for _, p := range Providers {
		if !p.Supports(market, item) {
			continue
		}
		if ii, err := p.GetItemInfo(market, item); err == nil {
			return ii, nil
		}
	}

	return nil, ErrNotAvailable
}
